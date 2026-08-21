// Copyright 2026 Ryan McGuinness
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package vertex

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/retail-cortex/castor/pkg/embedding"
	"github.com/retail-cortex/castor/pkg/model"
)

// Config configures Vertex AI and Gemini embedding generation.
type Config struct {
	ModelName    string
	ProjectID    string
	Region       string
	GeminiAPIKey string
	BaseURL      string
}

// Provider implements embedding.Provider for Vertex AI and Gemini Developer API.
type Provider struct {
	ModelName    string
	ProjectID    string
	Region       string
	GeminiAPIKey string
	BaseURL      string

	tokenMu     sync.RWMutex
	cachedToken string
	tokenExpiry time.Time
	httpClient  *http.Client
}

// NewProvider constructs a Vertex AI / Gemini embedding provider.
func NewProvider(cfg ...Config) *Provider {
	modelName := "multimodalembedding"
	projectID := os.Getenv("GCP_PROJECT_ID")
	if projectID == "" {
		projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}
	region := os.Getenv("GCP_REGION")
	if region == "" {
		region = "us-central1"
	}
	apiKey := os.Getenv("GEMINI_API_KEY")
	baseURL := os.Getenv("VERTEX_AI_BASE_URL")

	if len(cfg) > 0 {
		if cfg[0].ModelName != "" {
			modelName = cfg[0].ModelName
		}
		if cfg[0].ProjectID != "" {
			projectID = cfg[0].ProjectID
		}
		if cfg[0].Region != "" {
			region = cfg[0].Region
		}
		if cfg[0].GeminiAPIKey != "" {
			apiKey = cfg[0].GeminiAPIKey
		}
		if cfg[0].BaseURL != "" {
			baseURL = cfg[0].BaseURL
		}
	}

	return &Provider{
		ModelName:    modelName,
		ProjectID:    projectID,
		Region:       region,
		GeminiAPIKey: apiKey,
		BaseURL:      baseURL,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *Provider) Name() string {
	return "vertex-gemini"
}

func (p *Provider) Dimension() int {
	if strings.Contains(p.ModelName, "multimodal") {
		return 1408
	}
	return 768
}

func (p *Provider) CosineSimilarity(a, b []float64) float64 {
	return embedding.CosineSimilarity(a, b)
}

func (p *Provider) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}

	// 1. Direct Gemini Developer API key if configured
	if p.GeminiAPIKey != "" {
		vec, err := p.generateGeminiEmbedding(text)
		if err == nil && len(vec) > 0 {
			return vec, nil
		}
		if p.ProjectID == "" {
			return nil, err
		}
	}

	// 2. Vertex AI Application Default Credentials (ADC)
	token := p.getValidGCPToken()
	if token != "" && p.ProjectID != "" {
		vec, err := p.generateVertexEmbedding(text, token)
		if err != nil {
			return nil, err
		}
		if len(vec) > 0 {
			return vec, nil
		}
	}

	// 3. Fallback deterministic vector simulation when running unconfigured in test sandboxes
	if p.GeminiAPIKey == "" && token == "" {
		return embedding.GenerateDeterministicVector(text, p.Dimension()), nil
	}

	return nil, fmt.Errorf("no valid embedding credentials configured")
}

func (p *Provider) GenerateImageEmbedding(ctx context.Context, base64Image string) ([]float64, error) {
	base64Image = strings.TrimSpace(base64Image)
	if base64Image == "" {
		return nil, nil
	}

	token := p.getValidGCPToken()
	if token != "" && p.ProjectID != "" {
		vec, err := p.generateVertexImageEmbedding(base64Image, token)
		if err != nil {
			return nil, err
		}
		if len(vec) > 0 {
			return vec, nil
		}
	}

	if token == "" {
		return embedding.GenerateDeterministicVector(fmt.Sprintf("Image Payload Hash %d", len(base64Image)), p.Dimension()), nil
	}

	return nil, fmt.Errorf("vertex AI image embedding failed")
}

func (p *Provider) GenerateSkillEmbeddings(
	ctx context.Context,
	name, description, instructions string,
	triggers []string,
	references map[string]string,
	examples map[string]string,
) ([]model.SkillEmbeddingChunk, error) {
	type chunkTask struct {
		targetType string
		targetName string
		isImage    bool
		payload    string
	}

	var tasks []chunkTask

	// 1. Root Skill Summary & Instructions
	trigStr := strings.Join(triggers, " ")
	summaryText := fmt.Sprintf("%s %s %s", name, description, trigStr)
	for i, chunk := range embedding.SplitTextIntoChunks(summaryText, 900) {
		tName := "SKILL.md#summary"
		if i > 0 {
			tName = fmt.Sprintf("SKILL.md#summary-%d", i+1)
		}
		tasks = append(tasks, chunkTask{
			targetType: "skill",
			targetName: tName,
			payload:    chunk,
		})
	}

	if strings.TrimSpace(instructions) != "" {
		for i, chunk := range embedding.SplitTextIntoChunks(instructions, 900) {
			tName := "SKILL.md#instructions"
			if i > 0 {
				tName = fmt.Sprintf("SKILL.md#instructions-%d", i+1)
			}
			tasks = append(tasks, chunkTask{
				targetType: "skill",
				targetName: tName,
				payload:    chunk,
			})
		}
	}

	// 2. References (Multi-Modal)
	for refName, content := range references {
		if strings.HasPrefix(content, "data:image/") && strings.Contains(content, ";base64,") {
			parts := strings.SplitN(content, ";base64,", 2)
			tasks = append(tasks, chunkTask{
				targetType: "reference",
				targetName: refName,
				isImage:    true,
				payload:    parts[1],
			})
			tasks = append(tasks, chunkTask{
				targetType: "reference",
				targetName: refName,
				isImage:    false,
				payload:    fmt.Sprintf("Visual image reference asset: %s", refName),
			})
		} else if strings.HasPrefix(content, "data:") {
			tasks = append(tasks, chunkTask{
				targetType: "reference",
				targetName: refName,
				payload:    fmt.Sprintf("Binary reference resource asset: %s", refName),
			})
		} else {
			textChunks := embedding.SplitTextIntoChunks(content, 900)
			for i, chunk := range textChunks {
				tName := refName
				if len(textChunks) > 1 {
					tName = fmt.Sprintf("%s#part-%d", refName, i+1)
				}
				tasks = append(tasks, chunkTask{
					targetType: "reference",
					targetName: tName,
					payload:    fmt.Sprintf("Reference %s: %s", tName, chunk),
				})
			}
		}
	}

	// 3. Examples & Scripts
	for exName, content := range examples {
		if strings.HasPrefix(content, "data:") {
			tasks = append(tasks, chunkTask{
				targetType: "example",
				targetName: exName,
				payload:    fmt.Sprintf("Binary example resource asset: %s", exName),
			})
		} else {
			textChunks := embedding.SplitTextIntoChunks(content, 900)
			for i, chunk := range textChunks {
				tName := exName
				if len(textChunks) > 1 {
					tName = fmt.Sprintf("%s#part-%d", exName, i+1)
				}
				tasks = append(tasks, chunkTask{
					targetType: "example",
					targetName: tName,
					payload:    fmt.Sprintf("Example script %s: %s", tName, chunk),
				})
			}
		}
	}

	if len(tasks) == 0 {
		return nil, nil
	}

	// Channel-orchestrated Fan-Out / Fan-In
	taskChan := make(chan chunkTask, len(tasks))
	resultChan := make(chan model.SkillEmbeddingChunk, len(tasks))

	for _, t := range tasks {
		taskChan <- t
	}
	close(taskChan)

	numWorkers := 8
	if len(tasks) < numWorkers {
		numWorkers = len(tasks)
	}

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range taskChan {
				var vec []float64
				var err error
				if t.isImage {
					vec, err = p.GenerateImageEmbedding(ctx, t.payload)
				} else {
					vec, err = p.GenerateEmbedding(ctx, t.payload)
				}
				if err == nil && len(vec) > 0 {
					resultChan <- model.SkillEmbeddingChunk{
						TargetType: t.targetType,
						TargetName: t.targetName,
						Vector:     vec,
						ModelName:  p.ModelName,
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var chunks []model.SkillEmbeddingChunk
	for chunk := range resultChan {
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

func (p *Provider) generateVertexEmbedding(text, token string) ([]float64, error) {
	base := p.BaseURL
	if base == "" {
		base = fmt.Sprintf("https://%s-aiplatform.googleapis.com", p.Region)
	}

	endpoint := fmt.Sprintf("%s/v1/projects/%s/locations/%s/publishers/google/models/%s:predict",
		strings.TrimRight(base, "/"), p.ProjectID, p.Region, p.ModelName)

	var reqBody map[string]interface{}
	if strings.Contains(p.ModelName, "multimodal") {
		reqBody = map[string]interface{}{
			"instances": []map[string]interface{}{
				{"text": text},
			},
		}
	} else {
		reqBody = map[string]interface{}{
			"instances": []map[string]interface{}{
				{"content": text},
			},
		}
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vertex API returned HTTP %d", resp.StatusCode)
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	vec := extractVectorFromPrediction(raw, "textEmbedding")
	if len(vec) == 0 {
		return nil, fmt.Errorf("no vector found in prediction response")
	}
	return vec, nil
}

func (p *Provider) generateVertexImageEmbedding(base64Image, token string) ([]float64, error) {
	base := p.BaseURL
	if base == "" {
		base = fmt.Sprintf("https://%s-aiplatform.googleapis.com", p.Region)
	}

	endpoint := fmt.Sprintf("%s/v1/projects/%s/locations/%s/publishers/google/models/%s:predict",
		strings.TrimRight(base, "/"), p.ProjectID, p.Region, p.ModelName)

	reqBody := map[string]interface{}{
		"instances": []map[string]interface{}{
			{
				"image": map[string]interface{}{
					"bytesBase64Encoded": base64Image,
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vertex image API returned HTTP %d", resp.StatusCode)
	}

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	vec := extractVectorFromPrediction(raw, "imageEmbedding")
	if len(vec) == 0 {
		return nil, fmt.Errorf("no vector found in image prediction response")
	}
	return vec, nil
}

func extractVectorFromPrediction(raw map[string]interface{}, preferredKey string) []float64 {
	preds, ok := raw["predictions"].([]interface{})
	if !ok || len(preds) == 0 {
		return nil
	}

	first, ok := preds[0].(map[string]interface{})
	if !ok {
		return nil
	}

	if val, exists := first[preferredKey]; exists {
		if slice, ok := val.([]interface{}); ok {
			out := make([]float64, len(slice))
			for i, v := range slice {
				if f, ok := v.(float64); ok {
					out[i] = f
				}
			}
			return out
		}
	}

	if embMap, ok := first["embeddings"].(map[string]interface{}); ok {
		if val, ok := embMap["values"].([]interface{}); ok {
			out := make([]float64, len(val))
			for i, v := range val {
				if f, ok := v.(float64); ok {
					out[i] = f
				}
			}
			return out
		}
	}

	return nil
}

type geminiEmbedRequest struct {
	Model   string             `json:"model"`
	Content geminiEmbedContent `json:"content"`
}

type geminiEmbedContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiEmbedResponse struct {
	Embedding struct {
		Values []float64 `json:"values"`
	} `json:"embedding"`
}

func (p *Provider) generateGeminiEmbedding(text string) ([]float64, error) {
	base := p.BaseURL
	if base == "" {
		base = "https://generativelanguage.googleapis.com"
	}
	url := fmt.Sprintf("%s/v1beta/models/%s:embedContent?key=%s",
		strings.TrimRight(base, "/"), p.ModelName, p.GeminiAPIKey)

	payload := geminiEmbedRequest{
		Model: "models/" + p.ModelName,
		Content: geminiEmbedContent{
			Parts: []geminiPart{{Text: text}},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	resp, err := p.httpClient.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini API returned HTTP %d", resp.StatusCode)
	}

	var res geminiEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}

	return res.Embedding.Values, nil
}

func (p *Provider) getValidGCPToken() string {
	p.tokenMu.RLock()
	if p.cachedToken != "" && time.Now().UTC().Before(p.tokenExpiry.Add(-60*time.Second)) {
		token := p.cachedToken
		p.tokenMu.RUnlock()
		return token
	}
	p.tokenMu.RUnlock()

	p.tokenMu.Lock()
	defer p.tokenMu.Unlock()

	if p.cachedToken != "" && time.Now().UTC().Before(p.tokenExpiry.Add(-60*time.Second)) {
		return p.cachedToken
	}

	token, expiry, err := resolveADC()
	if err == nil && token != "" {
		p.cachedToken = token
		if !expiry.IsZero() {
			p.tokenExpiry = expiry
		} else {
			p.tokenExpiry = time.Now().UTC().Add(55 * time.Minute)
		}
		return token
	}

	return ""
}

func resolveADC() (string, time.Time, error) {
	if tok := os.Getenv("GOOGLE_OAUTH_ACCESS_TOKEN"); tok != "" {
		return tok, time.Now().UTC().Add(45 * time.Minute), nil
	}

	adcPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if adcPath == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			adcPath = filepath.Join(home, ".config", "gcloud", "application_default_credentials.json")
		}
	}

	if adcPath != "" {
		data, err := os.ReadFile(adcPath)
		if err == nil {
			var creds struct {
				Type         string `json:"type"`
				ClientID     string `json:"client_id"`
				ClientSecret string `json:"client_secret"`
				RefreshToken string `json:"refresh_token"`
				PrivateKey   string `json:"private_key"`
				ClientEmail  string `json:"client_email"`
				TokenURI     string `json:"token_uri"`
			}
			if jsonErr := json.Unmarshal(data, &creds); jsonErr == nil {
				if creds.RefreshToken != "" && creds.ClientID != "" && creds.ClientSecret != "" {
					return refreshOAuthToken(creds.ClientID, creds.ClientSecret, creds.RefreshToken)
				}
				if creds.Type == "service_account" && creds.PrivateKey != "" && creds.ClientEmail != "" {
					return exchangeServiceAccountJWT(creds.ClientEmail, creds.PrivateKey, creds.TokenURI)
				}
			}
		}
	}

	cmd := exec.Command("gcloud", "auth", "print-access-token")
	out, err := cmd.Output()
	if err == nil {
		tok := strings.TrimSpace(string(out))
		if tok != "" {
			return tok, time.Now().UTC().Add(45 * time.Minute), nil
		}
	}

	return "", time.Time{}, fmt.Errorf("unable to resolve GCP credentials")
}

func refreshOAuthToken(clientID, clientSecret, refreshToken string) (string, time.Time, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", time.Time{}, fmt.Errorf("token refresh failed (%d): %s", resp.StatusCode, string(b))
	}

	var res struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", time.Time{}, err
	}

	expiry := time.Now().UTC().Add(time.Duration(res.ExpiresIn) * time.Second)
	return res.AccessToken, expiry, nil
}

func exchangeServiceAccountJWT(clientEmail, privateKeyPEM, tokenURI string) (string, time.Time, error) {
	if tokenURI == "" {
		tokenURI = "https://oauth2.googleapis.com/token"
	}

	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", time.Time{}, fmt.Errorf("failed to decode private key PEM")
	}

	var privKey *rsa.PrivateKey
	parsedKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err == nil {
		var ok bool
		privKey, ok = parsedKey.(*rsa.PrivateKey)
		if !ok {
			return "", time.Time{}, fmt.Errorf("PKCS8 key is not RSA")
		}
	} else {
		privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return "", time.Time{}, fmt.Errorf("failed parsing private key: %w", err)
		}
	}

	now := time.Now().UTC()
	exp := now.Add(1 * time.Hour)

	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"RS256","typ":"JWT"}`))
	claimsJSON := fmt.Sprintf(`{"iss":%q,"scope":"https://www.googleapis.com/auth/cloud-platform","aud":%q,"exp":%d,"iat":%d}`,
		clientEmail, tokenURI, exp.Unix(), now.Unix())
	claims := base64.RawURLEncoding.EncodeToString([]byte(claimsJSON))

	sigInput := header + "." + claims
	hash := sha256.Sum256([]byte(sigInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, privKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", time.Time{}, fmt.Errorf("signing JWT failed: %w", err)
	}

	signedJWT := sigInput + "." + base64.RawURLEncoding.EncodeToString(sig)

	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	data.Set("assertion", signedJWT)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.PostForm(tokenURI, data)
	if err != nil {
		return "", time.Time{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", time.Time{}, fmt.Errorf("JWT exchange failed (%d): %s", resp.StatusCode, string(b))
	}

	var res struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", time.Time{}, err
	}

	return res.AccessToken, time.Now().UTC().Add(time.Duration(res.ExpiresIn) * time.Second), nil
}
