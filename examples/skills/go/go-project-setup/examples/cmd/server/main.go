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
