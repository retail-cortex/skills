# SKM (Skill Manager) CLI Client

**`skm`** is a high-performance, standalone Go CLI client built with Bazel for fetching, resolving, auditing, and scaffolding AI Agent Skills in compliance with the [agentskills.io](https://agentskills.io/specification) specification and Google Agent Development Kit (ADK) standards.

---

## Supported URI Schemes

`skm` supports polyglot URI resolution for pulling local and remote skills into your project's `.skills/` directory (or a custom `-d <path>`):

| Scheme | Format | Description |
| :--- | :--- | :--- |
| **`github://`** | `github://owner/repo[@ref][/subpath]` | Fetches skills from remote GitHub repositories or git branches. |
| **`mod://`** / **`go://`** | `mod://module_path[@version][/subpath]` | Resolves Go modules via local `$GOPATH/pkg/mod` or automated `go mod download`. |
| **`maven://`** / **`mvn://`** | `maven://groupId:artifactId:version[/subpath]` | Resolves Java Maven artifacts from `~/.m2/repository` or automated `mvn dependency:get`. |
| **`pkg://`** | `pkg://package-name` | Resolves enterprise skill packages within local workspaces and Bazel runfiles. |
| **`file://`** | `file:///path/to/skill` | Resolves skills directly from local filesystem paths. |

---

## Installation & Build

### 1. Build Native Cross-Platform Executables via Bazel (Recommended)

Bazel natively cross-compiles static `skm` executables for Linux, macOS, and Windows targets without requiring external Makefiles or toolchain switches:

```bash
# Build native host binary
bazel build //:skm

# Build all cross-platform executables (Windows x64, Linux x64/arm64, macOS x64/arm64)
bazel build //:skm-binaries

# Build specific OS/Arch targets
bazel build //cli:skm_windows_amd64  # Output: bazel-bin/cli/cmd/skm/skm_windows_amd64_/skm_windows_amd64.exe
bazel build //cli:skm_linux_amd64    # Output: bazel-bin/cli/cmd/skm/skm_linux_amd64_/skm_linux_amd64
bazel build //cli:skm_darwin_arm64   # Output: bazel-bin/cli/cmd/skm/skm_darwin_arm64_/skm_darwin_arm64

# Run directly via Bazel
bazel run //:skm -- version
```

---

## Command Reference

### 1. Add Skills (`skm add`)

Copy skills from remote URIs or local directories into `.skills/` (or `-d <path>`):

```bash
# Add from GitHub repository
./bin/skm add github://retail-cortex/skills@main/packages/skills-python

# Add from Go module cache or remote module download
./bin/skm add mod://github.com/retail-cortex/skills@v1.0.0/packages/skills-go

# Add from Java Maven artifact
./bin/skm add maven://com.retailcortex.skills:skills-java:1.0.0

# Add from local filesystem directory with force overwrite
./bin/skm add file:///path/to/my-skill -d ./skills --force
```

**Options for `add`:**
- `-d, --dir <path>`: Target destination directory (default: `.skills`)
- `-f, --force`: Overwrite existing destination skill directories
- `--filter <names>`: Comma-separated list of skill names to select

---

### 2. Audit & Validate Skills (`skm validate`)

Audit skill directories against the 5-point compliance standard (YAML frontmatter, `references/` & `examples/` L3 trees, CWE security checkpoints, HTTP 429 rate limit guidelines, and markdown `file:///` links):

```bash
# Audit a single skill directory
./bin/skm validate ./skills/my-skill

# Recursively audit all skills in a directory tree
./bin/skm validate -r ./packages

# Export structured JSON audit report
./bin/skm validate -r ./packages --json
```

**Options for `validate`:**
- `-r, --recursive`: Recursively audit skill subdirectories
- `--json`: Output audit summary as structured JSON

---

### 3. List Skills (`skm list`)

List registered skills in `.skills/`, current directory, or a target path:

```bash
# List skills in default .skills directory
./bin/skm list

# List skills in a custom directory
./bin/skm list -d ./packages

# Output list in JSON format
./bin/skm list -d .skills --json
```

---

### 4. Search Skills (`skm search`)

Search loaded skills by query term across name, description, and instructions:

```bash
./bin/skm search python -d .skills
```

---

### 5. Initialize New Skill (`skm init`)

Scaffold a new valid skill directory structure:

```bash
./bin/skm init my-custom-skill -d ./skills
```

Creates:
- `my-custom-skill/SKILL.md` (with YAML frontmatter, CWE checkpoints, 429 resilience headers, clickable links)
- `my-custom-skill/references/guide.md`
- `my-custom-skill/examples/example.md`

---

### 6. Shell Autocompletion & Oh My Zsh Plugin (`skm completion`)

Generate autocompletion scripts for Bash, Zsh, or Fish:

```bash
# Bash
eval "$(skm completion bash)"

# Zsh
eval "$(skm completion zsh)"

# Fish
skm completion fish | source
```

#### Oh My Zsh Plugin Installation

An official Oh My Zsh plugin is provided in [`plugins/oh-my-zsh/skm`](plugins/oh-my-zsh/skm/):

```bash
# 1. Install plugin to Oh My Zsh custom directory
cp -r cli/plugins/oh-my-zsh/skm ~/.oh-my-zsh/custom/plugins/

# 2. Add 'skm' to plugins array in ~/.zshrc
# plugins=(git skm)

# 3. Reload configuration
source ~/.zshrc
```

##### Included Aliases:
- `skma` -> `skm add`
- `skmv` -> `skm validate`
- `skml` -> `skm list`
- `skms` -> `skm search`
- `skmc` -> `skm compile`
- `skmi` -> `skm init`

---

## Testing

Run unit test suites for `skm` CLI, installer, and validator packages via Bazel or Go toolchain:

```bash
# Hermetic Bazel tests
bazel test //cli/...

# Go toolchain tests
go test -v ./cli/...
```
