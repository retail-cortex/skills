---
name: python-adk-fastapi
description: Superior enterprise SDLC for deploying Google Agent Development Kit (ADK) agents wrapped in FastAPI on Uvicorn. Covers Google OAuth2 end-user token delegation, HTTP 429 rate limit resilience, mock LLM TDD, and ToolContext IAM state isolation.
---

# ADK FastAPI Agent Integration & SDLC Skill (Superior Pattern)

This skill prescribes the **superior enterprise pattern** for deploying **Google Agent Development Kit (ADK)** agents: wrapping the agent execution runner inside a **FastAPI** web service, executing it on **Uvicorn**, and enforcing **Google OAuth2 End-User Token Delegation** alongside HTTP 429 resilience.

## Prerequisites & Pre-Flight Checklist

1. Python 3.13 and `uv` package manager installed.
2. Google Cloud project configured with OAuth 2.0 Web Client ID and Vertex AI access.
3. Verify that `google-genai` is used exclusively; `google-generativeai` is deprecated and prohibited.

## Google OAuth2 Token Delegation to ADK Tools & CAPI

1. **Authentication & Session Attachment**:
   - Inbound FastAPI requests are authenticated by validating the Google OAuth2 Bearer token (`id_token.verify_oauth2_token`).
   - Extract the user's validated OAuth access token (`user_oauth_token`) and attach it to the `Session.state`:
     ```python
     session = session_service.get_session(session_id)
     session.state["user_token"] = user_oauth_token
     session.state["user_email"] = id_info["email"]
     ```
2. **ToolContext Delegation (CWE-269)**:
   - When tools (such as BigQuery CAPI or Cloud Storage readers) execute, they access `tool_context.state["user_token"]` to instantiate credentials:
     ```python
     from google.oauth2.credentials import Credentials
     user_creds = Credentials(token=tool_context.state["user_token"])
     ```
   - This ensures all agent actions execute with the exact IAM permissions and row-level security of the authenticated human user.

## HTTP 429 Rate Limit & Quota Resilience Invariants

- **Exponential Backoff with Full Jitter**: Prohibit immediate retry loops. Use `tenacity` with `wait_random_exponential` when calling Gemini API.
- **Client-Side Concurrency Throttling**: Throttle outbound LLM requests using `aiolimiter.AsyncLimiter` to avoid TPM/RPM quota spikes.

## Security Checkpoints & CWE Invariants

- **CWE-284 (Improper Access Control)**: Session state inside `ToolContext` MUST be strictly partitioned per user ID and session UUID.
- **CWE-319 (Cleartext Transmission)**: Enforce TLS for Server-Sent Events (SSE) streaming connections.

## 3-Phase Execution Protocol

### Phase 1: Initialize ADK Agent & FastAPI Wrapper
Wrap ADK agent runner inside FastAPI, managing session lifecycles via `InMemorySessionService` and configuring Google OAuth2 auth guards.

### Phase 2: Implement Mock LLM & OAuth2 TDD Suite (90% Coverage)
Write unit tests evaluating agent reasoning, 429 rate limit backoff retries, and Google OAuth2 token delegation.

### Phase 3: Execute on Uvicorn & Publish MkDocs
```bash
uv run pytest tests/test_agent_api.py --cov=app --cov-fail-under=90
uv run uvicorn main:app --host 0.0.0.0 --port 8000 --reload
```

## Progressive Disclosure References

- **ADK Architecture Guide**: Read [`references/adk_fastapi_architecture.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/python-adk-fastapi/references/adk_fastapi_architecture.md).
- **Reference API Implementation**: View [`examples/agent_api.py`](file:///Users/rmcguinness/Projects/skill-builder/skills/python-adk-fastapi/examples/agent_api.py).
