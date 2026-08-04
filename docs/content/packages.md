# Packages, Microservices & Tooling

The Skill Builder repository is structured into backend microservices, standalone CLI tools, polyglot client libraries, and protocol buffer definitions across Go, Java, and Python.

---

## 1. Backend Microservice (`cmd/skills-service`)

`skills-service` is the central enterprise backend microservice written in Go (`cmd/skills-service`).

### Key Capabilities

- **Gin REST & gRPC API Endpoints**: Exposes REST endpoints (`POST /api/v1/skills`, `GET /api/v1/skills/:id`, `POST /api/v1/apps/verify`) and gRPC servicers (`pkg/service`).
- **Database Persistence**: GORM repository layer (`pkg/data`) for SQLite and AlloyDB storage of apps and skills.
- **Model Context Protocol (MCP)**: Embedded stdio and SSE MCP server (`pkg/mcp`) for agent tool execution.
- **Canonical URI Generation**: Assigns `skm://skills/{skill_id}` URIs upon skill registration.

### Execution

```bash
# Run backend service directly via Bazel
bazel run //:skills-service
```

---

## 2. Standalone CLI Client (`cmd/skm`)

`skm` is the enterprise CLI package manager built in Go (`cmd/skm`) supporting cross-compilation for Linux, macOS, and Windows.

### Core Commands

- **`skm config`**: Configures CLI settings (`SKM_SERVER_URL`, `SKM_API_KEY`) in `~/.skm/.env.toml`.
- **`skm register`**: Registers source skills (`github://`, `file://`) to `skills-service`.
- **`skm add`**: Resolves and installs skill dependencies (`skm://`, `github://`, `mod://`, `maven://`, `pkg://`, `file://`) into `.skills/`.
- **`skm verify`**: Cryptographically audits installed skills against `.manifest.lock`.
- **`skm validate`**: Runs 5-point SDLC compliance audit against skill directories.
- **`skm compile`**: Compiles skills into pre-compiled zero-I/O `skills_manifest.json`.
- **`skm list` & `skm search`**: Discovers and searches skills in local registries.
- **`skm init`**: Scaffolds new compliant skill directory trees.

### Execution

```bash
# Run CLI via Bazel
bazel run //:skm -- help

# Build all platform binaries
bazel build //cmd/skm:skm_binaries
```

---

## 3. Polyglot Client Libraries

Client libraries integrate directly into native build cycles to validate, resolve, and package skills into application artifacts:

1. **Go Client (`clients/go/pkg/skillsloader`)**: Go module loader, `//go:generate` pre-compilation, `modenv` property loading, and `rules_go` Bazel rules.
2. **Java Client (`clients/java`)**: Java client and native Maven Plugin (`skills-loader-maven-plugin` / `GenerateManifestMojo`), binding to `mvn compile`, Java System properties (`System.getProperty`), and `rules_jvm_external`.
3. **Python Client (`clients/python`)**: Python loader package (`retailcortex-loader`) with PEP 517 build backend wrapper (`loader.build_meta`), binding to `uv build` / `pip install`, `python-dotenv`, and `rules_python`.

---

## 4. Protocol Buffers IDL (`proto/`)

The authoritative service IDL and domain contracts are defined in Protocol Buffers:

- `proto/retailcortex/skills/v1/skill.proto`: Core domain definitions (`SkillDefinition`, `SkillSummary`).
- `proto/retailcortex/skills/v1/skill_service.proto`: gRPC and REST endpoints for `SkillService`.
- `proto/retailcortex/registration/v1/registration_service.proto`: Application verification endpoints.
