//go:build ignore

package main


import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/rrmcguinness/modenv/pkg/modenv"
)

type Config struct {
	Server struct {
		RestPort string `toml:"rest_port"`
	} `toml:"server"`
}

func main() {
	log.Println("Starting Enterprise Go Service...")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	os.Setenv("MODENV_PREFIX", "configs")
	var cfg Config
	clone, err := modenv.Load(&cfg)
	if err != nil {
		log.Fatalf("Config load error: %v", err)
	}
	appConfig := clone.(*Config)

	r := gin.Default()
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	go func() {
		if err := r.Run(":" + appConfig.Server.RestPort); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down Go service cleanly.")
}
