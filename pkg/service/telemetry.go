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
