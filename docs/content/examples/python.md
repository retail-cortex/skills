---
title: Python Integration Examples
description: Standalone Python client integration with python-dotenv property loading, Polyglot Monorepo Scaffolding Agent, and Python skill packages.
weight: 20
---

# Python Integration Examples

This section details the Python integration examples in `examples/python/`, covering standalone Python client integration, the Polyglot Monorepo Scaffolding Agent, and Python enterprise skill collections.

---

## 1. Standalone Python Client Example

Located at `examples/python/client`, this example demonstrates how a standalone Python application loads environment configuration using `python-dotenv` and parses skill URIs via `castor-client`.

### Key Features & Design
- **PEP 621 Standard Package**: Configured with `pyproject.toml` using `uv` dependency specifications.
- **Environment Property Loading**: Uses `python-dotenv` (`load_dotenv()`) to read settings from `examples/python/client/.env`.
- **Polyglot URI Parsing**: Exercises `castor_client.parse_skill_root_uri` across `file://`, `pkg://`, `github://`, and `castor://` schemes.

### Project Layout

```text
examples/python/client/
├── .env                     # Local environment configuration file
├── BUILD.bazel              # Bazel test rule (test_python_client_example)
├── main.py                  # Entry point demonstrating dotenv & URI resolution
├── pyproject.toml           # Standalone Python package configuration
└── test_client.py           # Native unittest suite
```

### Application Walkthrough (`main.py`)

```python
import os
import sys
from pathlib import Path
from dotenv import load_dotenv
from castor_client import parse_skill_root_uri

def run() -> dict[str, str]:
    # 1. Load environment properties
    env_path = Path(__file__).parent / ".env"
    load_dotenv(dotenv_path=env_path)

    server_url = os.getenv("CASTOR_SERVER_URL", "http://localhost:8080")
    api_key = os.getenv("CASTOR_API_KEY", "")
    print(f"Loaded Castor Server URL from dotenv: {server_url}")

    # 2. Parse polyglot skill URI
    uri = "github://google/skills@main/tree/main/skills/cloud/gemini-api"
    scheme, target, ref, subpath = parse_skill_root_uri(uri)
    print(f"Parsed URI: scheme={scheme}, target={target}, ref={ref}, subpath={subpath}")

    return {
        "server_url": server_url,
        "api_key": api_key,
        "scheme": scheme,
        "target": target,
    }

if __name__ == "__main__":
    run()
```

### Execution Commands

```bash
# Native Python Test Execution
cd examples/python/client
python3 -m unittest test_client.py

# Bazel Workspace Integration Test
bazel test //examples/python/client:test_python_client_example
```

---

## 2. Polyglot Bazel Developer Agent (`examples/python/polyglot`)

Located at `examples/python/polyglot`, this example showcases a custom Google ADK agentic CLI tool that loads cross-language skill packages (`skills-bazel`, `skills-go`, `skills-java`, `skills-python`) to automatically scaffold a polyglot Bazel 9 monorepo.

### Agent Workflow
1. Discovers and parses skill definitions from workspace package roots.
2. Evaluates rule sets and architectural templates for Go, Java, and Python targets.
3. Generates root `MODULE.bazel`, language-specific subdirectories, and `.bazelrc` build flags.

### CLI Execution

```bash
# Run monorepo scaffolding CLI tool via uv
uv run python examples/python/polyglot/main.py --target-dir ./my-polyglot-app

# Execute Bazel test suite
bazel test //examples/python/polyglot:test_polyglot_developer
```

---

## 3. Python Enterprise Skills Package

Located at `examples/python/skills`, this package contains enterprise Python SDLC skills:

| Skill Directory | Skill Name | Category | Description |
| :--- | :--- | :--- | :--- |
| `src/retailcortex_skills_python/skills/python-core` | `python-core` | Python | Python 3.13 standards, `uv` packaging, strict typing, and `pytest` table-driven tests. |
| `src/retailcortex_skills_python/skills/python-fastapi` | `python-fastapi` | Python | FastAPI asynchronous web APIs, Pydantic v2 schemas, and OpenAPI documentation. |
| `src/retailcortex_skills_python/skills/python-fastmcp` | `python-fastmcp` | Python | Model Context Protocol (MCP) server development using FastMCP. |
| `src/retailcortex_skills_python/skills/python-adk-fastapi` | `python-adk-fastapi` | Python | Google Agent Development Kit (ADK) integration with FastAPI backends. |
| `src/retailcortex_skills_python/skills/python-project-setup` | `python-project-setup` | Python | Monorepo layout, `pyproject.toml` workspace rules, and `Ruff` linting. |
