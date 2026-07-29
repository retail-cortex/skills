---
name: python-fastmcp
description: Elite Model Context Protocol (MCP) server SDLC using FastMCP. Covers async TDD, prompt injection defense (CWE-94), HTTP 429 rate limit backoff, Pydantic bounds validation, and SemVer.
license: Apache-2.0
metadata:
  author: Ryan McGuinness
  version: "1.0"
---

# FastMCP Server Development & SDLC Skill

This skill prescribes best practices for building, registering, testing, securing, and deploying **FastMCP** servers for AI Agent tool integration.

## Prerequisites & Pre-Flight Checklist

1. Python 3.13 and `uv` installed.
2. Target AI agent client configured to receive MCP tool JSON schemas.

## HTTP 429 Rate Limit & Backoff Invariants

- Tool handlers calling external AI models or third-party APIs must implement exponential backoff with jitter to gracefully handle HTTP 429 rate limits.

## Security Checkpoints & CWE Invariants

- **CWE-94 (Prompt Injection & Code Execution)**: Never pass unverified agent tool arguments directly to shell commands, SQL query strings, or administrative APIs.
- **CWE-20 (Improper Input Validation)**: Strictly validate and constrain all tool input fields via Pydantic (`Field(ge=1, le=100)`).
- **CWE-272 (Least Privilege Violation)**: Only expose the minimal set of tools required for the agent's specific domain.

## 3-Phase Execution Protocol

### Phase 1: Initialize FastMCP & System Instructions
Define server instance with explicit instructions and domain boundaries.

### Phase 2: Implement Tool Handlers & TDD Suite
Decorate async tool functions with `@mcp.tool()` and write unit tests evaluating tool response schemas.

### Phase 3: Validate MCP Schema & Run Server
```bash
uv run pytest tests/test_mcp_server.py
uv run python server.py --export-schema
uv run python server.py
```

## Progressive Disclosure References

- **FastMCP Protocol Guide**: Read [`references/fastmcp_spec.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/python-fastmcp/references/fastmcp_spec.md).
- **Reference FastMCP Server**: View [`examples/server.py`](file:///Users/rmcguinness/Projects/skill-builder/skills/python-fastmcp/examples/server.py).
