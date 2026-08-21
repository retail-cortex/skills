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
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/server"
	"github.com/retail-cortex/castor/pkg/data"
	"github.com/retail-cortex/castor/pkg/embedding"
	"github.com/retail-cortex/castor/pkg/embedding/alloydb"
	"github.com/retail-cortex/castor/pkg/embedding/vertex"
	"github.com/retail-cortex/castor/pkg/mcp"
	"github.com/retail-cortex/castor/pkg/service"
)

func SetupAppEngine(cfg *Config) *gin.Engine {
	log.Printf("Initializing database tables on target: %s", cfg.DatabaseURL)
	db, err := data.InitDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	log.Println("Setting up telemetry...")
	service.SetupTelemetry(service.TelemetryConfig{
		EnableTelemetry: cfg.EnableOpenTelemetry,
		ServiceName:     cfg.OTELServiceName,
		GCPProjectID:    cfg.GCPProjectID,
	})

	// Instantiate configured embedding provider
	var embProvider embedding.Provider
	switch strings.ToLower(strings.TrimSpace(cfg.EmbeddingProvider)) {
	case "alloydb", "alloydb-ai":
		log.Printf("Initializing AlloyDB AI in-database embedding provider (model: text-embedding-004, dimension: 768)...")
		embProvider = alloydb.NewProvider(alloydb.Config{
			DB:        db,
			ModelName: "text-embedding-004",
			Dimension: 768,
		})
	default:
		log.Printf("Initializing Vertex AI multimodal embedding provider (model: multimodalembedding, dimension: 1408)...")
		embProvider = vertex.NewProvider(vertex.Config{
			ProjectID: cfg.GCPProjectID,
			ModelName: "multimodalembedding",
		})
	}

	appsSvc := service.NewAppsService()
	castorSvc := service.NewCastorServiceWithProvider(embProvider)
	handlers := NewServerHandlers(appsSvc, castorSvc)

	mcpServer := mcp.NewMCPServer(appsSvc, castorSvc)
	sseServer := server.NewSSEServer(mcpServer.Server())

	router := gin.Default()

	// Health endpoint
	router.GET("/health", handlers.HealthCheck(cfg))

	// API v1 group
	v1 := router.Group("/api/v1")
	{
		apps := v1.Group("/apps")
		{
			apps.POST("/register", handlers.RegisterApp)
			apps.GET("/verify", handlers.VerifyApp)

			// RBAC Member Management
			apps.GET("/members", handlers.ListMembers)
			apps.POST("/members/invite", handlers.InviteMember)
			apps.GET("/members/accept", handlers.AcceptInvitation)
			apps.POST("/members/accept", handlers.AcceptInvitation)
			apps.PATCH("/members/:member_id", handlers.UpdateMemberRole)
			apps.DELETE("/members/:member_id", handlers.RemoveMember)

			// Scoped API Key Management
			apps.GET("/keys", handlers.ListAPIKeys)
			apps.POST("/keys", handlers.CreateScopedAPIKey)
			apps.DELETE("/keys/:key_id", handlers.RevokeAPIKey)
		}

		skills := v1.Group("/skills")
		{
			skills.GET("", handlers.ListSkills)
			skills.GET("/*skill_id_or_name", handlers.GetSkill)
			skills.POST("", handlers.RegisterSkill)
			skills.PUT("/*skill_id", handlers.ReplaceSkill)
			skills.PATCH("/*skill_id", handlers.UpdateSkill)
			skills.DELETE("/*skill_id", handlers.DeleteSkill)
		}
	}

	// MCP SSE endpoints
	router.GET("/mcp/sse", gin.WrapH(sseServer.SSEHandler()))
	router.POST("/mcp/messages", gin.WrapH(sseServer.MessageHandler()))

	return router
}

// StartServer starts the HTTP REST & MCP server and blocks until the context is cancelled.
func StartServer(ctx context.Context, cfg *Config) error {
	router := SetupAppEngine(cfg)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	errChan := make(chan error, 1)
	go func() {
		log.Printf("Starting Gin REST & MCP server on %s...", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		log.Println("Shutting down server gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

func main() {
	cfg := LoadConfig()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := StartServer(ctx, cfg); err != nil {
		log.Fatalf("Server error: %v", err)
	}
	log.Println("Server exited gracefully.")
}
