package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/genai"
	"google.golang.org/grpc"
)

type Config struct {
	RestPort string
	GrpcPort string
	Project  string
}

func StartServers(ctx context.Context, cfg *Config, aiClient *genai.Client) error {
	errChan := make(chan error, 2)

	// Start gRPC Server
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%s", cfg.GrpcPort))
	if err != nil {
		return fmt.Errorf("failed to listen on gRPC port %s: %w", cfg.GrpcPort, err)
	}

	gServer := grpc.NewServer()
	go func() {
		log.Printf("Starting gRPC server on :%s...", cfg.GrpcPort)
		if err := gServer.Serve(grpcListener); err != nil && err != grpc.ErrServerStopped {
			errChan <- err
		}
	}()

	// Start Gin REST Server
	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	restServer := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.RestPort),
		Handler: r,
	}

	go func() {
		log.Printf("Starting Gin REST server on :%s...", cfg.RestPort)
		if err := restServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	// Graceful shutdown listener
	go func() {
		<-ctx.Done()
		log.Println("Shutting down servers gracefully...")
		gServer.GracefulStop()
		restServer.Shutdown(context.Background())
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return nil
	}
}
