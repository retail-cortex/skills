---
title: "Packages & Core Libraries"
weight: 60
---

# Packages, Microservices & Tooling

The Castor repository is structured into backend microservices, standalone CLI tools, polyglot client libraries, and protocol buffer definitions across Go, Java, and Python.

---

## 1. Backend Microservice (`cmd/castor_server`)

`castor-server` is the central enterprise `Castor Registry` backend microservice written in Go (`cmd/castor_server`).

### Key Capabilities

- **Gin REST & gRPC API Endpoints**: Exposes REST endpoints (`POST /api/v1/skills`, `GET /api/v1/skills/:id`, `POST /api/v1/apps/verify`) and gRPC servicers (`pkg/service`).
- **Database Persistence**: GORM repository layer (`pkg/data`) for SQLite and AlloyDB storage of apps and skills.
- **Model Context Protocol (MCP)**: Embedded stdio and SSE MCP server (`pkg/mcp`) for agent tool execution.
- **Canonical URI Generation**: Assigns `castor://skills/{domain}/{category}/{name}/{version}` URIs upon skill registration.

### Execution

```bash
# Run backend service directly via Bazel
bazel run //:castor-server
```

---

## 2. Standalone CLI Client (`cmd/cstr`)

`cstr` is the enterprise `Castor CLI` package manager built in Go (`cmd/cstr`) supporting cross-compilation for Linux, macOS, and Windows.

### Core Commands

- **`cstr config`**: Configures CLI settings (`CASTOR_SERVER_URL`, `CASTOR_API_KEY`) in `~/.castor/.env.toml`.
- **`cstr register`**: Registers source skills (`github://`, `file://`) to `Castor Registry`.
- **`cstr add`**: Resolves and installs skill dependencies (`castor://`, `cstr://`, `github://`, `mod://`, `maven://`, `pkg://`, `file://`) into `.skills/`.
- **`cstr verify`**: Cryptographically audits installed skills against `.manifest.lock`.
- **`cstr validate`**: Runs 5-point SDLC compliance audit against skill directories.
- **`cstr compile`**: Compiles skills into pre-compiled zero-I/O `skills_manifest.json`.
- **`cstr list` & `cstr search`**: Discovers and searches skills in local and remote registries.
- **`cstr init`**: Scaffolds new compliant skill directory trees.

### Execution

```bash
# Run CLI via Bazel
bazel run //:cstr -- help

# Build all platform binaries
bazel build //cmd/cstr:cstr_binaries
```

---

## 3. Polyglot Client Libraries

Client libraries integrate directly into native build cycles to validate, resolve, and package skills into application artifacts:

1. **Go Client (`clients/go/pkg/castor_client`)**: Go module client (`castor_client`), `//go:generate` pre-compilation, `modenv` property loading, and `rules_go` Bazel rules.
2. **Java Client (`clients/java`)**: Java client (`com.retailcortex.castor.client`) and native Maven Plugin (`castor-client` / `GenerateManifestMojo`), binding to `mvn compile`, Java System properties (`System.getProperty`), and `rules_jvm_external`.
3. **Python Client (`clients/python`)**: Python client package (`castor-client`) with PEP 517 build backend wrapper (`castor_client.build_meta`), binding to `uv build` / `pip install`, `python-dotenv`, and `rules_python`.

---

## 4. Core Service Packages (`pkg/`)

The backend microservice and client tooling rely on shared Go libraries in `pkg/`:

- **[`pkg/embedding`](file:///Users/rmcguinness/Projects/retail-cortex/castor/pkg/embedding/provider.go)**: Standardized embedding provider interfaces ([`embedding.Provider`](file:///Users/rmcguinness/Projects/retail-cortex/castor/pkg/embedding/provider.go#L26)), sliding-window chunking algorithms ([`SplitTextIntoChunks`](file:///Users/rmcguinness/Projects/retail-cortex/castor/pkg/embedding/provider.go#L70)), cosine similarity computations, and deterministic offline vector simulation.
  - **[`pkg/embedding/vertex`](file:///Users/rmcguinness/Projects/retail-cortex/castor/pkg/embedding/vertex/vertex.go)**: Google Vertex AI (`multimodalembedding`, `text-embedding-004`) and Gemini Developer API provider with cached OAuth2 ADC authentication.
  - **[`pkg/embedding/alloydb`](file:///Users/rmcguinness/Projects/retail-cortex/castor/pkg/embedding/alloydb/alloydb.go)**: Google Cloud AlloyDB AI in-database SQL embedding provider (`SELECT embedding(...)`).
- **[`pkg/data`](file:///Users/rmcguinness/Projects/retail-cortex/castor/pkg/data/castor_repository.go)**: GORM persistence repository ([`CastorRepository`](file:///Users/rmcguinness/Projects/retail-cortex/castor/pkg/data/castor_repository.go#L24)) supporting SQLite and PostgreSQL/AlloyDB `pgvector` poly-column vector schemas (`embedding_768`, `embedding_1408`, `embedding_3072`) with HNSW index management.
- **[`pkg/mcp`](file:///Users/rmcguinness/Projects/retail-cortex/castor/pkg/mcp/server.go)**: Model Context Protocol (MCP) server integration supporting SSE and stdio transports for autonomous AI agent tool execution.
- **[`pkg/service`](file:///Users/rmcguinness/Projects/retail-cortex/castor/pkg/service/castor_service.go)**: Business logic service ([`CastorService`](file:///Users/rmcguinness/Projects/retail-cortex/castor/pkg/service/castor_service.go#L44)) for skill management, application registration, OpenTelemetry distributed tracing, and non-blocking background embedding worker pools.
- **[`pkg/model`](file:///Users/rmcguinness/Projects/retail-cortex/castor/pkg/model/skill.go)**: Domain data structures, DTOs, and REST pagination envelope models.

---

## 5. Protocol Buffers IDL (`proto/`)

The authoritative service IDL and domain contracts are defined in Protocol Buffers:

- `proto/castor/skills/v1/skill.proto`: Core domain definitions (`SkillDefinition`, `SkillSummary`).
- `proto/castor/skills/v1/skill_service.proto`: gRPC and REST endpoints for `SkillService`.
- `proto/castor/registration/v1/registration_service.proto`: Application verification and RBAC collaborator endpoints.
- `proto/castor/skills/v1/manifest.proto`: Manifest serialization contracts.
