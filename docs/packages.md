# Packages & Tooling

The Agent Skill Builder repository includes three core Python utility packages in the `packages/` directory, integrated seamlessly into both the `uv` workspace and Bazel 9.2 build system.

---

## 1. `packages/skills-loader`

The `skills-loader` package provides dynamic loading, parsing, and filtering for AI agent skills across local file systems and remote Git repositories.

### Key Capabilities

- **Qualified URI Loading**: Supports loading skills from `file:///path/to/skills` and `github://owner/repo:branch`.
- **Environment Filtering**: Reads `.env` configuration to selectively enable or disable skills based on runtime requirements.
- **Progressive Disclosure**: Loads metadata frontmatter first and full instructions (`SKILL.md`) on demand.

### Python API Usage

```python
from skills_loader import SkillLoader

loader = SkillLoader()
skills = loader.load_from_uri("file://skills")
for skill in skills:
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

The `validator` package implements a 5-point SDLC audit scanner that inspects skills for compliance with agentskills.io specifications, security rules, and testing standards.

### Key Capabilities

- **5-Point Audit Standard**: Validates YAML frontmatter, directory structures, paired TDD coverage, security policies, and Bazel targets.
- **Report Generation**: Outputs and persists audit results to [validator_report.json](file:///Users/rmcguinness/Projects/skill-builder/validator_report.json).

### Execution

```bash
# Via uv
PYTHONPATH=packages/skills-loader/src:packages/skills-agent/src:packages/validator/src uv run python -m validator.main

# Via Bazel
bazel run //:validate
```

---

## Bazel 9.2 Build & Test Targets

Root Bazel convenience targets defined in [BUILD.bazel](file:///Users/rmcguinness/Projects/skill-builder/BUILD.bazel):

| Command | Bazel Target | Description |
| :--- | :--- | :--- |
| `bazel run //:start-agent` | `//packages/skills-agent:start_agent` | Launches interactive ADK programming agent CLI REPL. |
| `bazel run //:validate` | `//packages/validator:validate_skills` | Executes 5-point SDLC validator on all skills in `skills/`. |
| `bazel run //:docs` | `//packages/skills-loader:docs` | Launches local MkDocs development documentation server. |
| `bazel test //...` | `//...` | Runs all hermetic test targets across packages. |
