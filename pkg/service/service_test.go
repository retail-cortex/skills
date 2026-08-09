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

package service_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/retail-cortex/skills/pkg/data"
	"github.com/retail-cortex/skills/pkg/model"
	"github.com/retail-cortex/skills/pkg/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) {
	data.ResetEngine()
	_, err := data.InitDB(filepath.Join(t.TempDir(), "test.db"))
	require.NoError(t, err)
}

func TestAppsService(t *testing.T) {
	setupTestDB(t)
	db := data.GetDB()
	customRepo := data.NewAppsRepository()
	svc := service.NewAppsService(customRepo)

	req := model.AppRegisterRequest{
		AppName: "svc-app",
		Email:   "svc@example.com",
	}
	regRes, err := svc.RegisterApp(db, req, "http://localhost:8000")
	require.NoError(t, err)
	assert.NotEmpty(t, regRes.AppID)

	vRes, err := svc.VerifyApp(db, regRes.VerificationToken)
	require.NoError(t, err)
	assert.True(t, vRes.IsActive)

	app, err := svc.AuthenticateAPIKey(db, regRes.APIKey)
	require.NoError(t, err)
	assert.Equal(t, regRes.AppID, app.AppID)
}

func TestSkillsService(t *testing.T) {
	setupTestDB(t)
	db := data.GetDB()
	appsSvc := service.NewAppsService()
	customSkillsRepo := data.NewSkillsRepository()
	skillsSvc := service.NewSkillsService(customSkillsRepo)

	regRes, err := appsSvc.RegisterApp(db, model.AppRegisterRequest{
		AppName: "svc-skills-app",
		Email:   "svcskills@example.com",
	}, "")
	require.NoError(t, err)
	_, _ = appsSvc.VerifyApp(db, regRes.VerificationToken)

	createReq := model.SkillCreateRequest{
		Name:           "svc-skill",
		Description:    "Service test skill",
		Instructions:   "Follow service rules",
		TriggerPhrases: []string{"trigger one", "trigger two"},
	}
	res, err := skillsSvc.CreateSkill(db, regRes.AppID, createReq)
	require.NoError(t, err)
	assert.Equal(t, "svc-skill", res.Name)

	fetched, err := skillsSvc.GetSkill(db, res.ID)
	require.NoError(t, err)
	assert.Equal(t, res.Name, fetched.Name)

	list, err := skillsSvc.ListSkills(db, "svc-skill")
	require.NoError(t, err)
	assert.Equal(t, int64(1), list.TotalCount)
	assert.Len(t, list.Items, 1)

	// List with empty query
	emptyList, err := skillsSvc.ListSkills(db, "   ")
	require.NoError(t, err)
	assert.Equal(t, int64(1), emptyList.TotalCount)
	assert.Len(t, emptyList.Items, 1)

	// Update Skill with instructions and trigger phrases
	newDesc := "Updated service desc"
	newInst := "Updated service instructions"
	newTrig := []string{"new trigger"}
	updRes, err := skillsSvc.UpdateSkill(db, res.ID, regRes.AppID, model.SkillUpdateRequest{
		Description:    &newDesc,
		Instructions:   &newInst,
		TriggerPhrases: &newTrig,
	}, true)
	require.NoError(t, err)
	assert.Equal(t, "Updated service desc", updRes.Description)
	assert.Equal(t, "Updated service instructions", updRes.Instructions)

	// Update with non-existent skill -> error
	_, err = skillsSvc.UpdateSkill(db, "non-existent", regRes.AppID, model.SkillUpdateRequest{}, false)
	assert.Error(t, err)

	delRes, err := skillsSvc.DeleteSkill(db, res.ID, regRes.AppID)
	require.NoError(t, err)
	assert.Equal(t, "success", delRes["status"])
}

func TestEmbeddingService(t *testing.T) {
	embSvc := service.NewEmbeddingService()
	assert.Equal(t, "multimodalembedding", embSvc.ModelName())

	// Cosine similarity
	sim := embSvc.CosineSimilarity([]float64{1.0, 0.0}, []float64{1.0, 0.0})
	assert.InDelta(t, 1.0, sim, 0.0001)

	// Unconfigured API key returns nil
	vec := embSvc.GenerateEmbedding("")
	assert.Nil(t, vec)

	// Mock HTTP Server success (Gemini developer API)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"embedding": map[string]interface{}{
				"values": []float64{0.1, 0.2, 0.3},
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	mockSvc := service.NewEmbeddingService(service.EmbeddingConfig{
		ModelName:    "text-embedding-004",
		GeminiAPIKey: "mock-key",
		BaseURL:      ts.URL,
	})

	embeddings := mockSvc.GenerateEmbedding("query text")
	require.NotNil(t, embeddings)
	assert.Equal(t, []float64{0.1, 0.2, 0.3}, embeddings)

	// Mock Vertex AI Prediction HTTP Server
	tsVertex := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer mock-token", r.Header.Get("Authorization"))
		resp := map[string]interface{}{
			"predictions": []map[string]interface{}{
				{
					"embeddings": map[string]interface{}{
						"values": []float64{0.7, 0.8, 0.9},
					},
				},
			},
		}
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer tsVertex.Close()

	t.Setenv("GOOGLE_OAUTH_ACCESS_TOKEN", "mock-token")
	vertexSvc := service.NewEmbeddingService(service.EmbeddingConfig{
		ModelName: "text-embedding-004",
		ProjectID: "test-gcp-project",
		Region:    "us-central1",
		BaseURL:   tsVertex.URL,
	})
	vEmbeddings := vertexSvc.GenerateEmbedding("vertex query text")
	require.NotNil(t, vEmbeddings)
	assert.Equal(t, []float64{0.7, 0.8, 0.9}, vEmbeddings)

	// Mock HTTP Server 500 Error
	tsErr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error"))
	}))
	defer tsErr.Close()

	mockErrSvc := service.NewEmbeddingService(service.EmbeddingConfig{
		ModelName:    "text-embedding-004",
		GeminiAPIKey: "mock-key",
		BaseURL:      tsErr.URL,
	})
	assert.Nil(t, mockErrSvc.GenerateEmbedding("query text"))
}

func TestTelemetry(t *testing.T) {
	service.SetupTelemetry(service.TelemetryConfig{
		EnableTelemetry: true,
		ServiceName:     "test-telemetry",
		GCPProjectID:    "test-project",
	})
	service.SetupTelemetry(service.TelemetryConfig{
		EnableTelemetry: false,
	})
}
