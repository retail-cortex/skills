package main

import (
	"log"
	"os"
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
}

func LoadConfig() *Config {
	cfg := &Config{
		Host:                "0.0.0.0",
		Port:                8000,
		DatabaseURL:         "skills.db",
		EnableOpenTelemetry: false,
		OTELServiceName:     "skills-service",
		GCPProjectID:        "",
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

	return cfg
}
