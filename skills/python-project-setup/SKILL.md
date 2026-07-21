---
name: python-project-setup
description: Elite meta-skill for scaffolding enterprise Python 3.13 projects using uv package management wrapped in Bazel rules_python. Enforces src/ layout, 90% TDD coverage, HTTP 429 rate limit backoff, and GCP Terraform configs.
---

# Python Project Setup Meta-Skill (uv & Bazel Standard)

This meta-skill provides automated scaffolding instructions and templates for initializing enterprise **Python 3.13** projects powered by **uv**, wrapped in **Bazel**, and configured with dedicated infrastructure manifests in `configs/`.

## Prerequisites & Pre-Flight Checklist

1. Python 3.13 runtime installed locally.
2. `uv` binary available on system PATH.
3. Google Cloud project created with Vertex AI and BigQuery APIs enabled.

## HTTP 429 Rate Limit & Quota Resilience Invariants

- AI and cloud service calls must wrap outbound requests in `tenacity` exponential backoff with randomized jitter to handle HTTP 429 rate limits.

## Security Checkpoints & CWE Invariants

- **CWE-94 (Code Injection)**: Strictly prohibit `eval()` or unverified dynamic execution of AI-generated code.
- **CWE-89 (SQL Injection)**: Enforce SQLModel/Pydantic parameterized queries; raw string concatenation in SQL is blocked.
- **CWE-306 (Missing Authentication)**: All FastAPI endpoints wrapping ADK runners MUST enforce OAuth2 token or session validation.
- **CWE-798 (Hardcoded Secrets)**: Configuration files MUST decrypt secrets via XOR cipher (`xor:...`) at runtime.

## 3-Phase Execution Protocol

### Phase 1: Validate & Initialize Layout
Scaffold directory tree and initialize uv virtual environment:
```bash
uv init --app enterprise-python-service
cd enterprise-python-service
mkdir -p src/enterprise_service tests configs/terraform .github/workflows docs
```

### Phase 2: Add Dependencies & Wrap in Bazel
Install enterprise dependencies and lock versions:
```bash
uv add "fastapi>=0.115.8" "uvicorn>=0.34.0" "pydantic>=2.13.3" "sqlmodel>=0.0.22" "google-adk>=1.16.0" "google-genai>=1.0.0" "google-cloud-logging>=3.12.0" "google-cloud-bigquery>=3.25.0" "tenacity>=9.0.0"
uv add --dev "pytest>=8.3.4" "pytest-asyncio>=0.23.8" "pytest-cov>=5.0.0" "ruff>=0.4.6"
```

### Phase 3: Run TDD Suite (90% Threshold) & Bazel Build
```bash
uv run ruff check . --fix
uv run pytest --cov=src --cov-report=term-missing --cov-fail-under=90
bazel test //...
```

## Progressive Disclosure References

- **Python Scaffold Guide**: Read [`references/python_scaffold_guide.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/python-project-setup/references/python_scaffold_guide.md).
- **Reference PyProject**: View [`examples/pyproject.toml`](file:///Users/rmcguinness/Projects/skill-builder/skills/python-project-setup/examples/pyproject.toml).
- **Reference Bazel Build**: View [`examples/BUILD.bazel`](file:///Users/rmcguinness/Projects/skill-builder/skills/python-project-setup/examples/BUILD.bazel).
- **Reference Terraform**: View [`examples/configs/terraform/main.tf`](file:///Users/rmcguinness/Projects/skill-builder/skills/python-project-setup/examples/configs/terraform/main.tf).
- **Reference Base Config**: View [`examples/configs/.env.toml`](file:///Users/rmcguinness/Projects/skill-builder/skills/python-project-setup/examples/configs/.env.toml).
