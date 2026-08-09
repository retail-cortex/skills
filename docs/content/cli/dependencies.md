---
title: "Skill Dependencies & Discovery"
weight: 40
---

# Skill Dependencies & Discovery

The dependencies and discovery workflow provides capabilities to search, inspect, download, verify, and compile skills for agent consumption in local and cloud environments.

---

## 1. Remote Vector Search (`skm search`)

`skm search` performs semantic vector similarity search when querying a remote `skills-service` instance, or keyword/reference pattern matching when scanning locally.

```bash
# Multi-Modal vector search against central skills-service
skm search "canvas image rendering raster" --remote

# Paginated remote search with bounded page sizes (1 to 25)
skm search "database migration sql" --remote --page 1 --max 5

# Local filesystem search in .skills directory
skm search "oauth token" -d ./skills

# Machine-readable JSON output
skm search "kubernetes helm" --remote --json
```

### Remote Vector Search Output Example

```
SKM Remote Search Results for 'canvas image rendering raster' (1 matches on http://localhost:8000)
===========================================================================
Page 1 of 1 (Total matching skills: 1, Page size: 5)
---------------------------------------------------------------------------
- canvas-image (v1.0.0) [skm://skills/sk-0797300c]
  Description: Canvas image processing and pixel manipulation utilities for raster graphics
  Tags: canvas, image, rendering, raster, 2d
```

---

## 2. Skill Listing (`skm list`)

`skm list` displays all skills registered on a remote `skills-service` server or installed in a local directory.

```bash
# List remote server skills with pagination
skm list --remote --page 1 --max 5

# List local skills in target directory
skm list -d ./skills

# Structured JSON output
skm list --remote --json
```

### Bounded Pagination Enforcement

Remote requests to `/api/v1/skills` enforce strict abuse prevention and pagination bounds:
- **`--page` (`-p`)**: 1-based page number (minimum `1`).
- **`--max` / `--page-size` (`-n`)**: Maximum results per page (default: `5`, maximum: `25`).
- The server injects standard pagination response headers: `X-Total-Count`, `X-Page`, `X-Page-Size`, and `X-Total-Pages`.

---

## 3. Adding Dependencies (`skm add`)

Fetch and install skills into `.skills/` (or `-d <path>`), generating a `.manifest.lock` with cryptographic SHA-256 digests:

```bash
# Add registered enterprise skill from central server
skm add skm://skills/sk-9b1deb4d

# Add from GitHub repository (branch, tag, or commit ref)
skm add github://retail-cortex/skills@main/packages/skills-python

# Add from Go module cache
skm add mod://github.com/retail-cortex/skills@v1.0.0/packages/skills-go

# Add from Java Maven artifact
skm add maven://com.retailcortex.skills:skills-java:1.0.0

# Add from local filesystem directory with force overwrite
skm add file:///path/to/my-skill -d ./skills --force

# Update .manifest.lock without copying files
skm add github://retail-cortex/skills@main/packages/skills-devops --manifest-only
```

### Polyglot URI Resolution Schemes

| Scheme | Example | Description |
| :--- | :--- | :--- |
| **`skm://`** | `skm://skills/sk-9b1deb4d` | Resolves registered enterprise skills from central `skills-service`. |
| **`github://`** | `github://owner/repo[@ref][/path]` | Fetches git trees via `git clone` or GitHub API zipballs. |
| **`mod://`** | `mod://module_path[@version][/path]` | Resolves via `$GOPATH/pkg/mod` or `go mod download`. |
| **`maven://`** | `maven://groupId:artifactId:version` | Resolves from `~/.m2/repository` or `mvn dependency:get`. |
| **`pkg://`** | `pkg://package-name` | Resolves workspace packages within local Bazel runfiles. |
| **`file://`** | `file:///path/to/skill` | Resolves directly from local filesystem paths. |

---

## 4. Integrity Verification (`skm verify`)

Cryptographically audit installed skills against `.manifest.lock` to detect tampering or drift:

```bash
# Verify integrity of all installed skills
skm verify -d ./skills

# Generate structured JSON integrity report for CI/CD gates
skm verify -d ./skills --json
```

---

## 5. Manifest Compilation (`skm compile`)

Compile skills into a pre-compiled `skills_manifest.json` for zero-I/O cold starts in serverless or container runtimes:

```bash
# Compile skills directory into manifest
skm compile -d ./skills -o ./skills_manifest.json
```

