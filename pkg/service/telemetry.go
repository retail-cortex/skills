package service

import (
	"log"
	"os"
)

type TelemetryConfig struct {
	EnableTelemetry bool
	ServiceName     string
	GCPProjectID    string
}

func SetupTelemetry(cfg TelemetryConfig) {
	sName := cfg.ServiceName
	if sName == "" {
		if envName := os.Getenv("OTEL_SERVICE_NAME"); envName != "" {
			sName = envName
		} else {
			sName = "skills-service"
		}
	}

	if cfg.EnableTelemetry {
		log.Printf("Initializing OpenTelemetry for service '%s' (GCP Project: '%s')...", sName, cfg.GCPProjectID)
	} else {
		log.Printf("OpenTelemetry disabled for service '%s'. Using standard logger.", sName)
	}
}
