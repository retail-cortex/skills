---
name: a2a
description: Agent-to-Agent (A2A) protocol & multi-agent orchestration SDLC. Covers component TDD, GitHub Actions CI validation, HTTP 429 rate limit resilience, token bucket backoff, and SemVer protocol releases.
license: Apache-2.0
metadata:
  author: Ryan McGuinness
  version: "1.0"
authors:
  - name: Retail Cortex Engineering
    url: https://github.com/retail-cortex/castor
category: a2a
tags:
  - a2a
  - protocol
  - agent
  - hmac
trigger_phrases:
  - "setup A2A protocol"
  - "agent to agent communication"
  - "A2A message authorization"
execution_hints:
  preferred_model: "gemini-3.1-pro"
  requires_human_approval: false
  environment_variables:
    - A2A_AUTH_TOKEN
  timeout_seconds: 180
---
# Agent-to-Agent (A2A) Protocol & Multi-Agent Orchestration SDLC Skill

This skill prescribes best practices for building distributed multi-agent systems using the **Agent-to-Agent (A2A)** communication protocol, message dispatchers, and autonomous agent orchestration.

## Prerequisites & Pre-Flight Checklist

1. Python 3.13 and `uv` package manager installed.
2. `a2a-agent-sdk==0.2.1` pinned in `pyproject.toml`.

## HTTP 429 Rate Limit & Quota Backoff Invariants

- A2A message routers and inter-agent dispatchers must implement token bucket rate limiting and exponential backoff with full randomized jitter to survive HTTP 429 quota exhaustion.

## Security Checkpoints & CWE Invariants

- **CWE-306 (Missing Authentication for Critical Function)**: Enforce mutual authentication, HMAC message signatures, and recipient verification for all inter-agent message exchanges.
- **CWE-269 (Improper Privilege Management)**: Isolate agent execution contexts; pass delegated end-user OAuth2 tokens via `ToolContext.state["user_token"]` rather than granting agents blanket administrative scopes.
- **CWE-400 (Uncontrolled Resource Consumption)**: Bound maximum message payload sizes and queue lengths on A2A routers to prevent denial-of-service and runaway cascading execution loops.

## 3-Phase Execution Protocol

### Phase 1: Payload Construction & Dispatcher Architecture
Define strongly-typed A2A message envelopes, correlation IDs, and routing channels.

### Phase 2: Implement A2A TDD Payload & Protocol Verification
Write unit test fixtures asserting that message dispatchers, envelope validation, and response handlers conform to A2A schema specifications.

### Phase 3: CI/CD Pipeline & SemVer Release
Validate protocol schemas and execute pytest test suites:
```bash
uv run pytest tests/test_a2a_protocol.py
```

## Progressive Disclosure References

- **A2A Protocol Specification**: Read [`references/a2a_spec.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/a2a/references/a2a_spec.md).
- **Reference Agent Router**: View [`examples/agent_router.py`](file:///Users/rmcguinness/Projects/skill-builder/skills/a2a/examples/agent_router.py).
