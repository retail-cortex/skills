//go:build ignore

package main


import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/enterprise/service/internal/server"
	"google.golang.org/genai"
)

func main() {
	log.Println("Initializing Enterprise Go Service...")

	// 1. Hook OS signals for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 2. Load configuration
	cfg := server.Config{
		RestPort: "8080",
		GrpcPort: "9090",
		Project:  os.Getenv("GOOGLE_CLOUD_PROJECT"),
	}

	// 3. Initialize Vertex AI GenAI client
	var genaiClient *genai.Client
	if cfg.Project != "" {
		var err error
		genaiClient, err = genai.NewClient(ctx, &genai.ClientConfig{
			Project:  cfg.Project,
			Location: "us-central1",
			Backend:  genai.BackendVertexAI,
		})
		if err != nil {
			log.Fatalf("Critical: Failed to initialize Vertex GenAI client: %v", err)
		}
	}

	// 4. Start concurrent Gin REST and gRPC listeners
	if err := server.StartServers(ctx, &cfg, genaiClient); err != nil {
		log.Fatalf("Critical: Server failure: %v", err)
	}

	log.Println("Service cleanly terminated.")
}
