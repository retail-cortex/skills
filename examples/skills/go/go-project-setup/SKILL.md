---
name: go-project-setup
description: Elite meta-skill for scaffolding enterprise Go microservices using /cmd, /internal, /pkg, /api layout wrapped in Bazel rules_go/gazelle. Enforces table-driven TDD, HTTP 429 rate limit backoff, and GCP Terraform.
license: Apache-2.0
metadata:
  author: Ryan McGuinness
  version: "1.0"
authors:
  - name: Retail Cortex Engineering
    url: https://github.com/retail-cortex/skills
category: go
tags:
  - go
  - scaffolding
  - bazel
  - cmd-internal
trigger_phrases:
  - "scaffold Go project"
  - "Go project layout setup"
  - "cmd internal Go structure"
execution_hints:
  preferred_model: "gemini-3.1-pro"
  requires_human_approval: false
  environment_variables:
    - GOPATH
  timeout_seconds: 240
---
# Go Project Setup Meta-Skill (Enterprise Layout Standard)

This meta-skill provides automated scaffolding instructions and templates for initializing enterprise **Go** microservices adhering to compiler-enforced package boundaries, **Bazel** hermetic builds, and dedicated `configs/` manifests.

## Prerequisites & Pre-Flight Checklist

1. Go 1.26+ installed locally.
2. Bazelisk installed on system PATH.
3. Access to GCP environment for GKE/AlloyDB provisioning.

## HTTP 429 Rate Limit & Quota Resilience Invariants

- Outbound API calls to external GCP services must use `hashicorp/go-retryablehttp` with exponential backoff and randomized jitter to handle HTTP 429 quota exhaustion.

## Security Checkpoints & CWE Invariants

- **CWE-200 (Information Exposure)**: Passwords in `configs/.env.toml` MUST use in-memory XOR decryption (`xor:...`) via `modenv`; plain-text secrets in git are strictly prohibited.
- **CWE-89 (SQL Injection)**: Database queries MUST use GORM parameterized bindings or prepared statements.
- **CWE-250 (Execution with Unnecessary Privileges)**: Final container runtimes MUST execute as non-root inside `scratch` or `distroless` images.

## 3-Phase Execution Protocol

### Phase 1: Initialize Directory Tree
```bash
go mod init github.com/enterprise/service
mkdir -p cmd/server internal/server internal/database pkg api/proto configs/terraform .github/workflows
```

### Phase 2: Add Dependencies & Configure Bazel
```bash
go get github.com/gin-gonic/gin google.golang.org/grpc gorm.io/gorm gorm.io/driver/postgres github.com/rrmcguinness/modenv google.golang.org/genai github.com/stretchr/testify github.com/hashicorp/go-retryablehttp
bazel run //:gazelle
```

### Phase 3: Run TDD Suite (85% Coverage) & Compile Statically
```bash
go test -v -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
bazel build //...
```

## Progressive Disclosure References

- **Go Scaffold Guide**: Read [`references/go_scaffold_guide.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/go-project-setup/references/go_scaffold_guide.md).
- **Reference Go Mod**: View [`examples/go.mod`](file:///Users/rmcguinness/Projects/skill-builder/skills/go-project-setup/examples/go.mod).
- **Reference Main**: View [`examples/cmd/server/main.go`](file:///Users/rmcguinness/Projects/skill-builder/skills/go-project-setup/examples/cmd/server/main.go).
- **Reference Bazel Build**: View [`examples/BUILD.bazel`](file:///Users/rmcguinness/Projects/skill-builder/skills/go-project-setup/examples/BUILD.bazel).
- **Reference Terraform**: View [`examples/configs/terraform/main.tf`](file:///Users/rmcguinness/Projects/skill-builder/skills/go-project-setup/examples/configs/terraform/main.tf).
- **Reference Base Config**: View [`examples/configs/.env.toml`](file:///Users/rmcguinness/Projects/skill-builder/skills/go-project-setup/examples/configs/.env.toml).
