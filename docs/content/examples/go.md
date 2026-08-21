---
title: Go Integration Examples
description: Standalone Go client integration with modenv cascading TOML configuration and Go enterprise skill packages.
weight: 10
---

# Go Integration Examples

This section details the Go integration examples in `examples/go/`, covering standalone Go module client integration, `modenv` property loading, and Go enterprise skill collections.

---

## 1. Standalone Go Client Example

Located at `examples/go/client`, this example demonstrates how a standalone Go application loads server configuration using `modenv` and resolves skills dynamically via `castor_client`.

### Key Features & Design
- **Zero Bazel Dependency for Native Go**: Contains its own `go.mod` module definition utilizing local relative replaces (`replace github.com/retail-cortex/castor => ../../../`).
- **Cascading TOML Property Resolution**: Uses [`github.com/rrmcguinness/modenv/pkg/modenv`](https://github.com/rrmcguinness/modenv) to load settings from `configs/.env.toml` with environment variable overrides.
- **Dynamic Skills Discovery**: Initializes `castor_client.LoadSkillsFromRoots` and queries `castor_client.LoadSkillsFromPackage`.

### Project Layout

```text
examples/go/client/
├── BUILD.bazel              # Bazel test rule (test_go_client_example)
├── configs/
│   └── .env.toml            # Cascading TOML properties (castor.server.url, etc.)
├── go.mod                   # Standalone Go module configuration
├── main.go                  # Application entry point using modenv & castor_client
└── main_test.go             # Native Go unit test suite
```

### Application Walkthrough (`main.go`)

```go
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/retail-cortex/castor/clients/go/pkg/castor_client"
	"github.com/rrmcguinness/modenv/pkg/modenv"
)

type AppConfig struct {
	Castor struct {
		ServerURL string `toml:"server_url"`
		APIKey    string `toml:"api_key"`
	} `toml:"castor"`
}

func LoadConfiguration() (*AppConfig, error) {
	os.Setenv("MODENV_PREFIX", "configs")
	var cfg AppConfig
	if _, err := modenv.Load(&cfg); err != nil {
		return nil, fmt.Errorf("modenv load error: %w", err)
	}
	if cfg.Castor.ServerURL == "" {
		cfg.Castor.ServerURL = "http://localhost:8080"
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
	scheme, target, ref, subpath := castor_client.ParseSkillRootURI("castor://skills/example.com/testing/test-skill/1.0.0")
	fmt.Printf("Parsed URI: scheme=%s target=%s ref=%s subpath=%s\n", scheme, target, ref, subpath)

	return nil
}

func main() {
	if err := Run(context.Background()); err != nil {
		log.Fatalf("Go Client Example failed: %v", err)
	}
}
```

### Execution Commands

```bash
# Native Go Test Execution
cd examples/go/client
go test -v ./...

# Bazel Workspace Integration Test
bazel test //examples/go/client:test_go_client_example
```

---

## 2. Go Enterprise Skills Package

Located at `examples/go/skills`, this repository provides enterprise SDLC skills tailored for Go software engineering teams:

| Skill Directory | Skill Name | Category | Description |
| :--- | :--- | :--- | :--- |
| `src/retailcortex_skills_go/skills/go-project-setup` | `go-project-setup` | Go | Standard `/cmd`, `/internal`, `/pkg` directory layout and root `Makefile` standards. |
| `src/retailcortex_skills_go/skills/configuration-modenv` | `configuration-modenv` | Go | Idiomatic TOML configuration management using `modenv`. |
| `src/retailcortex_skills_go/skills/go-lang` | `go-lang` | Go | Defensive Go 1.24+ guidelines, table-driven testing with `stretchr/testify`, and interface mocking. |

### Skill Structure Example (`SKILL.md`)

```yaml
---
name: go-project-setup
description: Standardized Go project architecture enforcing /cmd, /internal, /pkg layout, root Makefile, and golangci-lint.
license: Apache-2.0
metadata:
  author: Retail Cortex Engineering
  version: "1.0.0"
compatibility: "go >= 1.22"
allowed-tools:
  - run_command
  - view_file
  - write_to_file
---

# Go Project Setup Skill
...
```
