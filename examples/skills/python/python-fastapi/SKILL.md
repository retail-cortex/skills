---
name: python-fastapi
description: Elite production FastAPI REST microservice SDLC. Covers Google OAuth2 JWT verification, async TDD with httpx, HTTP 429 rate limiting, 90% coverage, and OpenAPI export to GitHub Pages.
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
  - fastapi
  - rest
  - slowapi
trigger_phrases:
  - "build FastAPI REST API"
  - "FastAPI rate limiting setup"
  - "FastAPI OAuth2 bearer"
execution_hints:
  preferred_model: "gemini-3.1-pro"
  requires_human_approval: false
  environment_variables: []
  timeout_seconds: 240
---
# FastAPI Enterprise REST SDLC Skill

This skill prescribes best practices for designing, securing, testing, and deploying enterprise-grade **FastAPI** web services in Python 3.13, incorporating **Google OAuth2 Authentication**, **HTTP 429 Rate Limiting**, and **TDD**.

## Prerequisites & Pre-Flight Checklist

1. Python 3.13 and `uv` package manager installed.
2. Google Cloud project with OAuth 2.0 Client ID and Secret provisioned in Google Cloud Console.
3. Target API audience (`GOOGLE_CLIENT_ID`) configured in `configs/.env.toml`.

## Google OAuth2 Provider Architecture for APIs

1. **Google ID Token & Access Token Verification**:
   - Protect routes using a FastAPI `Security(HTTPBearer())` dependency.
   - Validate token signature against Google's public JWKS using `google-auth`:
     ```python
     from fastapi import Depends, HTTPException, Security, status
     from fastapi.security import HTTPBearer, HTTPAuthorizationCredentials
     from google.oauth2 import id_token
     from google.auth.transport import requests

     security = HTTPBearer()

     async def verify_google_token(credentials: HTTPAuthorizationCredentials = Security(security)):
         token = credentials.credentials
         try:
             id_info = id_token.verify_oauth2_token(
                 token, requests.Request(), audience=GOOGLE_CLIENT_ID
             )
             return id_info # Contains email, sub (user_id), name
         except ValueError:
             raise HTTPException(
                 status_code=status.HTTP_401_UNAUTHORIZED,
                 detail="Invalid or expired Google OAuth2 token"
             )
     ```
2. **Session Cookie Encryption**:
   - For web sessions, use `SessionMiddleware(secret_key=..., https_only=True, same_site="lax")` storing encrypted session tokens.
3. **Negative OAuth2 TDD Assertions**:
   - Test suites MUST simulate valid Google tokens, expired tokens (`401 Unauthorized`), mismatched audience (`403 Forbidden`), and missing `Authorization` headers.

## HTTP 429 Rate Limiting & Invariants

- **Slowapi Inbound Rate Limiting**: Protect authenticated endpoints with token buckets. Return HTTP 429 status codes with `Retry-After` headers.
- **Outbound AI Call Resilience**: Wrap external API calls in `tenacity` exponential backoff with randomized jitter.

## Security Checkpoints & CWE Invariants

- **CWE-384 (Session Fixation)**: Regenerate session identifiers upon Google OAuth2 login callback.
- **CWE-346 (Origin Validation Error)**: Explicitly restrict CORS origins; never allow wildcard `allow_origins=["*"]` on authenticated endpoints.
- **CWE-20 (Improper Input Validation)**: Enforce Pydantic v2 validation on all request bodies and route parameters.

## 3-Phase Execution Protocol

### Phase 1: Initialize FastAPI App & Google OAuth2 Guards
Configure session management, Google OAuth2 token verifiers, and slowapi rate limiters.

### Phase 2: Implement Async TDD Suite (90% Coverage)
Write unit and integration tests using `httpx.AsyncClient` and `pytest-asyncio`, mocking Google Auth token verifiers.

### Phase 3: Export OpenAPI Docs & Launch on Uvicorn
```bash
uv run pytest --cov=api --cov-fail-under=90
uv run python -c "import json; from main import app; print(json.dumps(app.openapi()))" > docs/openapi.json
uv run uvicorn main:app --host 0.0.0.0 --port 8000 --reload
```

## Progressive Disclosure References

- **FastAPI Architecture Guide**: Read [`references/fastapi_patterns.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/python-fastapi/references/fastapi_patterns.md).
- **Reference Application**: View [`examples/main.py`](file:///Users/rmcguinness/Projects/skill-builder/skills/python-fastapi/examples/main.py).
