# SKM (Skill Manager) CLI Client

**`skm`** is a high-performance, standalone Go CLI client built with Bazel for fetching, resolving, auditing, verifying, and scaffolding AI Agent Skills in compliance with the [Enterprise AI Agent Skills Specification](specification.md) and Google Agent Development Kit (ADK) standards.

---

## 1. Overview & Key Capabilities

`skm` acts as the enterprise package manager for AI Agent Skills. Key capabilities include:

* **Polyglot URI Resolution**: Instantly pulls and resolves skills from remote GitHub repositories (`github://`), Go module caches (`mod://`), Java Maven artifacts (`maven://`), local packages (`pkg://`), and local filesystems (`file://`).
* **Cryptographic Integrity Locking**: Automatically generates and updates `.manifest.lock` with deterministic SHA-256 digests to prevent silent prompt drift or unauthorized tampering.
* **Automated 5-Point SDLC Quality Invariant Audit**: Validates YAML frontmatter, progressive disclosure sub-trees (`references/` & `examples/`), CWE security checkpoints, HTTP 429 rate-limiting guidelines, and clickable `file:///` URIs.
* **Zero-I/O Manifest Compilation**: Compiles registries into a single `skills_manifest.json` file for low-latency serverless or containerized agent cold starts.
* **Hermetic Bazel Build System**: Built with Go 1.22 and Bazel 9.2, cross-compiling static binaries for Linux, macOS, and Windows without external dependencies.

---

## 2. Polyglot URI Resolution Protocol

`skm` supports qualified URI schemes for fetching and installing skills:

| Scheme | Syntax Format | Description & Resolution Strategy |
| :--- | :--- | :--- |
| **`github://`** | `github://owner/repo[@ref][/subpath]` | Fetches remote git trees via `git clone` or GitHub API zipball fallback, extracting `subpath`. |
| **`mod://`** / **`go://`** | `mod://module_path[@version][/subpath]` | Resolves Go modules via local `$GOPATH/pkg/mod` or automated `go mod download`. |
| **`maven://`** / **`mvn://`** | `maven://groupId:artifactId:version[/subpath]` | Resolves Java Maven artifacts from `~/.m2/repository` or automated `mvn dependency:get`. |
| **`pkg://`** | `pkg://package-name` | Resolves enterprise skill packages within local workspaces and Bazel runfiles. |
| **`file://`** | `file:///path/to/skill` | Resolves skills directly from local filesystem paths. |

---

## 3. Installation & Cross-Platform Bazel Builds

### 3.1 Build Native Host Executable via Bazel

```bash
# Build skm CLI binary for host architecture
bazel build //:skm

# Binary location
./bazel-bin/cli/cmd/skm/skm_/skm version
```

### 3.2 Cross-Compile Executables for All Platforms

Bazel natively cross-compiles static `skm` executables for Linux, macOS, and Windows:

```bash
# Build all platform binaries
bazel build //:skm-binaries

# Build specific OS/Arch binaries
bazel build //cli:skm_linux_amd64    # Output: bazel-bin/cli/cmd/skm/skm_linux_amd64_/skm_linux_amd64
bazel build //cli:skm_darwin_arm64   # Output: bazel-bin/cli/cmd/skm/skm_darwin_arm64_/skm_darwin_arm64
bazel build //cli:skm_windows_amd64  # Output: bazel-bin/cli/cmd/skm/skm_windows_amd64_/skm_windows_amd64.exe
```

---

## 4. Command Reference

### 4.1 Add Skills (`skm add`)

Copy skills from remote URIs or local directories into `.skills/` (or `-d <path>`):

```bash
# Add from GitHub repository branch
skm add github://retail-cortex/skills@main/packages/skills-python

# Add from Go module cache or remote module download
skm add mod://github.com/retail-cortex/skills@v1.0.0/packages/skills-go

# Add from Java Maven artifact
skm add maven://com.retailcortex.skills:skills-java:1.0.0

# Add from local filesystem directory with force overwrite
skm add file:///path/to/my-skill -d ./skills --force
```

**Flags:**
- `-d, --dir <path>`: Target destination directory (default: `.skills`)
- `-f, --force`: Overwrite existing destination skill directories
- `--filter <names>`: Comma-separated list of skill names to select

> [!NOTE]
> `skm add` automatically generates and updates a `.manifest.lock` file in the destination directory containing the `skill_name`, source `uri`, and deterministic `sha256` checksum of the skill directory.

---

### 4.2 Verify Skill Integrity (`skm verify`)

Verify installed skills in `.skills/` (or `-d <path>`) against recorded SHA-256 checksums in `.manifest.lock`:

```bash
# Verify skills integrity in default .skills/ directory
skm verify

# Verify skills integrity in custom directory
skm verify -d ./skills

# Output verification report as structured JSON
skm verify -d ./skills --json
```

**Audit Status Outcomes:**
- `VERIFIED`: Calculated SHA-256 matches recorded checksum.
- `MODIFIED`: Directory exists but content has been altered or tampered with.
- `MISSING`: Recorded skill directory does not exist in target path.

---

### 4.3 Audit & Validate Skills (`skm validate`)

Audit skill directories against the 5-point SDLC compliance standard:

```bash
# Audit a single skill directory
skm validate ./skills/my-skill

# Recursively audit all skills in a directory tree
skm validate -r ./packages

# Export structured JSON audit report
skm validate -r ./packages --json
```

---

### 4.4 List Registered Skills (`skm list`)

List registered skills in `.skills/`, current directory, or a target path:

```bash
# List skills in default .skills directory
skm list

# List skills in a custom directory
skm list -d ./packages

# Output list in JSON format
skm list -d .skills --json
```

---

### 4.5 Search Skills (`skm search`)

Search loaded skills by keyword matching across name, description, and instructions:

```bash
skm search python -d .skills
```

---

### 4.6 Initialize New Skill (`skm init`)

Scaffold a new valid skill directory structure satisfying all 5 SDLC invariants:

```bash
skm init my-custom-skill -d ./skills
```

**Scaffolded Structure:**
- `my-custom-skill/SKILL.md` (with YAML frontmatter, CWE checkpoints, HTTP 429 resilience headers, clickable links)
- `my-custom-skill/references/guide.md`
- `my-custom-skill/examples/example.md`

---

### 4.7 Compile Pre-Compiled Manifest (`skm compile`)

Compile all loaded skills into a single zero-I/O `skills_manifest.json` file:

```bash
skm compile -d ./skills -o ./skills_manifest.json
```

---

### 4.8 Shell Autocompletion (`skm completion`)

Generate autocompletion scripts for Bash, Zsh, or Fish:

```bash
# Bash
eval "$(skm completion bash)"

# Zsh
eval "$(skm completion zsh)"

# Fish
skm completion fish | source
```

---

## 5. Oh My Zsh Plugin Integration

An official Oh My Zsh plugin is provided in [`cli/plugins/oh-my-zsh/skm`](https://github.com/retail-cortex/skills/tree/main/cli/plugins/oh-my-zsh/skm):

### Installation

```bash
# 1. Copy plugin to Oh My Zsh custom plugins directory
cp -r cli/plugins/oh-my-zsh/skm ~/.oh-my-zsh/custom/plugins/

# 2. Add 'skm' to plugins array in ~/.zshrc
# plugins=(git skm)

# 3. Reload zsh configuration
source ~/.zshrc
```

### Provided Aliases

| Alias | Full Command | Description |
| :--- | :--- | :--- |
| `skma` | `skm add` | Add and resolve skill URIs |
| `skmver` | `skm verify` | Check lockfile integrity |
| `skmv` | `skm validate` | Run 5-point SDLC audit |
| `skml` | `skm list` | Summarize registered skills |
| `skms` | `skm search` | Search skills by keyword |
| `skmc` | `skm compile` | Compile zero-I/O manifest |
| `skmi` | `skm init` | Scaffold new skill directory |

---

## 6. Development & Hermetic Testing

Run hermetic Bazel unit tests for the `skm` CLI router, installer, and validator packages:

```bash
# Run CLI package tests via Bazel
bazel test //cli/...

# Run via native Go toolchain
go test -v ./cli/...
```
