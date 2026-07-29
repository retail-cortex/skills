# ADK & FastAPI Integration Architecture & TDD Guide

## 1. Mock LLM TDD Testing with pytest

Unit test ADK agent reasoning and tool execution without calling live Gemini endpoints by patching the model backend:

```python
import pytest
from unittest.mock import AsyncMock, patch
from google.adk.agents import InvocationContext
from google.adk.sessions import Session, InMemorySessionService
from agent_api import adk_agent

@pytest.mark.asyncio
async def test_adk_agent_tool_execution():
    session_service = InMemorySessionService()
    session = Session(id="test_sess", appName="test_app", userId="user_123")
    session.state["user_token"] = "valid_token"

    context = InvocationContext(
        agent=adk_agent,
        session=session,
        session_service=session_service,
        invocation_id="test_inv",
        request="Show me account summary for ACC-999",
    )

    with patch.object(adk_agent, "run_async") as mock_run:
        async def mock_generator(ctx):
            yield "Account balance is 450000.0"
        mock_run.side_effect = mock_generator

        chunks = []
        async for chunk in adk_agent.run_async(context):
            chunks.append(chunk)

        assert "450000.0" in "".join(chunks)
```

## 2. Server-Sent Events (SSE) Streaming Client & Uvicorn Runtime

Wrap agent runners in FastAPI and stream responses asynchronously via Uvicorn:

```bash
# Production Execution via Uvicorn
uv run uvicorn agent_api:app --host 0.0.0.0 --port 8000 --workers 4
```

## 3. GitHub Actions CI/CD (`.github/workflows/agent_api.yml`)

```yaml
name: ADK Agent API CI

on: [push, pull_request]

jobs:
  agent-verification:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: astral-sh/setup-uv@v3
      - run: uv sync --frozen
      - run: uv run pytest --cov=app --cov-fail-under=90
```
