# Agent Skill Builder: Enterprise Project Pattern Registry

This registry provides a comprehensive, standardized suite of 23 AI Agent Skills built strictly in compliance with the [agentskills.io](https://agentskills.io/specification) specification and Google Agent Development Kit (ADK) progressive disclosure architecture.

---

## Documentation Site (MkDocs)

The full documentation site is published live at **[https://retail-cortex.github.io/skills/](https://retail-cortex.github.io/skills/)**.
Source files are organized in the [docs](docs/) directory and powered by MkDocs Material.

### Serving Documentation Locally

```bash
# via uv executable script
uv run serve-docs
# or via Bazel target
bazel run //:docs

# or direct mkdocs CLI
uv run mkdocs serve
```

Access the local documentation server at `http://127.0.0.1:8000`.

### Build Static Site

```bash
uv run build-docs
# or via mkdocs CLI
uv run mkdocs build
```

---

## CI/CD & GitHub Pages Deployment

The repository uses official GitHub Actions workflows for continuous integration and automated GitHub Pages deployment:

- **[Bazel CI Workflow](https://github.com/retail-cortex/skills/blob/main/.github/workflows/bazel-ci.yml)**: Automated build, hermetic testing (`bazel test //...`), and SDLC validation on every `push` and `pull_request` using Node 24 native actions (`bazelbuild/setup-bazelisk@v3`, `astral-sh/setup-uv@v7`, `actions/checkout@v7`, `actions/setup-python@v7`).
- **[GitHub Pages Workflow](https://github.com/retail-cortex/skills/blob/main/.github/workflows/deploy-docs.yml)**: Automated MkDocs build (`uv run build-docs`) and deployment to GitHub Pages using Node 24 native actions (`actions/upload-pages-artifact@v5`, `actions/deploy-pages@v5`, `actions/configure-pages@v6`).

---

## Documentation Sections

1. **[Project Overview](https://retail-cortex.github.io/skills/)**: Introduction, features, quickstart guidelines, and workspace structure.
2. **[Architecture](https://retail-cortex.github.io/skills/architecture/)**: Production engineering standards, Google OAuth2, HTTP 429 backoff resilience, defensive null safety, and Bazel 9.2 standards.
3. **[Skills Registry](https://retail-cortex.github.io/skills/skills/)**: Specialized domain and technology skills catalog.
4. **[Packages & Tooling](https://retail-cortex.github.io/skills/packages/)**: Modular domain packages (`skills-loader`, `skills-a2a`, `skills-a2ui`, `skills-bazel`, `skills-python`, `skills-go`, `skills-java`, `skills-protobuf`, `skills-frontend`, `skills-devops`, `skills-database`, `skills-google-adk-skill-builder`, `validator`) and Bazel targets.
5. **[Examples & Demos](https://retail-cortex.github.io/skills/examples/)**: Standalone integration packages (`example-adk` and `polyglot-developer`).

---

## Quickstart Commands

### 1. Run Native ADK Agent Example

```bash
uv run python examples/example-adk/main.py
```

### 2. Run Polyglot Developer Agent CLI

```bash
uv run python examples/polyglot-developer/main.py --target-dir ./scratch/my-polyglot-app
```

### 3. Run SDLC 5-Point Skill Validator

```bash
# via uv
uv run python -m validator.cli audit
# or via Bazel
bazel run //:validate
```

### 4. Run Test Suite

```bash
bazel test //...
# or via pytest
uv run pytest
```

---

## License & Legal Notices

This project is licensed under the Apache License, Version 2.0. See [LICENSE](https://github.com/retail-cortex/skills/blob/main/LICENSE) for details. Attribution notices are maintained in [NOTICE](https://github.com/retail-cortex/skills/blob/main/NOTICE).
