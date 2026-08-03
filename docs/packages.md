# Packages & Tooling

The Agent Skill Builder repository includes three core Python utility packages in the `packages/` directory, integrated seamlessly into both the `uv` workspace and Bazel 9.2 build system.

---

## 1. `packages/loader`

The `loader` package provides dynamic loading, parsing, and filtering for AI agent skills across local file systems and remote Git repositories.

### Key Capabilities

- **Qualified URI Loading**: Supports loading skills from `file:///path/to/skills` and `github://owner/repo:branch`.
- **Environment Filtering**: Reads `.env` configuration to selectively enable or disable skills based on runtime requirements.
- **Progressive Disclosure**: Loads metadata frontmatter first and full instructions (`SKILL.md`) on demand.

### Python API Usage

```python
from loader import SkillRegistry

registry = SkillRegistry()
skills = registry.skills
for name, skill in skills.items():
    print(f"Loaded: {skill.name} - {skill.description}")
```

---

## 2. `packages/skills-agent`

The `skills-agent` package houses the interactive terminal CLI REPL (`start-agent`) and the Google Agent Development Kit (ADK) control plane.

### Key Capabilities

- **Interactive CLI Interface**: Terminal prompt (`adk>`) for searching skills, inspecting skill metadata, and querying agents.
- **FastAPI Control Plane**: Wraps ADK execution runners inside FastAPI web services for streaming server-sent events (SSE).
- **OAuth Delegation**: Contextual user token injection into tool executions.

### Execution

```bash
# Via uv
PYTHONPATH=packages/skills-loader/src:packages/skills-agent/src uv run start-agent

# Via Bazel
bazel run //:start-agent
```

---

## 3. `packages/validator`

The `validator` package implements a 5-point SDLC audit scanner that inspects skills for compliance with the Enterprise AI Agent Skills Specification, security rules, and testing standards.

### Key Capabilities

- **5-Point Audit Standard**: Validates YAML frontmatter, directory structures, paired TDD coverage, security policies, and Bazel targets.
- **Report Generation**: Outputs and persists audit results to [validator_report.json]({{ config.repo_url }}/blob/main/validator_report.json).

### Execution

```bash
# Via uv
PYTHONPATH=packages/skills-loader/src:packages/skills-agent/src:packages/validator/src uv run python -m validator.main

# Via Bazel
bazel run //:validate
```

---

## 4. `apps/cli/cmd/skm` (SKM Skill Manager CLI)

`skm` is a standalone Go CLI tool built with Bazel for fetching, resolving, validating, and scaffolding AI Agent Skills across local and remote ecosystems.

### Polyglot URI Schemes Supported

- **`github://`**: Remote Git repositories (e.g. `github://retail-cortex/skills@main/packages/skills-python`)
- **`mod://` / `go://`**: Go modules via GOPATH cache or `go mod download` (e.g. `mod://github.com/retail-cortex/skills@v1.0.0/packages/skills-go`)
- **`maven://` / `mvn://`**: Java Maven artifacts via `~/.m2` or `mvn dependency:get` (e.g. `maven://com.retailcortex.skills:skills-java:1.0.0`)
- **`pkg://`**: Workspace packages (e.g. `pkg://skills-go`)
- **`file://`**: Local filesystem paths (e.g. `file:///path/to/my-skill`)

### Commands

```bash
# Add skills to .skills directory
bazel run //:skm -- add github://retail-cortex/skills@main/packages/skills-python
bazel run //:skm -- add mod://github.com/retail-cortex/skills@v1.0.0
bazel run //:skm -- add maven://com.retailcortex.skills:skills-java:1.0.0

# Pre-compile manifest for zero-I/O loading
bazel run //:skm -- compile -o ./skills_manifest.json

# Audit skill compliance
bazel run //:skm -- validate -r ./skills

# List and search skills
bazel run //:skm -- list -d .skills
bazel run //:skm -- search python -d .skills

# Scaffold new skill
bazel run //:skm -- init my-new-skill -d ./skills
```

---

## Bazel 9.2 Build & Test Targets

Root Bazel convenience targets defined in [BUILD.bazel]({{ config.repo_url }}/blob/main/BUILD.bazel):

| Command | Bazel Target | Description |
| :--- | :--- | :--- |
| `bazel run //:skm` | `//apps/cli/cmd/skm` | Executes standalone `skm` (*Skill Manager*) Go CLI binary on host system. |
| `bazel build //:skm-binaries` | `//apps/cli/cmd/skm:skm_binaries` | Native cross-compilation of `skm` executables for Windows (x64), Linux (x64/arm64), and macOS (x64/arm64). |
| `bazel run //:start-agent` | `//packages/skills-agent:start_agent` | Launches interactive ADK programming agent CLI REPL. |
| `bazel run //:validate` | `//packages/validator:validate_skills` | Executes 5-point SDLC validator on all skills in `skills/`. |
| `bazel run //:docs` | `//packages/skills-loader:docs` | Launches local MkDocs development documentation server. |
| `bazel test //...` | `//...` | Runs all hermetic test targets across Go, Java, Python, and CLI packages. |

