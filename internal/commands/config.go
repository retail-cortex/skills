package commands

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type CLIConfig struct {
	ServerURL string `json:"server_url"`
	APIKey    string `json:"api_key"`
}

func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".skm", ".env.toml")
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
				case "SKM_SERVER_URL", "server_url", "SERVER_URL":
					cfg.ServerURL = val
				case "SKM_API_KEY", "api_key", "API_KEY":
					cfg.APIKey = val
				}
			}
		}
	}

	if envServer := os.Getenv("SKM_SERVER_URL"); envServer != "" {
		cfg.ServerURL = envServer
	}
	if envKey := os.Getenv("SKM_API_KEY"); envKey != "" {
		cfg.APIKey = envKey
	}

	return cfg, nil
}

func SaveCLIConfig(cfg *CLIConfig) error {
	configPath := GetConfigPath()
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory %s: %w", dir, err)
	}

	content := fmt.Sprintf("# SKM Enterprise CLI Configuration\nSKM_SERVER_URL=%q\nSKM_API_KEY=%q\n", cfg.ServerURL, cfg.APIKey)
	return os.WriteFile(configPath, []byte(content), 0600)
}

func runConfig(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "show" || args[0] == "list" {
		cfg, err := LoadCLIConfig()
		if err != nil {
			fmt.Fprintf(stderr, "Failed to load CLI config: %v\n", err)
			return 1
		}
		fmt.Fprintf(stdout, "\nSKM CLI Configuration (~/.skm/.env.toml)\n")
		fmt.Fprintf(stdout, "%s\n", strings.Repeat("=", 60))
		fmt.Fprintf(stdout, "Server URL: %s\n", cfg.ServerURL)
		maskedKey := "<not set>"
		if cfg.APIKey != "" {
			if len(cfg.APIKey) > 8 {
				maskedKey = cfg.APIKey[:4] + "..." + cfg.APIKey[len(cfg.APIKey)-4:]
			} else {
				maskedKey = "***"
			}
		}
		fmt.Fprintf(stdout, "API Key:    %s\n\n", maskedKey)
		return 0
	}

	if args[0] == "set" {
		if len(args) < 3 {
			fmt.Fprintf(stderr, "Usage: skm config set <server|api_key> <value>\n")
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
		default:
			fmt.Fprintf(stderr, "Unknown config key %s. Valid keys: server, api_key\n", key)
			return 1
		}

		if err := SaveCLIConfig(cfg); err != nil {
			fmt.Fprintf(stderr, "Failed to save configuration: %v\n", err)
			return 1
		}
		return 0
	}

	fmt.Fprintf(stderr, "Unknown config command: %s. Use 'skm config show' or 'skm config set <key> <val>'\n", args[0])
	return 1
}
