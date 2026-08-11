---
title: Go Integration Examples
description: Standalone Go client integration with modenv cascading TOML configuration and Go enterprise skill packages.
weight: 10
---

# Go Integration Examples

This section details the Go integration examples in `examples/go/`, covering standalone Go module client integration, `modenv` property loading, and Go enterprise skill collections.

---

## 1. Standalone Go Client Example

Located at `examples/go/client`, this example demonstrates how a standalone Go application loads server configuration using `modenv` and resolves skills dynamically via `skillsloader`.

### Key Features & Design
- **Zero Bazel Dependency for Native Go**: Contains its own `go.mod` module definition utilizing local relative replaces (`replace github.com/retail-cortex/skills => ../../../`).
- **Cascading TOML Property Resolution**: Uses [`github.com/rrmcguinness/modenv/pkg/modenv`](https://github.com/rrmcguinness/modenv) to load settings from `configs/.env.toml` with environment variable overrides.
- **Dynamic Skills Discovery**: Initializes `skillsloader.LoadSkillsFromRoots` and queries `skillsloader.LoadSkillsFromPackage`.

### Project Layout

```text
examples/go/client/
├── BUILD.bazel              # Bazel test rule (test_go_client_example)
├── configs/
│   └── .env.toml            # Cascading TOML properties (castor.server.url, etc.)
├── go.mod                   # Standalone Go module configuration
├── main.go                  # Application entry point using modenv & skillsloader
└── main_test.go             # Native Go unit test suite
```

### Application Walkthrough (`main.go`)

```go
package main

import (
	"fmt"

	"github.com/retail-cortex/skills/clients/go/pkg/skillsloader"
	"github.com/rrmcguinness/modenv/pkg/modenv"
)

func main() {
	// 1. Cascading TOML Configuration via modenv
	cfg, err := modenv.Load("configs/.env.toml")
	if err != nil {
		fmt.Printf("Defaulting configuration: %v\n", err)
	}

	serverURL := cfg.GetString("castor.server.url", "http://localhost:8080")
	fmt.Printf("Connected to Castor Server at: %s\n", serverURL)

	// 2. Load Enterprise Skills
	skills, err := skillsloader.LoadSkillsFromPackage("retailcortex_skills_go", []string{"go-project-setup"})
	if err != nil {
		panic(err)
	}

	for name, def := range skills {
		fmt.Printf("Loaded Go Skill [%s]: %s\n", name, def.Description)
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
