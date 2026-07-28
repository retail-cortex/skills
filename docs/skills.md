# Skills Registry

This registry provides 23 enterprise-ready AI Agent Skills built strictly in compliance with [agentskills.io](https://agentskills.io/specification) and Google ADK progressive disclosure rules.

---

## Scaffolding & Setup Meta-Skills

| Meta-Skill | Technology & Build Stack | Description |
| :--- | :--- | :--- |
| [mono-repo-setup]({{ config.repo_url }}/blob/main/skills/mono-repo-setup/SKILL.md) | Bazel 8/9 Bzlmod Polyglot | Scaffolding for polyglot monorepos (Go 1.26+, Python 3.13 uv, Java 25 Maven, TypeScript React 19 pnpm), root Bazel configs, and `configs/terraform/`. |
| [python-project-setup]({{ config.repo_url }}/blob/main/skills/python-project-setup/SKILL.md) | Python 3.13, uv & Bazel | Scaffolding for enterprise Python services using `uv`, wrapped in Bazel `rules_python`, with Google OAuth2, None safety, and cascading TOML. |
| [go-project-setup]({{ config.repo_url }}/blob/main/skills/go-project-setup/SKILL.md) | Go 1.26+ & Bazel | Scaffolding for Go microservices using standard `/cmd`, `/internal`, `/pkg`, `/api` layout, Bazel `rules_go`/`gazelle`, nil safety, and `configs/terraform/`. |
| [java-project-setup]({{ config.repo_url }}/blob/main/skills/java-project-setup/SKILL.md) | Java 25 (LTS), Maven & Bazel | Scaffolding for Java 25 microservices using Maven POM with latest GCP BOM, wrapped in Bazel `rules_java`, with Optional/NonNull safety and `configs/terraform/`. |

---

## Domain & Technology Skills

| Skill Directory | Technology / Domain | SDLC, Google OAuth2, Error Handling & 429 Highlights |
| :--- | :--- | :--- |
| [bazel-modules]({{ config.repo_url }}/blob/main/skills/bazel-modules/SKILL.md) | Bazel 8/9 & Bzlmod | Bzlmod dependency graphs, hermetic TDD, `rules_hugo` GitHub Pages publishing, coverage collection, and SemVer. |
| [react-vite]({{ config.repo_url }}/blob/main/skills/react-vite/SKILL.md) | React 19 & Vite 6 | Google OAuth2 PKCE/GIS, HTTP 429 countdown banners, `strictNullChecks`, optional chaining, paired Vitest TDD, 85% coverage. |
| [python-core]({{ config.repo_url }}/blob/main/skills/python-core/SKILL.md) | Python 3.13 & UV | None safety, paired positive/negative pytest TDD, 90% coverage enforcement, MkDocs on Pages, and SemVer. |
| [python-fastapi]({{ config.repo_url }}/blob/main/skills/python-fastapi/SKILL.md) | FastAPI Enterprise | Google OAuth2 Bearer token verification, slowapi 429 rate limiting, async REST TDD with httpx, 90% coverage, and SemVer. |
| [python-fastmcp]({{ config.repo_url }}/blob/main/skills/python-fastmcp/SKILL.md) | FastMCP Protocol | FastMCP tool TDD, GitHub Actions schema validation, prompt injection defense, Pydantic bounds, and SemVer. |
| [python-adk-fastapi]({{ config.repo_url }}/blob/main/skills/python-adk-fastapi/SKILL.md) | FastAPI & Google ADK | Superior pattern wrapping ADK in FastAPI on Uvicorn, Google OAuth2 user token delegation to CAPI, tenacity 429 backoff with jitter. |
| [go-lang]({{ config.repo_url }}/blob/main/skills/go-lang/SKILL.md) | Go 1.26+ Services | Google OAuth2 ID token verification middleware, retryablehttp 429 backoff, nil pointer safety, paired table-driven TDD. |
| [alloydb]({{ config.repo_url }}/blob/main/skills/alloydb/SKILL.md) | AlloyDB & pgvector | Testcontainers TDD, migration CI/CD validation, private IP VPC security, SSL enforcement, and SemVer schema versioning. |
| [bigquery]({{ config.repo_url }}/blob/main/skills/bigquery/SKILL.md) | BigQuery & CAPI | Dry-run query TDD, SQLFluff linting, GE OAuth security delegation, Protobuf normalization, and SemVer. |
| [opentelemetry-google]({{ config.repo_url }}/blob/main/skills/opentelemetry-google/SKILL.md) | OpenTelemetry & GCP Trace | InMemorySpanExporter TDD, CI/CD pipeline verification, PII scrubbing security, batch flush lifecycles, and SemVer. |
| [terraform-gcp]({{ config.repo_url }}/blob/main/skills/terraform-gcp/SKILL.md) | Terraform for GCP | `terraform test` TDD, GitHub Actions CI/CD with tflint, GCS remote state security, least-privilege IAM, and SemVer. |
| [configuration-modenv]({{ config.repo_url }}/blob/main/skills/configuration-modenv/SKILL.md) | Configuration & modenv | Cascading TOML precedence TDD, GitHub Actions CI schema validation, XOR cipher memory decryption, and SemVer. |
| [java-enterprise]({{ config.repo_url }}/blob/main/skills/java-enterprise/SKILL.md) | Java 25 (LTS) & Javalin | Google OAuth2 ID token verifier, Resilience4j 429 retries, Optional/NonNull null safety, JUnit 5 TDD, JaCoCo 85% coverage. |
| [a2a]({{ config.repo_url }}/blob/main/skills/a2a/SKILL.md) | A2A Protocol & Orchestration | Component TDD, GitHub Actions CI validation, token bucket rate limiting, HMAC message auth (CWE-306), and SemVer. |
| [a2ui]({{ config.repo_url }}/blob/main/skills/a2ui/SKILL.md) | A2UI v0.8 & Gemini Enterprise | Sandboxed iframe CSP policies (CWE-1021), postMessage origin checks (CWE-346), right-aligned actions with `justify-end`, and component TDD. |
| [docker-containers]({{ config.repo_url }}/blob/main/skills/docker-containers/SKILL.md) | Multi-Stage Docker | Container health check TDD, GitHub Actions Trivy vulnerability scanning, non-root user security, and immutable SemVer tags. |
| [nx-monorepo]({{ config.repo_url }}/blob/main/skills/nx-monorepo/SKILL.md) | NX Polyglot Monorepo | Affected TDD testing, GitHub Actions CI with remote caching, module boundary security rules, and SemVer release orchestration. |
| [protobuf-grpc]({{ config.repo_url }}/blob/main/skills/protobuf-grpc/SKILL.md) | Protobuf & gRPC | In-process gRPC TDD, Buf breaking change detection, mTLS/auth security, Bazel rules_proto_grpc, and SemVer. |
| [adk-skill-factory]({{ config.repo_url }}/blob/main/skills/adk-skill-factory/SKILL.md) | ADK Meta-Skill Factory | ADK evaluator TDD, GitHub Actions schema validation, human-in-the-loop security reviews, progressive disclosure, and SemVer. |

---

## Skill Anatomical Blueprint

Each skill directory contains:

- `SKILL.md`: Root specification document with YAML frontmatter metadata, trigger conditions, core rules, and implementation patterns.
- `scripts/`: Helper utilities and setup scripts.
- `examples/`: Reference implementations and code snippets.
- `resources/` & `references/`: Supporting documentation, schema definitions, and templates.
