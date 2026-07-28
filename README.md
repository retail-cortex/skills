# Agent Skill Builder: Enterprise Project Pattern Registry

This registry provides a comprehensive, standardized suite of 23 AI Agent Skills built strictly in compliance with the [agentskills.io](https://agentskills.io/specification) specification and Google Agent Development Kit (ADK) progressive disclosure architecture.

---

## Documentation Site (MkDocs)

The full documentation is organized in the [docs](file:///Users/rmcguinness/Projects/skill-builder/docs) directory and powered by MkDocs Material.

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

- **[Bazel CI Workflow](file:///Users/rmcguinness/Projects/skill-builder/.github/workflows/bazel-ci.yml)**: Automated build, hermetic testing (`bazel test //...`), and SDLC validation on every `push` and `pull_request` using `bazelbuild/setup-bazelisk@v3` and `astral-sh/setup-uv@v5`.
- **[GitHub Pages Workflow](file:///Users/rmcguinness/Projects/skill-builder/.github/workflows/deploy-docs.yml)**: Automated MkDocs build (`uv run build-docs`) and deployment to GitHub Pages using `actions/upload-pages-artifact@v3` and `actions/deploy-pages@v4`.

---

## Documentation Sections

1. **[Project Overview](file:///Users/rmcguinness/Projects/skill-builder/docs/index.md)**: Introduction, features, quickstart guidelines, and workspace structure.
2. **[Architecture](file:///Users/rmcguinness/Projects/skill-builder/docs/architecture.md)**: Production engineering standards, Google OAuth2, HTTP 429 backoff resilience, defensive null safety, and Bazel 9.2 standards.
3. **[Skills Registry](file:///Users/rmcguinness/Projects/skill-builder/docs/skills.md)**: Scaffolding meta-skills (`mono-repo-setup`, `python-project-setup`, `go-project-setup`, `java-project-setup`) and 19 technology domain skills.
4. **[Packages & Tooling](file:///Users/rmcguinness/Projects/skill-builder/docs/packages.md)**: Utility packages (`skills-loader`, `skills-agent`, `validator`) and Bazel targets.
5. **[Examples & Demos](file:///Users/rmcguinness/Projects/skill-builder/docs/examples.md)**: Running native ADK agents and FastAPI web applications.

---

## Quickstart Commands

### 1. Run ADK Agent Example

```bash
PYTHONPATH=packages/skills-loader/src:packages/skills-agent/src:. uv run python -m examples.example_adk
```

### 2. Interactive CLI (`start-agent`)

```bash
PYTHONPATH=packages/skills-loader/src:packages/skills-agent/src uv run start-agent
# or using Bazel
bazel run //:start-agent
```

### 3. Run SDLC Skill Validator

```bash
# via uv
PYTHONPATH=packages/skills-loader/src:packages/skills-agent/src:packages/validator/src uv run python -m validator.main
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

This project is licensed under the Apache License, Version 2.0. See [LICENSE](file:///Users/rmcguinness/Projects/skill-builder/LICENSE) for details. Attribution notices are maintained in [NOTICE](file:///Users/rmcguinness/Projects/skill-builder/NOTICE).
