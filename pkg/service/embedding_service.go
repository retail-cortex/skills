package service

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/retail-cortex/skills/pkg/data"
)

type EmbeddingService struct {
	ModelName    string
	GeminiAPIKey string
	BaseURL      string
}

func NewEmbeddingService() *EmbeddingService {
	modelName := os.Getenv("EMBEDDING_MODEL")
	if modelName == "" {
		modelName = "text-embedding-004"
	}
	apiKey := os.Getenv("GEMINI_API_KEY")
	baseURL := os.Getenv("GEMINI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}
	return &EmbeddingService{
		ModelName:    modelName,
		GeminiAPIKey: apiKey,
		BaseURL:      baseURL,
	}
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

// GenerateEmbedding calls Gemini API if API Key is configured or returns nil.
func (s *EmbeddingService) GenerateEmbedding(text string) []float64 {
	text = strings.TrimSpace(text)
	if text == "" || s.GeminiAPIKey == "" {
		return nil
	}

	base := s.BaseURL
	if base == "" {
		base = "https://generativelanguage.googleapis.com"
	}
	url := base + "/v1beta/models/" + s.ModelName + ":embedContent?key=" + s.GeminiAPIKey
	payload := geminiEmbedRequest{
		Model: "models/" + s.ModelName,
		Content: geminiEmbedContent{
			Parts: []geminiPart{{Text: text}},
		},
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(bodyBytes))
	if err != nil {
		log.Printf("Gemini embedding request failed: %v", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		log.Printf("Gemini embedding returned non-200 (%d): %s", resp.StatusCode, string(respBody))
		return nil
	}

	var res geminiEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil
	}

	return res.Embedding.Values
}

func (s *EmbeddingService) CosineSimilarity(vec1, vec2 []float64) float64 {
	return data.CosineSimilarity(vec1, vec2)
}
