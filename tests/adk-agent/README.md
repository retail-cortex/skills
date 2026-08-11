# Castor Skills Agent (`castor-skills-agent`)

Enterprise Python 3.13 package deploying an interactive Google Agent Development Kit (ADK) programming agent wrapped inside a FastAPI web service for the Castor registry ecosystem.

## Architecture Overview

The `castor-agent` package exposes an interactive CLI and FastAPI service that orchestrates all enterprise skills.

### Key Components

- [SkillRegistry](file:///Users/rmcguinness/Projects/skill-builder/packages/skills-agent/src/skills_agent/skills_loader.py): Dynamically loads and indexes all enterprise skills from `/skills`, parsing YAML frontmatter, markdown specifications, `references/`, and `examples/`.
- [ADKProgrammingAgent](file:///Users/rmcguinness/Projects/skill-builder/packages/skills-agent/src/skills_agent/agent.py): Implements the ADK programming agent, equipped with skill toolsets, token delegation via `ToolContext`, and `tenacity` exponential backoff for HTTP 429 rate limit resilience.
- [FastAPIAgentServer](file:///Users/rmcguinness/Projects/skill-builder/packages/skills-agent/src/skills_agent/server.py): Wraps the ADK agent inside a FastAPI application with REST endpoints, SSE streaming (`/api/v1/agent/chat`), session lifecycles, and health checks.
- [InteractiveCLI](file:///Users/rmcguinness/Projects/skill-builder/packages/skills-agent/src/skills_agent/cli.py): Terminal REPL providing direct CLI interaction with the ADK agent and skills ecosystem via `bazel run //:start-agent`.

## Execution

### Bazel 9.2 (Hermetic CLI Entrypoint)

```bash
bazel run //:start-agent
```


### Python / uv

```bash
uv run python -m skills_agent.cli
```

### FastAPI Server Mode

```bash
uv run uvicorn skills_agent.server:app --host 127.0.0.1 --port 8000 --reload
```
