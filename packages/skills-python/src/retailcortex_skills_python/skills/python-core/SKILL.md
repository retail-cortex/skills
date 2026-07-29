---
name: python-core
description: Elite Python 3.13 backend SDLC using uv wrapped in Bazel. Enforces pytest TDD with paired positive/negative tests, HTTP 429 rate limit backoff, None safety, 90% coverage, and SemVer.
license: Apache-2.0
metadata:
  author: Ryan McGuinness
  version: "1.0"
authors:
  - name: Retail Cortex Engineering
    url: https://github.com/retail-cortex/skills
category: python
tags:
  - python
  - uv
  - typing
  - pytest
trigger_phrases:
  - "write Python core code"
  - "uv Python project setup"
  - "Python strict type hints"
execution_hints:
  preferred_model: "gemini-3.1-pro"
  requires_human_approval: false
  environment_variables: []
  timeout_seconds: 180
---
# Python Core & Ecosystem SDLC Skill

This skill enforces enterprise Python engineering standards leveraging **Python 3.13**, **uv**, **Pydantic v2**, **SQLModel**, **Google GenAI SDK** (`google-genai`), **Test-Driven Development (TDD)**, **90% Code Coverage**, **HTTP 429 Rate Limit Resilience**, and **Null/None Safety**.

## Prerequisites & Pre-Flight Checklist

1. Python 3.13 runtime and `uv` package manager installed.
2. Google Cloud SDK authenticated with access to Vertex AI.

## HTTP 429 Rate Limit & Quota Resilience Invariants

- All external API calls to Gemini and Google Cloud SDKs MUST use `tenacity` exponential backoff with full randomized jitter to mitigate HTTP 429 quota exhaustion.

## Defensive Error Handling & Null/None Safety Invariants

- All function signatures MUST use strict type annotations (`typing.Optional[T]` or `T | None`). Bare `typing.Any` is strictly prohibited.
- Guard against `NoneType` errors via Pydantic or defensive `.get()` with explicit defaults.
- Every module MUST include paired positive and negative test cases in `tests/`.

## Security Checkpoints & CWE Invariants

- **CWE-89 (SQL Injection)**: Use SQLModel/SQLAlchemy parameterized queries exclusively. Raw SQL string formatting is strictly rejected.
- **CWE-94 (Code Injection)**: `eval()`, `exec()`, or dynamic imports based on untrusted user strings are prohibited.
- **CWE-798 (Hardcoded Credentials)**: All GCP service keys and DB passwords must be resolved dynamically via GCP Secret Manager or XOR-encrypted TOML configs.

## 3-Phase Execution Protocol

### Phase 1: Dependency Management with uv
Initialize environment, add dependencies, and generate `uv.lock`.

### Phase 2: Implement Positive & Negative TDD Suite (90% Coverage)
Write unit tests using `pytest` and `pytest-asyncio`, mocking all external Google Cloud API calls and asserting both happy path and exception propagation.

### Phase 3: Lint, Verify Coverage & Publish MkDocs
```bash
uv run ruff check . --fix
uv run pytest --cov=src --cov-report=term-missing --cov-fail-under=90
uv run mkdocs build
bazel test //...
```

## Progressive Disclosure References

- **Package & Model Reference**: Read [`references/uv_pydantic_sqlmodel.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/python-core/references/uv_pydantic_sqlmodel.md).
- **Reference PyProject**: View [`examples/pyproject.toml`](file:///Users/rmcguinness/Projects/skill-builder/skills/python-core/examples/pyproject.toml).
- **Reference Models**: View [`examples/models.py`](file:///Users/rmcguinness/Projects/skill-builder/skills/python-core/examples/models.py).
