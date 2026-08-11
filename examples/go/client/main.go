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
	"os"

	"github.com/retail-cortex/skills/clients/go/pkg/skillsloader"
	"github.com/rrmcguinness/modenv/pkg/modenv"
)

type AppConfig struct {
	Castor struct {
		ServerURL string `toml:"server_url"`
		APIKey    string `toml:"api_key"`
	} `toml:"castor"`
	SKM struct {
		ServerURL string `toml:"server_url"`
		APIKey    string `toml:"api_key"`
	} `toml:"skm"`
}

func LoadConfiguration() (*AppConfig, error) {
	os.Setenv("MODENV_PREFIX", "configs")
	var cfg AppConfig
	if _, err := modenv.Load(&cfg); err != nil {
		return nil, fmt.Errorf("modenv load error: %w", err)
	}
	if cfg.Castor.ServerURL == "" {
		if cfg.SKM.ServerURL != "" {
			cfg.Castor.ServerURL = cfg.SKM.ServerURL
			cfg.Castor.APIKey = cfg.SKM.APIKey
		} else {
			cfg.Castor.ServerURL = "http://localhost:8080"
		}
	}
	return &cfg, nil
}

func Run(ctx context.Context) error {
	cfg, err := LoadConfiguration()
	if err != nil {
		log.Printf("Notice: modenv configuration notice: %v", err)
		cfg = &AppConfig{}
		cfg.Castor.ServerURL = "http://localhost:8080"
	}

	fmt.Printf("Loaded Castor Server URL from modenv: %s\n", cfg.Castor.ServerURL)

	// Demonstrate polyglot URI parsing
	scheme, target, ref, subpath := skillsloader.ParseSkillRootURI("castor://skills/example.com/testing/test-skill/1.0.0")
	fmt.Printf("Parsed URI: scheme=%s target=%s ref=%s subpath=%s\n", scheme, target, ref, subpath)

	return nil
}

func main() {
	if err := Run(context.Background()); err != nil {
		log.Fatalf("Go Client Example failed: %v", err)
	}
}
