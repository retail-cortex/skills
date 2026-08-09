# FastAPI Enterprise REST Architecture & TDD Guide

## 1. Asynchronous TDD with httpx.AsyncClient

Test all route handlers, OAuth authentication guards, and dependency overrides using `httpx.AsyncClient` and `pytest-asyncio`:

```python
import pytest
from httpx import AsyncClient, ASGITransport
from main import app

@pytest.mark.asyncio
async def test_health_check():
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as ac:
        response = await ac.get("/healthz")
    assert response.status_code == 200
    assert response.json()["status"] == "healthy"

@pytest.mark.asyncio
async def test_unauthorized_access():
    transport = ASGITransport(app=app)
    async with AsyncClient(transport=transport, base_url="http://test") as ac:
        response = await ac.get("/api/v1/customers")
    assert response.status_code == 401
```

## 2. GitHub Actions CI/CD Workflow (`.github/workflows/fastapi.yml`)

```yaml
name: FastAPI CI/CD

on: [push, pull_request]

jobs:
  test-and-lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: astral-sh/setup-uv@v3
      - run: uv sync --frozen
      - run: uv run ruff check .
      - run: uv run pytest --cov=api --cov-fail-under=90
      
      - name: Export OpenAPI Specification for GitHub Pages
        run: |
          mkdir -p site
          uv run python -c "import json; from main import app; print(json.dumps(app.openapi()))" > site/openapi.json
```

## 3. Secure Coding Checklist

- **Cookie Security**: Always configure `SessionMiddleware(secret_key=..., https_only=True, same_site="lax")`.
- **CORS Policies**: Explicitly define allowed origins; never allow `allow_origins=["*"]` on authenticated endpoints.
- **Input Validation**: Use Pydantic v2 `Field` constraints (e.g. `Field(min_length=1, max_length=255)`) to eliminate injection vectors.
