---
title: "CLI Overview"
weight: 10
---

# Castor CLI (`cstr`)

`cstr` is the enterprise CLI client and package manager for AI Agent Skills. It manages the complete lifecycle of skills—from local scaffolding and SDLC validation to enterprise server registration, multi-modal vector search, and multi-language dependency resolution (`castor://`, `cstr://`, `github://`, `mod://`, `maven://`, `pkg://`, `file://`).

---

## Architecture Overview

```mermaid
graph TD
    A[cstr CLI Client] --> B[1. App Registration & Config]
    A --> C[2. Skill Registration]
    A --> D[3. Remote Search & Listing]
    A --> E[4. Skill Dependencies & Lockfiles]
    A --> F[5. Scaffolding & Quality Audit]

    B --> B1["cstr login <email> / cstr config set"]
    C --> C1["cstr register <source_uri>"]
    D --> D1["cstr search / cstr list (--remote, --page, --max)"]
    E --> E1["cstr add / cstr verify / cstr compile"]
    F --> F1["cstr validate / cstr init"]
```

---

## Command Reference

| Command | Usage | Description |
| :--- | :--- | :--- |
| **`search`** | `cstr search <query> [-r] [-p <page>] [-n <max>] [--json]` | Searches skills locally or queries remote `Castor Registry` vector index. |
| **`list`** | `cstr list [-r] [-p <page>] [-n <max>] [--json]` | Lists skills from local filesystem or remote server with bounded pagination. |
| **`add`** | `cstr add <uri> [-d <dir>] [--force] [--manifest-only]` | Resolves and installs skills from polyglot URIs, generating `.manifest.lock`. |
| **`register`**| `cstr register <source_uri>` | Registers source skill to central server, computing multi-modal embeddings. |
| **`login`** | `cstr login <email> [app_name]` | Initiates developer app registration and email verification challenge. |
| **`config`** | `cstr config set <key> <value>` | Sets connection configuration in `~/.castor/.env.toml`. |
| **`config`** | `cstr config show` | Displays active CLI configuration and masked credentials. |
| **`validate`**| `cstr validate <path> [-r] [--json]` | Runs 5-point SDLC quality audit (Frontmatter, Structure, CWE, 429, Links). |
| **`verify`** | `cstr verify [-d <dir>] [--json]` | Verifies cryptographic SHA-256 integrity against `.manifest.lock`. |
| **`compile`** | `cstr compile [-d <dir>] [-o <file>]` | Compiles skills into a zero-I/O `skills_manifest.json` bundle. |
| **`init`** | `cstr init <name> [-d <dir>]` | Scaffolds a new skill directory conforming to SDLC specifications. |

---

## Global & Common Flags

* **`-r, --remote`**: Directs `list` or `search` operations to query the central `Castor Registry` server (`CASTOR_SERVER_URL`).
* **`-p, --page <num>`**: 1-based page index for remote paginated results (default: `1`).
* **`-n, --max, --page-size <num>`**: Number of items per page. Clamped between `1` and `25` (default: `5`).
* **`-d, --dir <path>`**: Target directory for skill operations (default: `./.skills` or workspace root).
* **`-s, --server <url>`**: Override target server URL for remote requests (default: `http://localhost:8000`).
* **`--json`**: Formats command output as structured JSON.
* **`--force`**: Overwrites existing destination skills during `cstr add`.
* **`--manifest-only`**: Updates `.manifest.lock` without copying files to disk.

---

## Installation & Cross-Compilation

### Pre-Compiled Standalone Binary

```bash
# Download binary for your platform (macOS / Linux / Windows)
curl -sL "https://github.com/retail-cortex/castor/releases/latest/download/cstr_$(uname -s)_$(uname -m)" -o cstr
chmod +x cstr
sudo mv cstr /usr/local/bin/cstr

cstr version
```

### Hermetic Bazel Cross-Compilation

```bash
# Build native host binary
bazel build //cmd/cstr

# Cross-compile for all enterprise target architectures
bazel build //cmd/cstr:cstr_linux_amd64
bazel build //cmd/cstr:cstr_darwin_arm64
bazel build //cmd/cstr:cstr_windows_amd64
```

