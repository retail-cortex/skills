# Examples & Demos

The `examples/` directory contains sample implementations showcasing how to integrate `skills-loader` and `skills-agent` into Google Agent Development Kit (ADK) projects.

---

## 1. Native ADK Python Example (`examples/example_adk.py`)

Demonstrates loading skills from both local workspaces (`file://skills`) and remote GitHub repositories (`github://google/adk:main`).

### Running the Example

```bash
PYTHONPATH=packages/skills-loader/src:packages/skills-agent/src:. uv run python -m examples.example_adk
```

### Key Highlights

- Demonstrates skill loading via `SkillLoader`.
- Registers skills as tools in an ADK runner context.
- Demonstrates prompt execution with grounded skill retrieval.

---

## 2. ADK FastAPI Web Runner (`examples/example-adk/`)

Demonstrates wrapping an ADK agent inside a FastAPI web application with `.env` configuration and environment isolation.

### Directory Layout

```
examples/example-adk/
├── README.md           # Sub-project guide
├── .env.example        # Environment variable template
├── main.py             # FastAPI server entrypoint
└── pyproject.toml      # Package definition
```

### Running the Web Server

```bash
# Using uv script entrypoint
uv run python examples/example-adk/main.py

# Or via FastAPI CLI / uvicorn
uv run uvicorn examples.example-adk.main:app --reload --port 8000
```

---

## Adding New Examples

When contributing new examples:

1. Create a dedicated directory under `examples/` or a python script `examples/example_<name>.py`.
2. Ensure dependencies are declared in `pyproject.toml`.
3. Include clear inline docstrings and usage commands.
