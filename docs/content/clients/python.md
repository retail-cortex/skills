---
title: "Python Client & Build Integration"
weight: 30
---

# Python Client & `uv` Build Integration (`loader.build_meta`)

The Python client (`retailcortex-loader`) implements a **PEP 517/518 build backend wrapper** (`loader.build_meta`) alongside high-performance **JIT Dynamic Pre-Call Retrieval** for Google ADK agents.

By declaring `loader.build_meta` in your project's `pyproject.toml`, native Python package managers (`uv build`, `pip install .`, `build`) execute pre-build hooks that download, audit, and stage skill dependencies into the package tree prior to generating wheels (`.whl`) or source distributions (`.tar.gz`).

---

## 1. Native `uv` & PEP 517 Build Integration (`pyproject.toml`)

Configure `loader.build_meta` as your project's build backend in `pyproject.toml`:

```toml
[build-system]
requires = ["retailcortex-loader>=1.0.0", "setuptools>=68.0.0"]
build-backend = "loader.build_meta"

[project]
name = "my-enterprise-agent"
version = "1.0.0"
description = "Enterprise Agent Application with Embedded Skills"
readme = "README.md"
requires-python = ">=3.11"
dependencies = [
    "google-adk>=0.1.0",
    "retailcortex-loader>=1.0.0",
    "pydantic>=2.7.0",
]

[tool.retailcortex-loader]
# Target staging directory inside python package tree
dest = "src/my_enterprise_agent/.skills"
# Skill root URIs to resolve and package during build
dependencies = [
    "skm://skills/sk-9b1deb4d",
    "github://google/skills@main/tree/main/skills/cloud/gemini-api",
    "file://./local_skills/customer-search"
]
```

### What Happens During `uv build` or `pip install .`:

1. **PEP 517 Hook Interception**: When `uv build` or `pip install` executes, the `loader.build_meta` build backend intercepts the build process (`build_wheel`, `build_sdist`, `prepare_metadata_for_build_wheel`).
2. **Dependency Resolution**: Reads `[tool.retailcortex-loader.dependencies]` and downloads/resolves all specified skills from SKM servers, GitHub, or local paths.
3. **SDLC Audit Gate**: Performs 5-point SDLC checks. If validation fails, **the build terminates with an error**.
4. **Staging & Packaging**: Stages resolved skills into `dest` (`src/my_enterprise_agent/.skills/`), ensuring they are packaged directly into the final `.whl` artifact.

---

## 2. JIT Dynamic Pre-Call Retrieval (`suggest_skills`)

In autonomous ADK agent workflows, loading all tools statically can cause context window bloat and tool hallucinations. `SkillRegistry.suggest_skills()` performs JIT semantic retrieval to rank and bound relevant skills to at most $k \le 3$:

```python
from loader import SkillRegistry

# Initialize registry from workspace or embedded package tree
registry = SkillRegistry()

# Dynamically retrieve top 3 skills ranked by vector relevance
suggested = registry.suggest_skills(
    prompt="Generate BigQuery SQL analytics statement for customer churn",
    max_skills=3,
    server_url="http://localhost:8000"
)

for skill in suggested:
    print(f"Loaded Skill: {skill.name} ({skill.description})")
```

### Fallback Resilience Strategy
1. **Remote Vector Search**: Dispatches `GET /api/v1/skills?s={query}&page_size=3` to the central `skills-service`.
2. **Local Vector Search**: If offline or unreachable, executes local TF-IDF cosine similarity search via `DiscoveryEngine`.
3. **Keyword & Substring Fallback**: Falls back to tokenized keyword matching and domain heuristics.

---

## 3. Grounding Google ADK Agents with Dynamic Skills

```python
import asyncio
from google.adk.agent import Agent
from loader import SkillRegistry

class EnterpriseCodingAgent:
    def __init__(self, registry: SkillRegistry) -> None:
        self.registry = registry

    async def handle_prompt(self, user_prompt: str) -> str:
        # Pre-call optimization: retrieve top 3 most relevant skills
        relevant_skills = self.registry.suggest_skills(user_prompt, max_skills=3)
        
        # Synthesize system instructions and tools
        system_instructions = "You are an enterprise software engineering agent.\n"
        tools = []
        for skill in relevant_skills:
            system_instructions += f"\n### Skill: {skill.name}\n{skill.instructions}\n"
            tools.extend(skill.to_adk_tools())

        agent = Agent(
            name="enterprise-programmer",
            model="gemini-2.0-flash",
            instructions=system_instructions,
            tools=tools,
        )

        return await agent.run_async(user_prompt)
```

---

## 4. Hermetic Bazel Build Integration (`rules_python`)

In `BUILD.bazel`:

```starlark
py_library(
    name = "agent_lib",
    srcs = glob(["src/my_enterprise_agent/*.py"]),
    deps = [
        "//clients/python:loader_lib",
    ],
)

py_binary(
    name = "agent_binary",
    srcs = ["src/my_enterprise_agent/main.py"],
    deps = [":agent_lib"],
)
```

---

## Recommended Python Standards

1. **Virtual Environment Isolation**: Always manage dependencies using `uv` and run scripts via `uv run python main.py`. Never run global `pip install`.
2. **Type Safety**: Apply strict type hints across custom tools (`def my_tool(param: str) -> dict[str, Any]:`). No `Any` unless unavoidable.
3. **Async Parameter Handling**: Await asynchronous parameter resolution in Next.js / FastAPI API route handlers.
