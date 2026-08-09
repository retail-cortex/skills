# Project Overview: Agent Skill Builder

Welcome to the **Agent Skill Builder** documentation. This registry provides a comprehensive, standardized suite of 33 AI Agent Skills. While built upon the foundational [agentskills.io](https://agentskills.io/specification) specification, this framework **heavily extends** the standard to meet strict enterprise security and performance requirements for the Google Agent Development Kit (ADK).

### Enterprise Extensions
- **Just-in-Time (JIT) Semantic Discovery**: Replaces static loading with RAG-MCP semantic retrieval to prevent LLM context bloat.
- **Compiled References & Schema Strictness**: Strips verbose natural language into cryptographically hashed, strict JSON Schema constraints (`additionalProperties: false`).
- **Cryptographic Manifest Locking (`.manifest.lock`)**: Enforces immutable execution parameters preventing prompt-injected payload tampering.
- **Human-in-the-Loop (HITL) Architecture**: Implements tiered intervention gates and explicit compliance validation components to guarantee Agent-Human Interaction (AHI) safety.

---

## Quickstart: Running Standalone Examples & Agents

Explore native Google Agent Development Kit (ADK) agent execution, qualified URI loading (`file://` and `github://...:branch`), and selective `.env` skill filtering.

### 1. Run Native ADK Example Package (`examples/python/client`)

Run the native ADK agent example demonstrating unified local workspace and remote GitHub skill loading:

```bash
uv run python examples/python/client/main.py
```

### 2. Run Polyglot Developer Agent (`examples/python/polyglot`)

Run the custom polyglot developer CLI agent using domain skills (`skills-bazel`, `skills-go`, `skills-java`, `skills-protobuf`, `skills-python`, `skills-frontend`) to scaffold a Bazel monorepo:

```bash
uv run python examples/python/polyglot/main.py --target-dir ./scratch/my-polyglot-app
```

### 3. Run All Workspace Test Suites

- **Hermetic Bazel Execution (Primary Standard)**:
  ```bash
  bazel test //...
  ```

- **Client-Specific Unit Tests**:
  ```bash
  # Go client tests
  bazel test //clients/go/...
  # Python client tests
  bazel test //clients/python/...
  # Java client tests
  bazel test //clients/java/...
  # Backend service & embedding provider tests
  bazel test //pkg/... //cmd/...
  ```

---

## Documentation Structure

This documentation site is organized into logical sections:

- [Project Overview](./): Introduction, quickstart commands, repository layout, and licensing.
- [Specification](specification/): Enterprise AI Agent Skills Specification (v1.0.0), frontmatter schema, 5-point SDLC compliance, `.manifest.lock` cryptographic integrity, and polyglot URI resolution.
- [Architecture](architecture/): Engineering standards, pluggable embedding providers, pgvector poly-column schemas, Google OAuth2 integration, and HTTP 429 rate limit resilience.
- [Cloud Deployment](deployment/): Enterprise infrastructure automation via Terraform, GKE clusters (`dev`, `qa`, `prod`), AlloyDB AI, and Kustomize overlays.
- [Critical Analysis](analysis/): Comparative analysis against agentskills.io specification and ecosystem showcase clients.
- [Skills Registry](examples/skills/): Specialized domain and technology enterprise skills catalog.
- [CLI Client (skm)](cli/): Standalone `skm` Go CLI client manual, cross-platform builds, polyglot URI resolution, subcommands, and Oh My Zsh plugin.
- [Packages & Architecture](packages/): Backend services (`skills-service`), Go CLI (`skm`), core packages (`pkg/embedding`, `pkg/data`, `pkg/mcp`, `pkg/service`), polyglot client SDKs, and Protocol Buffers.
- [Examples](examples/): Standalone integration packages, polyglot clients, web servers, and setup guides for Go, Python, and Java.

---

## Workspace Directory Layout

The project is governed by a root [MODULE.bazel](https://github.com/retail-cortex/skills/blob/main/MODULE.bazel) (Bazel 9.2) and a root Python 3.13 `uv` workspace:

```text
skill-builder/
├── MODULE.bazel               # Bazel 9.2 Bzlmod module definition
├── BUILD.bazel                # Root Bazel aliases and targets
├── pyproject.toml             # Root uv workspace configuration
├── hugo.toml                  # Hugo Geekdoc documentation site configuration
├── docs/                      # Documentation site source files
├── cmd/                       # Microservice & CLI entry points
│   ├── skm/                   # SKM (Skill Manager) Go CLI package manager
│   └── skills-service/        # Central Go REST, MCP, and gRPC backend service
├── pkg/                       # Shared Go domain packages
│   ├── data/                  # GORM repository with pgvector poly-column schema
│   ├── embedding/             # Embedding provider interface & sliding chunking
│   │   ├── alloydb/           # AlloyDB AI in-database embedding provider
│   │   └── vertex/            # Vertex AI & Gemini Developer API embedding provider
│   ├── mcp/                   # Model Context Protocol SSE/stdio server
│   ├── model/                 # Shared domain data models and pagination DTOs
│   └── service/               # Application services & background embedding workers
├── proto/                     # Protocol Buffers IDL contracts
│   └── retailcortex/          # Enterprise service definitions (skills, registration)
├── clients/                   # Polyglot Client SDKs & Build Plugins
│   ├── go/                    # Go client (skillsloader) with //go:generate
│   ├── java/                  # Java client & skills-loader-maven-plugin
│   └── python/                # Python client (loader) with PEP 517 build_meta
├── examples/                  # Language integration examples & skill packages
│   ├── go/                    # Standalone Go client & Go skill collection
│   ├── java/                  # Standalone Java client & Java skill collection
│   ├── python/                # Standalone Python client, Polyglot Agent & skills
│   └── skills/                # Standalone multi-content test skills
├── LICENSE                    # Apache 2.0 License
├── NOTICE                     # Legal attribution notices
└── validator_report.json      # Persisted 5-point SDLC audit results
```

---

## License & Legal Notices

This project is licensed under the Apache License, Version 2.0. See [LICENSE](https://github.com/retail-cortex/skills/blob/main/LICENSE) for details. Attribution notices are maintained in [NOTICE](https://github.com/retail-cortex/skills/blob/main/NOTICE).
