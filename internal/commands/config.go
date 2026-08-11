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

package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type CLIConfig struct {
	ServerURL      string `json:"server_url"`
	APIKey         string `json:"api_key"`
	Domain         string `json:"domain"`
	OrganizationID string `json:"organization_id"`
}

func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	castorPath := filepath.Join(home, ".castor", ".env.toml")
	if _, err := os.Stat(castorPath); err == nil {
		return castorPath
	}
	cstrPath := filepath.Join(home, ".cstr", ".env.toml")
	if _, err := os.Stat(cstrPath); err == nil {
		return cstrPath
	}
	skmPath := filepath.Join(home, ".skm", ".env.toml")
	if _, err := os.Stat(skmPath); err == nil {
		return skmPath
	}
	return castorPath
}

func LoadCLIConfig() (*CLIConfig, error) {
	cfg := &CLIConfig{
		ServerURL: "http://localhost:8080",
	}

	configPath := GetConfigPath()
	if content, err := os.ReadFile(configPath); err == nil {
		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
				switch key {
				case "CASTOR_SERVER_URL", "CSTR_SERVER_URL", "SKM_SERVER_URL", "server_url", "SERVER_URL":
					cfg.ServerURL = val
				case "CASTOR_API_KEY", "CSTR_API_KEY", "SKM_API_KEY", "api_key", "API_KEY":
					cfg.APIKey = val
				case "CASTOR_DOMAIN", "CSTR_DOMAIN", "SKM_DOMAIN", "domain", "DOMAIN":
					cfg.Domain = val
				case "CASTOR_ORGANIZATION_ID", "CSTR_ORGANIZATION_ID", "SKM_ORGANIZATION_ID", "organization_id", "ORGANIZATION_ID", "org":
					cfg.OrganizationID = val
				}
			}
		}
	}

	if envServer := os.Getenv("CASTOR_SERVER_URL"); envServer != "" {
		cfg.ServerURL = envServer
	} else if envServer := os.Getenv("CSTR_SERVER_URL"); envServer != "" {
		cfg.ServerURL = envServer
	} else if envServer := os.Getenv("SKM_SERVER_URL"); envServer != "" {
		cfg.ServerURL = envServer
	}

	if envKey := os.Getenv("CASTOR_API_KEY"); envKey != "" {
		cfg.APIKey = envKey
	} else if envKey := os.Getenv("CSTR_API_KEY"); envKey != "" {
		cfg.APIKey = envKey
	} else if envKey := os.Getenv("SKM_API_KEY"); envKey != "" {
		cfg.APIKey = envKey
	}

	if envDomain := os.Getenv("CASTOR_DOMAIN"); envDomain != "" {
		cfg.Domain = envDomain
	} else if envDomain := os.Getenv("CSTR_DOMAIN"); envDomain != "" {
		cfg.Domain = envDomain
	} else if envDomain := os.Getenv("SKM_DOMAIN"); envDomain != "" {
		cfg.Domain = envDomain
	}

	if envOrg := os.Getenv("CASTOR_ORGANIZATION_ID"); envOrg != "" {
		cfg.OrganizationID = envOrg
	} else if envOrg := os.Getenv("CSTR_ORGANIZATION_ID"); envOrg != "" {
		cfg.OrganizationID = envOrg
	} else if envOrg := os.Getenv("SKM_ORGANIZATION_ID"); envOrg != "" {
		cfg.OrganizationID = envOrg
	}

	return cfg, nil
}

func SaveCLIConfig(cfg *CLIConfig) error {
	configPath := GetConfigPath()
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	content := fmt.Sprintf("# Castor Enterprise CLI Configuration\nCASTOR_SERVER_URL=%q\nCASTOR_API_KEY=%q\nCASTOR_DOMAIN=%q\nCASTOR_ORGANIZATION_ID=%q\n",
		cfg.ServerURL, cfg.APIKey, cfg.Domain, cfg.OrganizationID)
	return os.WriteFile(configPath, []byte(content), 0600)
}

func runConfig(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "show" || args[0] == "list" {
		cfg, err := LoadCLIConfig()
		if err != nil {
			fmt.Fprintf(stderr, "Failed to load CLI config: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "\nCastor CLI Configuration (%s)\n", GetConfigPath())
		fmt.Fprintf(stdout, "%s\n", strings.Repeat("=", 60))
		fmt.Fprintf(stdout, "Server URL:      %s\n", cfg.ServerURL)
		maskedKey := "<not set>"
		if cfg.APIKey != "" {
			if len(cfg.APIKey) > 8 {
				maskedKey = cfg.APIKey[:4] + "..." + cfg.APIKey[len(cfg.APIKey)-4:]
			} else {
				maskedKey = "***"
			}
		}
		fmt.Fprintf(stdout, "API Key:         %s\n", maskedKey)
		domainStr := "<not set>"
		if cfg.Domain != "" {
			domainStr = cfg.Domain
		}
		fmt.Fprintf(stdout, "Domain:          %s\n", domainStr)
		orgStr := "<not set>"
		if cfg.OrganizationID != "" {
			orgStr = cfg.OrganizationID
		}
		fmt.Fprintf(stdout, "Organization ID: %s\n\n", orgStr)
		return 0
	}

	if args[0] == "set" {
		if len(args) < 3 {
			fmt.Fprintf(stderr, "Usage: cstr config set <server|api_key|domain|org> <value>\n")
			return 1
		}
		key := strings.ToLower(args[1])
		val := args[2]

		cfg, _ := LoadCLIConfig()
		switch key {
		case "server", "server_url", "server-url":
			cfg.ServerURL = val
			fmt.Fprintf(stdout, "Updated server URL: %s\n", val)
		case "api_key", "api-key", "key":
			cfg.APIKey = val
			fmt.Fprintf(stdout, "Updated API key.\n")
		case "domain", "castor_domain", "cstr_domain", "skm_domain":
			cfg.Domain = val
			fmt.Fprintf(stdout, "Updated Domain: %s\n", val)
		case "org", "organization_id", "organization-id", "organization":
			cfg.OrganizationID = val
			fmt.Fprintf(stdout, "Updated Organization ID: %s\n", val)
		default:
			fmt.Fprintf(stderr, "Unknown config key %s. Valid keys: server, api_key, domain, org\n", key)
			return 1
		}

		if err := SaveCLIConfig(cfg); err != nil {
			fmt.Fprintf(stderr, "Failed to save configuration: %v\n", err)
			return 1
		}
		return 0
	}

	fmt.Fprintf(stderr, "Unknown config command: %s. Use 'cstr config show' or 'cstr config set <key> <val>'\n", args[0])
	return 1
}
