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
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/rrmcguinness/modenv/pkg/modenv"
)

type Config struct {
	Host                string `toml:"host"`
	Port                int    `toml:"port"`
	DatabaseURL         string `toml:"database_url"`
	EnableOpenTelemetry bool   `toml:"enable_opentelemetry"`
	OTELServiceName     string `toml:"otel_service_name"`
	GCPProjectID        string `toml:"gcp_project_id"`
	EmbeddingProvider   string `toml:"embedding_provider"`
}

func LoadConfig() *Config {
	cfg := &Config{
		Host:                "0.0.0.0",
		Port:                8000,
		DatabaseURL:         "castor.db",
		EnableOpenTelemetry: false,
		OTELServiceName:     "castor-registry",
		GCPProjectID:        "",
		EmbeddingProvider:   "vertex-gemini",
	}

	// If MODENV_PREFIX is unset, auto-resolve from workspace directory or local directory
	if os.Getenv("MODENV_PREFIX") == "" {
		candidates := []string{".", "cmd/castor-server", "cmd/skills-service"}
		if wsDir := os.Getenv("BUILD_WORKSPACE_DIRECTORY"); wsDir != "" {
			candidates = append([]string{
				filepath.Join(wsDir, "cmd/castor-server"),
				filepath.Join(wsDir, "cmd/skills-service"),
				wsDir,
			}, candidates...)
		}
		for _, c := range candidates {
			if _, err := os.Stat(filepath.Join(c, ".env.toml")); err == nil {
				_ = os.Setenv("MODENV_PREFIX", c)
				break
			}
		}
	}

	// Map ENV or APP_ENV to MODENV_ENV if set
	if env := os.Getenv("ENV"); env != "" && os.Getenv("MODENV_ENV") == "" {
		_ = os.Setenv("MODENV_ENV", env)
	} else if appEnv := os.Getenv("APP_ENV"); appEnv != "" && os.Getenv("MODENV_ENV") == "" {
		_ = os.Setenv("MODENV_ENV", appEnv)
	}

	// Load hierarchical configuration using modenv (.env.toml, .env.<runtime>.toml, .env.local.toml)
	if _, err := modenv.Load(cfg); err != nil {
		log.Printf("modenv load notice: %v (using defaults/env vars)", err)
	}

	// Explicit environment variables override configuration files
	if host := os.Getenv("HOST"); host != "" {
		cfg.Host = host
	}
	if portStr := os.Getenv("PORT"); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil {
			cfg.Port = p
		}
	}
	if dbURL := os.Getenv("DATABASE_URL"); dbURL != "" {
		cfg.DatabaseURL = dbURL
	}
	if enableOTel := os.Getenv("ENABLE_OPENTELEMETRY"); enableOTel != "" {
		cfg.EnableOpenTelemetry = enableOTel == "true"
	}
	if otelService := os.Getenv("OTEL_SERVICE_NAME"); otelService != "" {
		cfg.OTELServiceName = otelService
	}
	if gcpProject := os.Getenv("GCP_PROJECT_ID"); gcpProject != "" {
		cfg.GCPProjectID = gcpProject
	}
	if embProv := os.Getenv("EMBEDDING_PROVIDER"); embProv != "" {
		cfg.EmbeddingProvider = embProv
	}

	return cfg
}
