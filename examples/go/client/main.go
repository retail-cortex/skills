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
	if cfg.SKM.ServerURL == "" {
		cfg.SKM.ServerURL = "http://localhost:8080"
	}
	return &cfg, nil
}

func Run(ctx context.Context) error {
	cfg, err := LoadConfiguration()
	if err != nil {
		log.Printf("Notice: modenv configuration notice: %v", err)
		cfg = &AppConfig{}
		cfg.SKM.ServerURL = "http://localhost:8080"
	}

	fmt.Printf("Loaded SKM Server URL from modenv: %s\n", cfg.SKM.ServerURL)

	// Demonstrate polyglot URI parsing
	scheme, target, ref, subpath := skillsloader.ParseSkillRootURI("github://google/skills@main/tree/main/skills/cloud/gemini-api")
	fmt.Printf("Parsed URI: scheme=%s target=%s ref=%s subpath=%s\n", scheme, target, ref, subpath)

	return nil
}

func main() {
	if err := Run(context.Background()); err != nil {
		log.Fatalf("Go Client Example failed: %v", err)
	}
}
