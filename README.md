# Agent Skill Builder: Enterprise Project Pattern Registry

This registry provides a comprehensive, standardized suite of 23 AI Agent Skills built strictly in compliance with the [agentskills.io](https://agentskills.io/specification) specification and Google Agent Development Kit (ADK) progressive disclosure architecture.

## Rigorous Engineering, Google OAuth2 & Resilience Invariants

Every skill in this registry enforces a complete, production-grade Software Development Lifecycle (SDLC) with zero-tolerance security, Google OAuth2 authentication, and HTTP 429 rate limit resilience:

1. **Google OAuth2 Identity Provider (UI & APIs)**:
   - **Frontend UI (React 19 / Next.js)**: Standardized on Google Identity Services (GIS) and Authorization Code Flow with PKCE (`@react-oauth/google`). Access tokens stored in secure, encrypted, `HttpOnly`, `SameSite=Lax` cookies; unencrypted tokens in `localStorage` are prohibited (CWE-79).
   - **Backend APIs (Python, Go, Java 25)**: Validate Google Bearer ID tokens against Google JWKS (`google-auth`, `google.golang.org/api/idtoken`, `GoogleIdTokenVerifier`).
   - **User Token Delegation**: Backend agents delegate authenticated user OAuth2 tokens into `ToolContext.state["user_token"]` to execute BigQuery CAPI and storage tools under the end user's IAM permissions (CWE-269).
2. **HTTP 429 Rate Limit & Quota Resilience (The AI API Invariant)**:
   - All AI-driven API calls (Gemini, BigQuery CAPI, Vertex AI) MUST implement **Exponential Backoff with Full Randomized Jitter** using libraries like `tenacity` (Python), `hashicorp/go-retryablehttp` (Go), and `Resilience4j` (Java).
   - Inbound endpoints protect against quota exhaustion via token bucket rate limiters (`slowapi`, `golang.org/x/time/rate`, `Bucket4j`).
   - Frontend clients catch 429 responses, parse `Retry-After` headers, and render live countdown timers on disabled action triggers.
3. **Paired Positive, Negative & Boundary TDD**:
   - Every feature and service component MUST be tested against paired scenarios:
     - **Positive / Happy Path**: Valid inputs, expected schemas, 200/201 HTTP status codes.
     - **Negative / Error State**: Missing/expired OAuth tokens, 400/401/404/422/429/500 HTTP responses, database connection timeouts, and exception propagation.
     - **Empty Value Boundaries**: Empty strings `""`, whitespace `"   "`, empty lists `[]`, zero numeric counts `0`, and missing dictionary keys.
4. **Defensive Error Handling & Null/Nil Safety**:
   - **Go**: Strict `nil` checks on pointers/slices/maps; defensive error wrapping (`fmt.Errorf("...: %w", err)`).
   - **Java 25 (LTS)**: Enforce `java.util.Optional<T>` return types, `@NonNull` parameter annotations, SpotBugs `NP_NULL_ON_SOME_PATH` analysis, and SLF4J 2.x structured logging (zero `printStackTrace()`).
   - **Python 3.13**: Strict type annotations (`T | None`), prohibition of `Any`, defensive `.get()` dictionary defaults, and Pydantic v2 input bounds.
   - **TypeScript / React 19**: Enforce `strictNullChecks: true`, optional chaining (`?.`), and nullish coalescing (`??`).
5. **Java 25 (LTS) Standard**:
   - Standardized strictly on **Java 25 (LTS)** (`<maven.compiler.release>25</maven.compiler.release>`) leveraging virtual threads and modern records.
   - Dependencies managed via the latest Google Cloud `libraries-bom`, Javalin 6+, SLF4J 2.x, Jackson 2.18+, and wrapped in Bazel `rules_java`.
6. **Build System Interoperability (Bazel 9.2 Overarching Standard)**:
   - Every language standardizes on **Bazel 9.2** for hermetic CI/CD and monorepo execution:
     - **Python (3.13+)**: Managed locally via **uv** (`pyproject.toml`), wrapped in Bazel `rules_python` (`1.7.0`).
     - **Node.js (22+) / React (19+)**: Managed locally via **pnpm** (`pnpm-lock.yaml`), wrapped in Bazel `aspect_rules_js` / `npm.npm_translate_lock`.
     - **Java (25 LTS)**: Managed locally via **Maven** (`pom.xml`), wrapped in Bazel `rules_java`.
     - **Go (1.26+)**: Managed locally via **Go modules** (`go.mod`), wrapped in Bazel `rules_go` and `gazelle`.
7. **ADK Superior Deployment Pattern**:
   - Google ADK agents standardly wrap agent execution runners inside **FastAPI** web services served asynchronously via **Uvicorn** for streaming SSE and high concurrency.
8. **Hierarchical TOML Configuration & Security**:
   - Standardized multi-tier cascading TOML configurations (`.env.toml`, `.env.local.toml`, `.env.${RUNTIME}.toml`) using `modenv` and in-memory XOR secret encryption (`xor:...`).

---

## Workspace Structure

The project is governed by a root [MODULE.bazel](file:///Users/rmcguinness/Projects/skill-builder/MODULE.bazel) (Bazel 9.2) and a root Python 3.13 `uv` workspace:

```
skill-builder/
├── MODULE.bazel               # Bazel 9.2 Bzlmod module definition
├── BUILD.bazel                # Root Bazel aliases and filegroups
├── .bazelignore               # Bazel directory exclusions
├── pyproject.toml             # Root uv workspace configuration
├── Makefile                   # Developer convenience wrapper
├── LICENSE                    # Apache 2.0 License
├── NOTICE                     # Legal attribution notices
├── validator_report.json      # Persisted 5-point SDLC audit results
├── packages/                  # Workspace utility packages
│   └── validator/             # 5-point SDLC validator package
│       └── BUILD.bazel        # Python library, binary, and test targets
└── skills/                    # Specialized AI Agent Skills
    ├── a2a/
    ├── a2ui/
    ├── ...
    └── terraform-gcp/
```

---

## Scaffolding & Setup Meta-Skills

| Meta-Skill | Technology & Build Stack | Description |
| :--- | :--- | :--- |
| [mono-repo-setup](file:///Users/rmcguinness/Projects/skill-builder/skills/mono-repo-setup/SKILL.md) | Bazel 8/9 Bzlmod Polyglot | Scaffolding for polyglot monorepos (Go 1.26+, Python 3.13 uv, Java 25 Maven, TypeScript React 19 pnpm), root Bazel configs, and `configs/terraform/`. |
| [python-project-setup](file:///Users/rmcguinness/Projects/skill-builder/skills/python-project-setup/SKILL.md) | Python 3.13, uv & Bazel | Scaffolding for enterprise Python services using `uv`, wrapped in Bazel `rules_python`, with Google OAuth2, None safety, and cascading TOML. |
| [go-project-setup](file:///Users/rmcguinness/Projects/skill-builder/skills/go-project-setup/SKILL.md) | Go 1.26+ & Bazel | Scaffolding for Go microservices using standard `/cmd`, `/internal`, `/pkg`, `/api` layout, Bazel `rules_go`/`gazelle`, nil safety, and `configs/terraform/`. |
| [java-project-setup](file:///Users/rmcguinness/Projects/skill-builder/skills/java-project-setup/SKILL.md) | Java 25 (LTS), Maven & Bazel | Scaffolding for Java 25 microservices using Maven POM with latest GCP BOM, wrapped in Bazel `rules_java`, with Optional/NonNull safety and `configs/terraform/`. |

---

## Domain & Technology Skills

| Skill Directory | Technology / Domain | SDLC, Google OAuth2, Error Handling & 429 Highlights |
| :--- | :--- | :--- |
| [bazel-modules](file:///Users/rmcguinness/Projects/skill-builder/skills/bazel-modules/SKILL.md) | Bazel 8/9 & Bzlmod | Bzlmod dependency graphs, hermetic TDD, `rules_hugo` GitHub Pages publishing, coverage collection, and SemVer. |
| [react-vite](file:///Users/rmcguinness/Projects/skill-builder/skills/react-vite/SKILL.md) | React 19 & Vite 6 | Google OAuth2 PKCE/GIS, HTTP 429 countdown banners, `strictNullChecks`, optional chaining, paired Vitest TDD, 85% coverage. |
| [python-core](file:///Users/rmcguinness/Projects/skill-builder/skills/python-core/SKILL.md) | Python 3.13 & UV | None safety, paired positive/negative pytest TDD, 90% coverage enforcement, MkDocs on Pages, and SemVer. |
| [python-fastapi](file:///Users/rmcguinness/Projects/skill-builder/skills/python-fastapi/SKILL.md) | FastAPI Enterprise | Google OAuth2 Bearer token verification, slowapi 429 rate limiting, async REST TDD with httpx, 90% coverage, and SemVer. |
| [python-fastmcp](file:///Users/rmcguinness/Projects/skill-builder/skills/python-fastmcp/SKILL.md) | FastMCP Protocol | FastMCP tool TDD, GitHub Actions schema validation, prompt injection defense, Pydantic bounds, and SemVer. |
| [python-adk-fastapi](file:///Users/rmcguinness/Projects/skill-builder/skills/python-adk-fastapi/SKILL.md) | FastAPI & Google ADK | Superior pattern wrapping ADK in FastAPI on Uvicorn, Google OAuth2 user token delegation to CAPI, tenacity 429 backoff with jitter. |
| [go-lang](file:///Users/rmcguinness/Projects/skill-builder/skills/go-lang/SKILL.md) | Go 1.26+ Services | Google OAuth2 ID token verification middleware, retryablehttp 429 backoff, nil pointer safety, paired table-driven TDD. |
| [alloydb](file:///Users/rmcguinness/Projects/skill-builder/skills/alloydb/SKILL.md) | AlloyDB & pgvector | Testcontainers TDD, migration CI/CD validation, private IP VPC security, SSL enforcement, and SemVer schema versioning. |
| [bigquery](file:///Users/rmcguinness/Projects/skill-builder/skills/bigquery/SKILL.md) | BigQuery & CAPI | Dry-run query TDD, SQLFluff linting, GE OAuth security delegation, Protobuf normalization, and SemVer. |
| [opentelemetry-google](file:///Users/rmcguinness/Projects/skill-builder/skills/opentelemetry-google/SKILL.md) | OpenTelemetry & GCP Trace | InMemorySpanExporter TDD, CI/CD pipeline verification, PII scrubbing security, batch flush lifecycles, and SemVer. |
| [terraform-gcp](file:///Users/rmcguinness/Projects/skill-builder/skills/terraform-gcp/SKILL.md) | Terraform for GCP | `terraform test` TDD, GitHub Actions CI/CD with tflint, GCS remote state security, least-privilege IAM, and SemVer. |
| [configuration-modenv](file:///Users/rmcguinness/Projects/skill-builder/skills/configuration-modenv/SKILL.md) | Configuration & modenv | Cascading TOML precedence TDD, GitHub Actions CI schema validation, XOR cipher memory decryption, and SemVer. |
| [java-enterprise](file:///Users/rmcguinness/Projects/skill-builder/skills/java-enterprise/SKILL.md) | Java 25 (LTS) & Javalin | Google OAuth2 ID token verifier, Resilience4j 429 retries, Optional/NonNull null safety, JUnit 5 TDD, JaCoCo 85% coverage. |
| [a2a](file:///Users/rmcguinness/Projects/skill-builder/skills/a2a/SKILL.md) | A2A Protocol & Orchestration | Component TDD, GitHub Actions CI validation, token bucket rate limiting, HMAC message auth (CWE-306), and SemVer. |
| [a2ui](file:///Users/rmcguinness/Projects/skill-builder/skills/a2ui/SKILL.md) | A2UI v0.8 & Gemini Enterprise | Sandboxed iframe CSP policies (CWE-1021), postMessage origin checks (CWE-346), right-aligned actions with `justify-end`, and component TDD. |
| [docker-containers](file:///Users/rmcguinness/Projects/skill-builder/skills/docker-containers/SKILL.md) | Multi-Stage Docker | Container health check TDD, GitHub Actions Trivy vulnerability scanning, non-root user security, and immutable SemVer tags. |
| [nx-monorepo](file:///Users/rmcguinness/Projects/skill-builder/skills/nx-monorepo/SKILL.md) | NX Polyglot Monorepo | Affected TDD testing, GitHub Actions CI with remote caching, module boundary security rules, and SemVer release orchestration. |
| [protobuf-grpc](file:///Users/rmcguinness/Projects/skill-builder/skills/protobuf-grpc/SKILL.md) | Protobuf & gRPC | In-process gRPC TDD, Buf breaking change detection, mTLS/auth security, Bazel rules_proto_grpc, and SemVer. |
| [adk-skill-factory](file:///Users/rmcguinness/Projects/skill-builder/skills/adk-skill-factory/SKILL.md) | ADK Meta-Skill Factory | ADK evaluator TDD, GitHub Actions schema validation, human-in-the-loop security reviews, progressive disclosure, and SemVer. |

---

## Bazel 9.2 Build & Test Targets

The primary build interface is powered by Bazel 9.2 ([MODULE.bazel](file:///Users/rmcguinness/Projects/skill-builder/MODULE.bazel)) with root convenience aliases defined in [BUILD.bazel](file:///Users/rmcguinness/Projects/skill-builder/BUILD.bazel):

- **Run Full 5-Point SDLC Validator**:
  ```bash
  bazel run //:validate
  ```
  Executes the validation binary ([//packages/validator:validate_skills](file:///Users/rmcguinness/Projects/skill-builder/packages/validator/BUILD.bazel)) to audit all 23 skills in [skills/](file:///Users/rmcguinness/Projects/skill-builder/skills) and persist [validator_report.json](file:///Users/rmcguinness/Projects/skill-builder/validator_report.json).

- **Run Bazel Test Suites**:
  ```bash
  bazel test //...
  ```
  Executes hermetic test targets across all workspace packages including [//packages/validator:test_validator](file:///Users/rmcguinness/Projects/skill-builder/packages/validator/BUILD.bazel).

- **View Audit Artifact**:
  ```bash
  make report
  ```
  Displays the structured JSON verification report.

---

## License & Legal Notices

This project is licensed under the Apache License, Version 2.0. See the [LICENSE](file:///Users/rmcguinness/Projects/skill-builder/LICENSE) file for details. Attribution notices are maintained in the [NOTICE](file:///Users/rmcguinness/Projects/skill-builder/NOTICE) file.
