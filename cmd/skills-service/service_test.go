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

package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/retail-cortex/skills/pkg/data"
	"github.com/retail-cortex/skills/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() (*gin.Engine, *Config) {
	gin.SetMode(gin.TestMode)
	data.ResetEngine()

	cfg := &Config{
		Host:        "localhost",
		Port:        8000,
		DatabaseURL: filepath.Join(os.TempDir(), fmt.Sprintf("skills_srv_test_%d.db", time.Now().UnixNano())),
	}

	router := SetupAppEngine(cfg)
	return router, cfg
}

func TestLoadConfig(t *testing.T) {
	t.Setenv("HOST", "127.0.0.1")
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "test.db")
	t.Setenv("ENABLE_OPENTELEMETRY", "true")
	t.Setenv("OTEL_SERVICE_NAME", "custom-service")
	t.Setenv("GCP_PROJECT_ID", "my-gcp-project")

	cfg := LoadConfig()
	assert.Equal(t, "127.0.0.1", cfg.Host)
	assert.Equal(t, 9090, cfg.Port)
	assert.Equal(t, "test.db", cfg.DatabaseURL)
	assert.True(t, cfg.EnableOpenTelemetry)
	assert.Equal(t, "custom-service", cfg.OTELServiceName)
	assert.Equal(t, "my-gcp-project", cfg.GCPProjectID)
}

func TestHealthEndpoint(t *testing.T) {
	router, _ := setupTestRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var body map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &body)
	require.NoError(t, err)
	assert.Equal(t, "ok", body["status"])
	assert.Equal(t, "skills-service", body["service"])
}

func TestInvalidAPIKeySkillOps(t *testing.T) {
	router, _ := setupTestRouter()

	endpoints := []struct {
		method string
		url    string
	}{
		{"POST", "/api/v1/skills"},
		{"PUT", "/api/v1/skills/123"},
		{"PATCH", "/api/v1/skills/123"},
		{"DELETE", "/api/v1/skills/123"},
	}

	for _, ep := range endpoints {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(ep.method, ep.url, bytes.NewBuffer([]byte("{}")))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", "skm_live_invalidkey")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	}
}

func TestStartServer(t *testing.T) {
	cfg := &Config{
		Host:        "127.0.0.1",
		Port:        0,
		DatabaseURL: filepath.Join(t.TempDir(), "test_main.db"),
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := StartServer(ctx, cfg)
	assert.NoError(t, err)
}

func TestAppsAndSkillsRESTWorkflow(t *testing.T) {
	router, _ := setupTestRouter()

	// 1. Register App with X-Forwarded-Proto header
	regPayload := model.AppRegisterRequest{
		AppName: "rest-app",
		Email:   "rest@example.com",
	}
	bodyBytes, _ := json.Marshal(regPayload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/apps/register", bytes.NewBuffer(bodyBytes))
	req.Host = "localhost:8000"
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-Proto", "https")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var appResp model.AppRegisterResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &appResp))
	assert.NotEmpty(t, appResp.APIKey)
	assert.Contains(t, appResp.VerificationURL, "https://")

	// Register with TLS request connection
	regTLS := model.AppRegisterRequest{
		AppName: "tls-app",
		Email:   "tls@example.com",
	}
	tlsBytes, _ := json.Marshal(regTLS)
	w = httptest.NewRecorder()
	reqTLS, _ := http.NewRequest("POST", "/api/v1/apps/register", bytes.NewBuffer(tlsBytes))
	reqTLS.Host = "localhost:8000"
	reqTLS.TLS = &tls.ConnectionState{}
	router.ServeHTTP(w, reqTLS)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Register second app
	regPayload2 := model.AppRegisterRequest{
		AppName: "rest-app-2",
		Email:   "rest2@example.com",
	}
	bBytes2, _ := json.Marshal(regPayload2)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/apps/register", bytes.NewBuffer(bBytes2))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	var appResp2 model.AppRegisterResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &appResp2))
	wVerify2 := httptest.NewRecorder()
	reqVerify2, _ := http.NewRequest("GET", "/api/v1/apps/verify?token="+appResp2.VerificationToken, nil)
	router.ServeHTTP(wVerify2, reqVerify2)
	assert.Equal(t, http.StatusOK, wVerify2.Code)

	// Register duplicate app -> 400 Bad Request
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/apps/register", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Register app with invalid payload -> 400 Bad Request
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/apps/register", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 2. Register skill unverified -> 403 Forbidden
	skillPayload := model.SkillCreateRequest{
		Name:         "rest-skill",
		Description:  "REST test skill",
		Instructions: "Follow instructions",
	}
	sBytes, _ := json.Marshal(skillPayload)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/skills", bytes.NewBuffer(sBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", appResp.APIKey)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)

	// Unverified app on PUT -> 403 Forbidden
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/v1/skills/some-id", bytes.NewBuffer(sBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", appResp.APIKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// Unverified app on PATCH -> 403 Forbidden
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PATCH", "/api/v1/skills/some-id", bytes.NewBuffer(sBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", appResp.APIKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// Unverified app on DELETE -> 403 Forbidden
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/v1/skills/some-id", nil)
	req.Header.Set("X-API-Key", appResp.APIKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// Missing API Key -> 401 Unauthorized
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/skills", bytes.NewBuffer(sBytes))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	// Verify App missing token -> 400 Bad Request
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/apps/verify", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Verify App invalid token -> 404 Not Found
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/apps/verify?token=invalid", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// 3. Verify App 1
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/apps/verify?token="+appResp.VerificationToken, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Register skill with invalid body -> 400 Bad Request
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/skills", bytes.NewBuffer([]byte("{bad}")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", appResp.APIKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 4. Register skill verified -> 201 Created
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/v1/skills", bytes.NewBuffer(sBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", appResp.APIKey)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var skillResp model.SkillResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &skillResp))
	assert.Equal(t, "rest-skill", skillResp.Name)

	// 5. List Skills
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/skills?s=rest&page=1&page_size=5", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "1", w.Header().Get("X-Total-Count"))
	assert.Equal(t, "1", w.Header().Get("X-Page"))
	assert.Equal(t, "5", w.Header().Get("X-Page-Size"))
	assert.Equal(t, "1", w.Header().Get("X-Total-Pages"))
	var listResp []model.SkillResponse
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &listResp))
	assert.Len(t, listResp, 1)

	// List Skills with envelope format
	wEnv := httptest.NewRecorder()
	reqEnv, _ := http.NewRequest("GET", "/api/v1/skills?s=rest&envelope=true", nil)
	router.ServeHTTP(wEnv, reqEnv)
	assert.Equal(t, http.StatusOK, wEnv.Code)
	var envResp model.PaginatedSkillResponse
	require.NoError(t, json.Unmarshal(wEnv.Body.Bytes(), &envResp))
	assert.Equal(t, int64(1), envResp.TotalCount)
	assert.Len(t, envResp.Items, 1)

	// 6. Get Skill by ID
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/skills/"+skillResp.ID, nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Get non-existent skill -> 404 Not Found
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/v1/skills/non-existent", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// 7. PUT Replace Skill (unauthorized app 2 -> 403 Forbidden)
	newDesc := "Replaced description"
	updBytes, _ := json.Marshal(model.SkillUpdateRequest{Description: &newDesc})

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/v1/skills/"+skillResp.ID, bytes.NewBuffer(updBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", appResp2.APIKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// PUT Replace Skill (authorized app 1 -> 200 OK)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/v1/skills/"+skillResp.ID, bytes.NewBuffer(updBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", appResp.APIKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Replace non-existent skill -> 404 Not Found
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/v1/skills/non-existent", bytes.NewBuffer(updBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", appResp.APIKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Replace skill invalid payload -> 400 Bad Request
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PUT", "/api/v1/skills/"+skillResp.ID, bytes.NewBuffer([]byte("{invalid}")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", appResp.APIKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 8. PATCH Update Skill (unauthorized app 2 -> 403 Forbidden)
	patchDesc := "Patched description"
	patchBytes, _ := json.Marshal(model.SkillUpdateRequest{Description: &patchDesc})

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PATCH", "/api/v1/skills/"+skillResp.ID, bytes.NewBuffer(patchBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", appResp2.APIKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// PATCH Update Skill (authorized app 1 -> 200 OK)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PATCH", "/api/v1/skills/"+skillResp.ID, bytes.NewBuffer(patchBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", appResp.APIKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Patch non-existent skill -> 404 Not Found
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PATCH", "/api/v1/skills/non-existent", bytes.NewBuffer(patchBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", appResp.APIKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// Patch skill invalid payload -> 400 Bad Request
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("PATCH", "/api/v1/skills/"+skillResp.ID, bytes.NewBuffer([]byte("{invalid}")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", appResp.APIKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// 9. Delete Skill (unauthorized app 2 -> 403 Forbidden)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/v1/skills/"+skillResp.ID, nil)
	req.Header.Set("X-API-Key", appResp2.APIKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// Delete non-existent skill -> 404 Not Found
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/v1/skills/non-existent", nil)
	req.Header.Set("X-API-Key", appResp.APIKey)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)

	// 10. Delete Skill
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/v1/skills/"+skillResp.ID, nil)
	req.Header.Set("X-API-Key", appResp.APIKey)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
