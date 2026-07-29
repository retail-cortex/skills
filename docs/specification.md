# Enterprise AI Agent Skills Specification (v1.0.0)

## 1. Overview & Scope

The **Enterprise AI Agent Skills Specification** defines the architectural contract, directory layout, metadata schema, resolution protocol, and cryptographic verification standard for AI Agent Skills compatible with Google Agent Development Kit (ADK) and polyglot agent runtimes.

This specification extends and supersedes baseline specifications (such as `agentskills.io`) by introducing:
1. **Cryptographic Lockfiles (`.manifest.lock`)**: Deterministic SHA-256 checksum tracking to prevent skill drift or unauthorized agent/developer tampering.
2. **5-Point SDLC Quality Invariants**: Mandatory frontmatter, progressive disclosure sub-trees, CWE security checkpoints, HTTP 429 resilience rules, and strict `file:///` link resolution.
3. **Polyglot URI Resolution**: Unified URI syntax supporting GitHub repositories (`github://`), Go modules (`mod://`), Java Maven artifacts (`maven://`), local packages (`pkg://`), and local filesystems (`file://`).
4. **Zero-I/O Pre-compiled Manifests (`skills_manifest.json`)**: In-memory skill registration for low-latency agent startup.

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

### 3.2 Frontmatter Example

```yaml
---
name: python-adk-fastapi
description: Enterprise Google ADK agent service integration wrapped in FastAPI.
license: Apache-2.0
author: Retail Cortex Engineering
authors:
  - name: Retail Cortex Core Team
    email: eng@retailcortex.com
    url: https://retailcortex.com
version: 1.0.0
compatibility: ADK-v2
category: python
tags:
  - fastapi
  - adk
  - python
  - REST
trigger_phrases:
  - "create ADK agent service"
  - "wrap ADK in FastAPI"
  - "build FastAPI agent endpoint"
allowed-tools: Bash(git:*) Read
tool_requirements:
  - name: Bash
    scopes: ["git:*", "uv:*"]
    description: Execute git and uv commands during project setup
  - name: Read
    scopes: ["*"]
    description: Inspect local source tree
execution_hints:
  preferred_model: "gemini-3.1-pro"
  requires_human_approval: false
  environment_variables:
    - GITHUB_TOKEN
  timeout_seconds: 300
---
```

---

## 4. 5-Point SDLC Compliance Standard

To pass enterprise audit and validation checks (`skm validate` or loader auditor suites), a skill MUST satisfy all 5 SDLC invariants:

1. **Frontmatter Validation**: `SKILL.md` contains valid YAML frontmatter specifying `name`, `description`, `license`, `author`, and `version`.
2. **Progressive Disclosure Tree**: The skill directory contains non-empty `references/` and `examples/` subdirectories with valid documentation files.
3. **CWE Security Checkpoints**: The body of `SKILL.md` or referenced files MUST contain explicit security invariants addressing relevant Common Weakness Enumerations (e.g., input validation for CWE-20/CWE-79, token auth for CWE-306, path traversal protection for CWE-22/CWE-59).
4. **HTTP 429 Resilience Guidelines**: The skill instructions MUST specify rate limiting, exponential backoff, or retry strategies for handling HTTP 429 (Too Many Requests) responses.
5. **Clickable Link Resolution**: All intra-repository document links within markdown files MUST use explicit `file:///` URIs to guarantee cross-editor clickability and boundary security.

---

## 5. Polyglot URI Resolution Protocol

Skills loaders and CLI clients MUST support resolution across five URI schemes:

| Scheme | URI Syntax Example | Resolution Strategy |
| :--- | :--- | :--- |
| **`github://`** | `github://owner/repo@ref/subpath` | Clones via git or downloads GitHub API zipballs, extracting `subpath`. |
| **`mod://` / `go://`** | `mod://module_path@version/subpath` | Resolves from `$GOPATH/pkg/mod` cache or runs `go mod download`. |
| **`maven://` / `mvn://`** | `maven://groupId:artifactId:version/subpath` | Resolves from `~/.m2/repository` or runs `mvn dependency:get`. |
| **`pkg://`** | `pkg://package_name` | Resolves from local workspace packages (`packages/skills-*`) or entry points. |
| **`file://`** | `file:///path/to/skill` | Resolves directly from local or relative filesystem paths. |

---

## 6. Cryptographic Integrity & Manifest Lockfile (`.manifest.lock`)

### 6.1 Purpose

When skills are downloaded or installed into a target destination directory (e.g. `.skills/`), the loader/installer MUST create or update a `.manifest.lock` file in that directory. The lockfile records the precise state of each installed skill.

### 6.2 Lockfile Schema

The `.manifest.lock` file is formatted as formatted JSON:

```json
{
  "version": "1.0.0",
  "skills": {
    "python-core": {
      "skill_name": "python-core",
      "uri": "github://retail-cortex/skills@main/packages/skills-python",
      "sha256": "4a7c8e9b01d2e3f4a5b6c7d8e9f01a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f"
    }
  }
}
```

### 6.3 Deterministic Checksum Algorithm

The SHA-256 checksum for a skill directory MUST be computed as follows:

1. Recursively discover all files inside `<dest_dir>/<skill_name>`.
2. Exclude `.DS_Store` and `.manifest.lock` files.
3. Construct a list of `(relative_path, file_content)` tuples, where `relative_path` is normalized using forward slashes (`/`).
4. Sort the list alphabetically by `relative_path`.
5. Initialize a SHA-256 digest context.
6. For each item in the sorted list:
   - Feed the UTF-8 bytes of `relative_path` into the SHA-256 digest context.
   - Feed the raw file content bytes into the SHA-256 digest context.
7. Output the lowercase hexadecimal representation of the digest.

### 6.4 Verification Lifecycle (`skm verify`)

Verification compares current skill directory contents against `.manifest.lock`:
- **`VERIFIED`**: Current calculated SHA-256 matches recorded `sha256`.
- **`MODIFIED`**: Skill directory exists but current SHA-256 differs from recorded `sha256` (files altered or tampered with).
- **`MISSING`**: Recorded skill directory does not exist in target path.

---

## 7. Zero-I/O Pre-compiled Manifest (`skills_manifest.json`)

To eliminate filesystem scanning during agent initialization, loaders support compiling all registered skills into a single `skills_manifest.json` file. Loaders read this pre-compiled JSON directly into memory for instant startup.

---

## 8. Multi-Language Loader Runtime API Contracts

Loader implementations in **Go** (`skillsloader`), **Python** (`loader`), and **Java** (`SkillLoader`) MUST expose equivalent programmatic interfaces for:

1. `LoadAllSkills(root, filter)` / `load_all_skills()` / `loadAllSkills()`
2. `CalculateSkillChecksum(skillDir)` / `calculate_skill_checksum()` / `calculateSkillChecksum()`
3. `ReadManifestLock(destDir)` / `read_manifest_lock()` / `readManifestLock()`
4. `WriteManifestLock(destDir, lockData)` / `write_manifest_lock()` / `writeManifestLock()`
5. `UpdateManifestLock(destDir, skillName, uri, checksum)` / `update_manifest_lock()` / `updateManifestLock()`
6. `VerifyManifestLock(destDir)` / `verify_manifest_lock()` / `verifyManifestLock()`

---

## 9. Protocol Buffer API Specifications (`api/v1/`)

All core data structures in this specification are defined in Protocol Buffer (proto3) schema files within [api/v1/](file:///Users/rmcguinness/Projects/skill-builder/api/v1/):

- [api/v1/skill.proto](file:///Users/rmcguinness/Projects/skill-builder/api/v1/skill.proto): Defines `SkillDefinition` and `SkillSummary`.
- [api/v1/manifest.proto](file:///Users/rmcguinness/Projects/skill-builder/api/v1/manifest.proto): Defines `ManifestLockEntry`, `ManifestLock`, `VerificationResult`, `VerificationReport`, and `VerificationStatus`.

Compiled bindings are maintained for:
- **Go**: `github.com/retail-cortex/skills/api/v1` ([api/v1/BUILD.bazel](file:///Users/rmcguinness/Projects/skill-builder/api/v1/BUILD.bazel))
- **Python**: `loader.api.v1` ([packages/loader/src/loader/api/v1/](file:///Users/rmcguinness/Projects/skill-builder/packages/loader/src/loader/api/v1/))
- **Java**: `com.retailcortex.skills.api.v1` ([clients/java/BUILD.bazel](file:///Users/rmcguinness/Projects/skill-builder/clients/java/BUILD.bazel))

