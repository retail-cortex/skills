# Project Overview: Agent Skill Builder

Welcome to the **Agent Skill Builder** documentation. This registry provides a comprehensive, standardized suite of 33 AI Agent Skills built strictly in compliance with the [agentskills.io](https://agentskills.io/specification) specification and Google Agent Development Kit (ADK) progressive disclosure architecture.

---

## Quickstart: Running Standalone Examples & Agents

Explore native Google Agent Development Kit (ADK) agent execution, qualified URI loading (`file://` and `github://...:branch`), and selective `.env` skill filtering.

### 1. Run Native ADK Example Package (`examples/example-adk`)

Run the native ADK agent example demonstrating unified local workspace and remote GitHub skill loading:

```bash
uv run python examples/example-adk/main.py
```

### 2. Run Polyglot Developer Agent (`examples/polyglot-developer`)

Run the custom polyglot developer CLI agent using domain skills (`skills-bazel`, `skills-go`, `skills-java`, `skills-protobuf`, `skills-python`, `skills-frontend`) to scaffold a Bazel monorepo:

```bash
uv run python examples/polyglot-developer/main.py --target-dir ./scratch/my-polyglot-app
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

- **Standard `unittest` Suite**:
  ```bash
  uv run python -m unittest packages/skills-bazel/tests/test_skills_bazel.py packages/skills-a2a/tests/test_skills_a2a.py packages/skills-a2ui/tests/test_skills_a2ui.py packages/skills-loader/tests/test_loader.py packages/validator/tests/test_validator.py examples/polyglot-developer/tests/test_polyglot_developer.py
  ```

---

## Documentation Structure

This documentation site is organized into logical sections:

- [Project Overview](index.md): Introduction, quickstart commands, repository layout, and licensing.
- [Specification](specification.md): Enterprise AI Agent Skills Specification (v1.0.0), frontmatter schema, 5-point SDLC compliance, `.manifest.lock` cryptographic integrity, and polyglot URI resolution.
- [Architecture](architecture.md): Engineering standards, Google OAuth2 integration, HTTP 429 rate limit resilience, and multi-language defensive safety.
- [Critical Analysis](analysis.md): Comparative analysis against agentskills.io specification and ecosystem showcase clients.
- [Skills Registry](skills.md): Specialized domain and technology skills catalog.
- [CLI Client (skm)](cli.md): Standalone `skm` Go CLI client manual, cross-platform builds, polyglot URI resolution, subcommands, and Oh My Zsh plugin.
- [Packages](packages.md): Modular Python workspace packages (`skills-loader`, `skills-a2a`, `skills-a2ui`, `skills-bazel`, `skills-python`, `skills-go`, `skills-java`, `skills-protobuf`, `skills-frontend`, `skills-devops`, `skills-database`, `skills-google-adk-skill-builder`, `validator`) and Bazel targets.
- [Examples](examples.md): Standalone integration packages, web servers, and setup guides.


---

## Workspace Directory Layout

The project is governed by a root [MODULE.bazel]({{ config.repo_url }}/blob/main/MODULE.bazel) (Bazel 9.2) and a root Python 3.13 `uv` workspace:

```text
skill-builder/
├── MODULE.bazel               # Bazel 9.2 Bzlmod module definition
├── BUILD.bazel                # Root Bazel aliases and filegroups
├── Makefile                   # Root automation Makefile
├── pyproject.toml             # Root uv workspace configuration
├── mkdocs.yml                 # MkDocs documentation site configuration
├── docs/                      # Documentation site source files
├── LICENSE                    # Apache 2.0 License
├── NOTICE                     # Legal attribution notices
├── validator_report.json      # Persisted 5-point SDLC audit results
├── cli/                       # SKM (Skill Manager) Go CLI package
│   ├── cmd/skm/               # Minimal main.go entry point
│   ├── internal/installer/    # Polyglot URI skill resolver (github, mod, maven, pkg, file)
│   ├── internal/validator/    # 5-point skill audit & validation engine
│   └── internal/commands/     # Subcommand execution router
├── skills/                    # Root skills directory categorized by domain
│   ├── a2a/                   # Agent-to-Agent protocol skills
│   ├── a2ui/                  # Agent-to-User Interface skills
│   ├── bazel/                 # Hermetic Bazel 9.2 module skills
│   ├── database/              # BigQuery and AlloyDB skills
│   ├── devops/                # Docker, Terraform GCP, NX, and OTel skills
│   ├── frontend/              # React and Vite frontend skills
│   ├── go/                    # Go microservice & project setup skills
│   ├── google-adk-skill-builder/ # ADK skill generator meta-skill
│   ├── java/                  # Java enterprise & Maven skills
│   ├── protobuf/              # Protocol Buffers & gRPC contract skills
│   └── python/                # Python core, FastAPI, ADK & MCP skills
├── examples/                  # Standalone example packages
│   ├── example-adk/           # Native ADK agent example package
│   └── polyglot-developer/    # Polyglot Bazel monorepo developer agent CLI package
└── packages/                  # SDK packages & language test packages
    ├── loader/                # Dynamic skill loader Python package
    ├── validator/             # 5-point SDLC validator Python package
    ├── skills-go/             # Go language test skill package
    ├── skills-java/           # Java language test skill package
    └── skills-python/         # Python language test skill package
```

---

## License & Legal Notices

This project is licensed under the Apache License, Version 2.0. See [LICENSE]({{ config.repo_url }}/blob/main/LICENSE) for details. Attribution notices are maintained in [NOTICE]({{ config.repo_url }}/blob/main/NOTICE).
