# Enterprise AI Agent Skills Specification (v1.0.0)

## 1. Overview & Scope

The **Enterprise AI Agent Skills Specification** defines the architectural contract, directory layout, metadata schema, resolution protocol, and cryptographic verification standard for AI Agent Skills compatible with Google Agent Development Kit (ADK) and polyglot agent runtimes.

This specification extends and supersedes baseline specifications (such as `agentskills.io`) by introducing:
1. **Canonical Central Registry & Server Registration (`skm://`)**: Registration protocol (`POST /api/v1/skills`) assigning unique `skill_id` keys and canonical URIs (`skm://skills/{skill_id}`).
2. **Native Build Lifecycle Integration**: Client build hooks (`skills-loader-maven-plugin` for Maven, `loader.build_meta` for Python PEP 517 / `uv`, `//go:generate` for Go) that audit skills and inject pre-compiled resources directly into application artifacts.
3. **Cryptographic Lockfiles (`.manifest.lock`)**: Deterministic SHA-256 checksum tracking to prevent skill drift or unauthorized agent/developer tampering.
4. **5-Point SDLC Quality Invariants**: Mandatory frontmatter, progressive disclosure sub-trees, CWE security checkpoints, HTTP 429 resilience rules, and strict `file:///` link resolution.
5. **Polyglot URI Resolution**: Unified URI syntax supporting central servers (`skm://`), GitHub repositories (`github://`), Go modules (`mod://`), Java Maven artifacts (`maven://`), local packages (`pkg://`), and local filesystems (`file://`).
6. **Zero-I/O Pre-compiled Manifests (`skills_manifest.json`)**: In-memory skill registration for low-latency agent startup.
7. **Just-in-Time (JIT) Semantic Discovery**: Semantic retrieval mapping user intent to specific skills, eliminating the need to statically load entire registries.
8. **Human-in-the-Loop (HITL) Intervention Gates**: Explicit compliance validation checkpoints designed to isolate read vs. write workloads for AHI safety.

---

## 2. Skill Directory Anatomy

A compliant skill MUST be structured as a self-contained directory containing a root `SKILL.md` file alongside supporting progressive disclosure subdirectories:

```
<skill-directory>/
├── SKILL.md                 # Primary skill specification & instructions (Required)
├── references/              # Detailed architectural guides & reference docs (Required)
│   └── *.md
├── examples/                # Code snippets, usage patterns & sample payloads (Required)
│   └── *
├── scripts/                 # Optional setup or execution scripts
└── resources/               # Optional static assets or schema definitions
```

---

## 3. SKILL.md Frontmatter & Content Specification

### 3.1 YAML Frontmatter Metadata Schema

Every `SKILL.md` MUST begin with a YAML frontmatter block enclosed by triple dashes (`---`). The frontmatter MUST contain the following fields:

| Field | Type | Required | Description |
| :--- | :--- | :--- | :--- |
| `name` | `string` | **Yes** | Kebab-case identifier matching the directory name (e.g. `python-adk-fastapi`). |
| `description` | `string` | **Yes** | Concise single-line summary of the skill's purpose and capability. |
| `license` | `string` | **Yes** | Standard SPDX license identifier (e.g. `Apache-2.0`). |
| `author` | `string` | **Yes** | Legacy single-string author name for backwards compatibility. |
| `authors` | `list[object]` | No | Structured list of contributors (`name`, `email`, `url`). |
| `version` | `string` | **Yes** | Semantic version string (e.g. `1.0.0`). |
| `compatibility` | `string` | No | Target runtime or framework compatibility requirement. |
| `allowed-tools` | `string` | No | Legacy flat tool permissions string (e.g. `Bash(git:*) Read`). |
| `tool_requirements` | `list[object]` | No | Strongly-typed tool permissions (`name`, `scopes`, `description`). |
| `category` | `string` | No | Primary domain classification (e.g. `python`, `devops`, `database`). |
| `tags` | `list[string]` | No | Keywords for registry discovery and indexing. |
| `trigger_phrases` | `list[string]` | No | Intent triggers for agent routers (e.g. `["scaffold Go microservice"]`). |
| `execution_hints` | `object` | No | Operational guidelines (`preferred_model`, `requires_human_approval`, `environment_variables`, `timeout_seconds`). |

Any unrecognized keys in the frontmatter MUST be preserved under a generic `metadata` dictionary.

---

## 4. Polyglot URI Resolution Protocol

A compliant skill loader MUST support resolving qualified skill root URIs:

| Scheme | Example | Resolution Mechanics |
| :--- | :--- | :--- |
| **`skm://`** | `skm://skills/sk-9b1deb4d` | Queries central `skills-service` HTTP endpoint (`GET /api/v1/skills/{id}`) using `X-API-Key`. |
| **`github://`** | `github://owner/repo[@ref][/path]` | Fetches git trees via `git clone` or GitHub zipballs. |
| **`mod://`** | `mod://module_path[@version][/path]` | Resolves via `$GOPATH/pkg/mod` or `go mod download`. |
| **`maven://`** | `maven://groupId:artifactId:version` | Resolves from `~/.m2/repository` or `mvn dependency:get`. |
| **`pkg://`** | `pkg://package-name` | Resolves workspace packages within local Bazel runfiles. |
| **`file://`** | `file:///path/to/skill` | Resolves directly from local filesystem paths. |

---

## 5. Enterprise Registration Protocol (`skm register`)

Central server registration publishes source skills and assigns immutable `skill_id` keys:

1. **Client Request**: `skm register <source_uri>` sends `POST /api/v1/skills` with `X-API-Key` authentication header.
2. **Server Processing**: `skills-service` validates source skill frontmatter, assigns `skill_id` (e.g., `sk-9b1deb4d`), and stores `source_uri` and canonical `uri` (`skm://skills/{skill_id}`).
3. **Client Response**: Returns HTTP `201 Created` with canonical URI `skm://skills/{skill_id}`.

---

## 6. Domain Scoping & Application Registration Standard (`urn:skm:app:...`)

Every application registered with `skills-service` MUST be bound to a verified domain authority and assigned a canonical RFC 8141 URN:

### 6.1 URN Syntax
- **Application URN**: `urn:skm:app:<domain>:<app_name>` (e.g. `urn:skm:app:retailcortex.com:checkout-agent`)
- **Skill URN**: `urn:skm:skill:<domain>:<app_name>:<skill_name>:<version>` (e.g. `urn:skm:skill:retailcortex.com:checkout-agent:payment-gateway:v1.0.0`)

### 6.2 Domain Ownership Validation Rules
1. **SSO Email Match (`VERIFIED_SSO`)**: If the developer's email domain matches the requested registration domain (e.g. `dev@retailcortex.com` registering `retailcortex.com`), domain verification is automatically confirmed.
2. **Freemail Prohibition (`ErrFreemailDomainProhibited`)**: Public freemail provider addresses (`gmail.com`, `yahoo.com`, `outlook.com`, `hotmail.com`, `icloud.com`, etc.) are explicitly forbidden from claiming corporate domain namespaces.
3. **DNS Challenge Fallback (`PENDING_DNS`)**: Claiming a custom third-party domain generates a DNS TXT challenge token (`skm-domain-verify-<uuid>`) that must be published under `_skm-challenge.<domain>` before activation.

---

## 7. Multi-Modal Vector Embedding & Poly-Column Indexing

Central registries MUST compute and maintain multi-modal semantic embeddings for registered skills:

1. **Poly-Column Dimensions**:
   - `embedding_768`: 768-dimensional float32 vector for standard text embedding models (`text-embedding-004`, `alloydb-ai`).
   - `embedding_1408`: 1408-dimensional float32 vector for multi-modal image and media models (`multimodalembedding`).
   - `embedding_3072`: 3072-dimensional float32 vector for high-dimensional models.
2. **HNSW Acceleration**: PostgreSQL/AlloyDB indexes MUST utilize HNSW (`m=16`, `ef_construction=64`) over `vector_cosine_ops`.
3. **Dual-Tier Resolution**: Skill metadata and individual progressive disclosure reference assets MUST be embedded independently to support precise RAG retrieval.

---

## 8. Bounded REST Pagination Protocol

All REST list and search endpoints (`/api/v1/skills`) MUST enforce strict request bounding:

1. **Parameter Constraints**:
   - `page`: Integer $\ge 1$ (default: `1`).
   - `page_size` / `max`: Integer $1 \le \text{page\_size} \le 25$ (default: `5`).
2. **Mandatory Response Headers**:
   - `X-Total-Count`: Total matched entities across the entire dataset.
   - `X-Page`: Current page index.
   - `X-Page-Size`: Effective bounded page size.
   - `X-Total-Pages`: Total calculated page count ($\lceil \text{Total} / \text{PageSize} \rceil$).
3. **Optional Envelope Parameter**: If `?envelope=true` is set, the server MUST wrap items in `{ "items": [...], "total_count": N, "page": P, "page_size": S, "total_pages": T }`.


