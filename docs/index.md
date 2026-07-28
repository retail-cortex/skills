# Project Overview: Agent Skill Builder

Welcome to the **Agent Skill Builder** documentation. This registry provides a comprehensive, standardized suite of 23 AI Agent Skills built strictly in compliance with the [agentskills.io](https://agentskills.io/specification) specification and Google Agent Development Kit (ADK) progressive disclosure architecture.

---

## Quickstart: Running Examples & Agents

Explore native Google Agent Development Kit (ADK) agent execution, qualified URI loading (`file://` and `github://...:branch`), and selective `.env` skill filtering.

### 1. Run Native ADK Example Script

Run the native ADK agent example demonstrating unified local workspace and remote GitHub skill loading:

```bash
PYTHONPATH=packages/skills-loader/src:packages/skills-agent/src:. uv run python -m examples.example_adk
# or direct script execution
uv run python examples/example-adk/main.py
```

### 2. Interactive ADK CLI Interface (`start-agent`)

Launch the interactive REPL terminal CLI to query and execute with loaded skills:

```bash
PYTHONPATH=packages/skills-loader/src:packages/skills-agent/src uv run start-agent
# or using Bazel
bazel run //:start-agent
```

Inside the `adk>` prompt:

```text
adk> skills
adk> search bigquery
adk> How do I build resilient BigQuery CAPI services with Gemini API skills?
```

### 3. Run All Workspace Test Suites

- **Hermetic Bazel Execution (Primary Standard)**:
  ```bash
  bazel test //...
  ```

- **Local Developer Execution (`pytest`)**:
  ```bash
  uv run pytest
  ```

- **Built-in Offline Fallback (`unittest`)**:
  ```bash
  PYTHONPATH=packages/skills-loader/src:packages/skills-agent/src:packages/validator/src uv run python -m unittest discover -s packages/skills-loader/tests && \
  PYTHONPATH=packages/skills-loader/src:packages/skills-agent/src:packages/validator/src uv run python -m unittest discover -s packages/skills-agent/tests && \
  PYTHONPATH=packages/skills-loader/src:packages/skills-agent/src:packages/validator/src uv run python -m unittest discover -s packages/validator/tests
  ```

---

## Documentation Structure

This documentation site is organized into logical sections:

- [Project Overview](index.md): Introduction, quickstart commands, repository layout, and licensing.
- [Architecture](architecture.md): Engineering standards, Google OAuth2 integration, HTTP 429 rate limit resilience, and multi-language defensive safety.
- [Skills Registry](skills.md): Scaffolding meta-skills and specialized domain/technology skills catalog.
- [Packages](packages.md): Core Python workspace packages (`skills-loader`, `skills-agent`, `validator`) and Bazel targets.
- [Examples](examples.md): Integration scripts, web servers, and setup guides.

---

## Workspace Directory Layout

The project is governed by a root [MODULE.bazel](https://github.com/retail-cortex/skills/blob/main/MODULE.bazel) (Bazel 9.2) and a root Python 3.13 `uv` workspace:

```
skill-builder/
├── MODULE.bazel               # Bazel 9.2 Bzlmod module definition
├── BUILD.bazel                # Root Bazel aliases and filegroups
├── .bazelignore               # Bazel directory exclusions
├── pyproject.toml             # Root uv workspace configuration
├── mkdocs.yml                 # MkDocs documentation site configuration
├── docs/                      # Documentation site source files
├── LICENSE                    # Apache 2.0 License
├── NOTICE                     # Legal attribution notices
├── validator_report.json      # Persisted 5-point SDLC audit results
├── packages/                  # Workspace utility packages
│   ├── skills-agent/          # Interactive CLI and ADK control plane
│   ├── skills-loader/         # Dynamic skill loader package
│   └── validator/             # 5-point SDLC validator package
└── skills/                    # Specialized AI Agent Skills
    ├── a2a/
    ├── a2ui/
    ├── ...
    └── terraform-gcp/
```

---

## License & Legal Notices

This project is licensed under the Apache License, Version 2.0. See [LICENSE](https://github.com/retail-cortex/skills/blob/main/LICENSE) for details. Attribution notices are maintained in [NOTICE](https://github.com/retail-cortex/skills/blob/main/NOTICE).
