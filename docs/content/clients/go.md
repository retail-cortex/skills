---
title: "Go Client & Toolchain Integration"
weight: 20
---

# Go Client & Toolchain Integration (`skillsloader`)

The Go client library (`github.com/retail-cortex/skills/clients/go/pkg/skillsloader`) integrates into Go's native build workflows via **`//go:generate` directives**, **`go test` validation hooks**, and **Bazel `rules_go`** targets.

---

## 1. Native `go generate` & Build Integration

In corporate Go codebases, skill compilation and validation are bound directly to the build phase using `//go:generate` directives in your main package or service package:

```go
package main

// Generate pre-compiled zero-I/O manifest prior to compilation
//go:generate skm compile -d ./skills -o ./skills_manifest.json

// Verify cryptographic integrity of skills lockfile during build
//go:generate skm verify -d ./skills

import (
	"embed"
	"fmt"
	"log"

	"github.com/retail-cortex/skills/clients/go/pkg/skillsloader"
)

// Embed pre-compiled skills manifest into static Go binary
//go:embed skills_manifest.json
var embeddedManifest embed.FS

func main() {
	manifestData, err := embeddedManifest.ReadFile("skills_manifest.json")
	if err != nil {
		log.Fatalf("Failed to read embedded manifest: %v", err)
	}

	skills, err := skillsloader.LoadSkillsFromManifestData(manifestData)
	if err != nil {
		log.Fatalf("Corrupt manifest: %v", err)
	}

	fmt.Printf("Instantly loaded %d embedded skills into Go binary.\n", len(skills))
}
```

### Build Execution Flow (`go generate ./...` & `go build ./...`)

```bash
# 1. Trigger pre-build generation & validation hooks
go generate ./...

# 2. Compile statically linked binary with embedded skills
go build -o agent_service ./cmd/agent
```

If `skm verify` or `skm compile` detects corrupt frontmatter or checksum mismatches during `go generate`, **the build pipeline terminates before binary compilation**.

---

## 2. Hermetic Bazel Build Integration (`rules_go`)

In your `BUILD.bazel`:

```starlark
load("@rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "agent_lib",
    srcs = ["main.go"],
    embedsrcs = ["skills_manifest.json"],
    importpath = "com.company.agent/lib",
    deps = [
        "//clients/go/pkg/skillsloader",
    ],
)

go_binary(
    name = "agent_service",
    embed = [":agent_lib"],
)
```

---

## 3. Automated Unit Test Validation (`go test`)

Incorporate skill verification into your automated `_test.go` suites:

```go
package main_test

import (
	"testing"

	"github.com/retail-cortex/skills/clients/go/pkg/skillsloader"
	"github.com/stretchr/testify/assert"
)

func TestSkillsIntegrity(t *testing.T) {
	report, err := skillsloader.VerifySkills("./skills", "./skills/.manifest.lock")
	assert.NoError(t, err)
	assert.Equal(t, 0, report.ModifiedCount, "Skills directory has been tampered with")
	assert.Equal(t, 0, report.MissingCount, "Skills files are missing")
}
```

## 4. Integrating Skills with Google ADK Agents

Loaded `SkillDefinition` objects map directly to Google ADK Agent system instructions, prompt guidelines, and executable tools. Environment and server properties are loaded via `modenv` (`github.com/rrmcguinness/modenv/pkg/modenv`):

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/adk/agent"
	"github.com/retail-cortex/skills/clients/go/pkg/skillsloader"
	"github.com/rrmcguinness/modenv/pkg/modenv"
)

type AppConfig struct {
	SKM struct {
		ServerURL string `toml:"server_url"`
		APIKey    string `toml:"api_key"`
	} `toml:"skm"`
}

func main() {
	ctx := context.Background()

	// 1. Load application properties from cascading TOML configuration using modenv
	var cfg AppConfig
	if _, err := modenv.Load(&cfg); err != nil {
		log.Printf("Warning: failed to load modenv configuration: %v", err)
	}

	serverURL := cfg.SKM.ServerURL
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}

	// 2. Load skill definition from central SKM server using loaded modenv properties
	skills, err := skillsloader.LoadSkillsFromSKMServer("sk-9b1deb4d", nil, serverURL, cfg.SKM.APIKey)
	if err != nil {
		log.Fatalf("Failed to load skill: %v", err)
	}
	skill := skills["gemini-api"]

	// 3. Instantiate Google ADK Agent grounded in skill instructions
	adkAgent := agent.New(agent.Config{
		Name:               skill.Name,
		Model:              "gemini-2.0-flash",
		SystemInstructions: skill.Instructions,
	})

	// 4. Run prompt through ADK agent pipeline
	res, err := adkAgent.Run(ctx, "Synthesize demand forecasting query for store #42")
	if err != nil {
		log.Fatalf("ADK execution error: %v", err)
	}

	fmt.Println("ADK Agent Response:", res.Text)
}
```


---

## Best Practices for Go Services

1. **Distroless Packaging**: Statically link Go binaries (`CGO_ENABLED=0`) with embedded skill manifests for deployment in minimal `scratch` or `distroless` container images.
2. **Runtime Fallback**: Configure runtime fallback to `LoadSkillsFromSKMServer` if local embedded skills are missing.
3. **Hermetic Testing**: Mock external HTTP requests in unit tests using `httptest.NewServer`.

