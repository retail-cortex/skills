# SKM (Skill Manager) CLI

`skm` is the enterprise package manager for AI Agent Skills. It allows you to instantly fetch, resolve, audit, and scaffold skills across local file systems, remote GitHub repositories, Go modules, and Java Maven artifacts.

---

## Quick Start

### 1. Installation

Download the latest pre-compiled standalone binary from GitHub Releases:

**macOS / Linux:**
```bash
# Download the binary (adjust OS/Arch as needed)
curl -sL "https://github.com/retail-cortex/skills/releases/latest/download/skm_$(uname -s)_$(uname -m)" -o skm

# Make executable and move to PATH
chmod +x skm
sudo mv skm /usr/local/bin/skm

skm version
```

*(To build from source using Bazel, see the Advanced section below).*

### 2. Common Workflow

Add a new skill from a remote GitHub repository to your local `.skills/` directory:
```bash
skm add github://retail-cortex/skills@main/packages/skills-python
```

Audit the newly added skill to ensure it meets the 5-point Enterprise SDLC security standard:
```bash
skm validate ./skills/skills-python
```

Verify the cryptographic integrity of your local skills against their `.manifest.lock` file to detect tampering:
```bash
skm verify
```

---

## Command Reference

### `skm add`
Downloads and installs skills from remote URIs or local directories into `.skills/` (or `-d <path>`). It automatically generates a `.manifest.lock` file with a deterministic SHA-256 digest to prevent prompt drift.

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

### `skm verify`
Verifies installed skills against recorded SHA-256 checksums in `.manifest.lock` to detect if files have been tampered with (`MODIFIED`), are perfectly intact (`VERIFIED`), or have been deleted (`MISSING`).

```bash
skm verify -d ./skills
skm verify -d ./skills --json  # Output as structured JSON
```

### `skm validate`
Runs the automated 5-point SDLC Quality Audit against a skill directory, checking for YAML frontmatter compliance, CWE security checkpoints, rate-limit resilience, and clickable file URIs.

```bash
skm validate ./skills/my-skill
skm validate -r ./packages --json # Recursively audit all skills and export JSON
```

### `skm list` & `skm search`
List or search currently registered skills.

```bash
skm list -d .skills
skm search python -d .skills
```

### `skm init`
Scaffolds a new, valid skill directory structure satisfying all SDLC invariants immediately.

```bash
skm init my-custom-skill -d ./skills
```

### `skm compile`
Compiles all loaded skills into a single zero-I/O `skills_manifest.json` file for low-latency serverless/container agent cold starts.

```bash
skm compile -d ./skills -o ./skills_manifest.json
```

---

## Supported Polyglot URIs

`skm add` supports advanced qualified URI resolution:

| Scheme | Example | Description |
| :--- | :--- | :--- |
| **`github://`** | `github://owner/repo[@ref][/path]` | Fetches git trees via `git clone` or GitHub API zipballs. |
| **`mod://`** | `mod://module_path[@version][/path]` | Resolves via `$GOPATH/pkg/mod` or `go mod download`. |
| **`maven://`** | `maven://groupId:artifactId:version` | Resolves from `~/.m2/repository` or `mvn dependency:get`. |
| **`pkg://`** | `pkg://package-name` | Resolves workspace packages within local Bazel runfiles. |
| **`file://`** | `file:///path/to/skill` | Resolves directly from local filesystem paths. |

---

## Advanced Setup & Integrations

### Shell Autocompletion
Generate autocompletion scripts for your shell:
```bash
eval "$(skm completion bash)" # Bash
eval "$(skm completion zsh)"  # Zsh
skm completion fish | source  # Fish
```

### Oh My Zsh Plugin Integration
An official plugin with aliases is provided in `cli/plugins/oh-my-zsh/skm`.
```bash
cp -r cli/plugins/oh-my-zsh/skm ~/.oh-my-zsh/custom/plugins/
# Then add 'skm' to your plugins=(...) array in ~/.zshrc and source ~/.zshrc
```
**Aliases**: `skma` (add), `skmver` (verify), `skmv` (validate), `skml` (list), `skms` (search), `skmc` (compile), `skmi` (init).

### Cross-Compilation with Bazel
Bazel natively cross-compiles static `skm` executables for Linux, macOS, and Windows without external dependencies:

```bash
# Build all platform binaries
bazel build //:skm-binaries

# Or target specific architectures
bazel build //cli:skm_linux_amd64
bazel build //cli:skm_darwin_arm64
bazel build //cli:skm_windows_amd64
```

---

## Language-Specific Integrations

While `skm` is a standalone Go binary, it seamlessly integrates with your favorite language ecosystems to fetch and resolve skills natively.

### Python (`pkg://` & `github://`)
When scaffolding Python agents (like ADK), `skm` can resolve skills directly from Python packages installed in your `uv` environment or remote git repositories.
```bash
skm add pkg://skills-python
skm add github://retail-cortex/skills@main/packages/skills-fastapi
```

### Go (`mod://`)
For Go microservices, `skm` natively hooks into the Go Module cache (`$GOPATH/pkg/mod`). If a skill isn't found locally, it will automatically trigger `go mod download` to fetch it.
```bash
skm add mod://github.com/retail-cortex/skills@v1.0.0/packages/skills-go
```

### Java (`maven://`)
For Java enterprise applications, `skm` resolves artifacts directly from your local `~/.m2/repository`. If missing, it shells out to `mvn dependency:get` to pull the skill JAR from your configured Maven registries.
```bash
skm add maven://com.retailcortex.skills:skills-java:1.0.0
```
