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
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mark3labs/mcp-go/server"
	"github.com/retail-cortex/skills/pkg/data"
	"github.com/retail-cortex/skills/pkg/embedding"
	"github.com/retail-cortex/skills/pkg/embedding/alloydb"
	"github.com/retail-cortex/skills/pkg/embedding/vertex"
	"github.com/retail-cortex/skills/pkg/mcp"
	"github.com/retail-cortex/skills/pkg/service"
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
	skillsSvc := service.NewSkillsServiceWithProvider(embProvider)
	handlers := NewServerHandlers(appsSvc, skillsSvc)

	mcpServer := mcp.NewMCPServer(appsSvc, skillsSvc)
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

func main() {
	cfg := LoadConfig()
	router := SetupAppEngine(cfg)

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		log.Printf("Starting Gin REST & MCP server on %s...", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server listener error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully.")
}
