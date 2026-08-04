---
title: "Python Client & PEP 517 Build Integration"
weight: 30
---

# Python Client & `uv` Build Integration (`loader.build_meta`)

The Python client (`retailcortex-loader`) implements a **PEP 517/518 build backend wrapper** (`loader.build_meta`).

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

## 2. Hermetic Bazel Build Integration (`rules_python`)

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

## 3. Zero-I/O Cold-Start Loading in Python Runtimes

Because skills are staged directly inside the installed Python package tree, your application loads skills with **zero network overhead**:

```python
import importlib.resources
import os
from loader import load_skills_from_dir, load_skills_from_manifest

def init_agent_skills():
    # Resolve staged skills directory inside installed package
    package_skills_dir = importlib.resources.files("my_enterprise_agent").joinpath(".skills")
    
    if package_skills_dir.is_dir():
        skills = load_skills_from_dir(str(package_skills_dir))
        print(f"Instantly loaded {len(skills)} skills from package tree.")
        return skills
    
    # Fallback to pre-compiled manifest
    return load_skills_from_manifest("skills_manifest.json")
```

---

## 4. Grounding Google ADK Agents with Staged Skills

Environment and connection properties are loaded using `python-dotenv` (`load_dotenv()`):

```python
import os
import importlib.resources
from dotenv import load_dotenv
from google.adk.agent import Agent
from loader import load_skills, load_skills_from_dir

# 1. Load environment properties from .env file
load_dotenv()

server_url = os.getenv("SKM_SERVER_URL", "http://localhost:8080")
api_key = os.getenv("SKM_API_KEY")

# 2. Load staged skills from package or remote SKM server
skills_path = importlib.resources.files("my_enterprise_agent").joinpath(".skills")
if skills_path.is_dir():
    skills = load_skills_from_dir(str(skills_path))
else:
    skills = load_skills("skm://skills/sk-9b1deb4d", server_url=server_url, api_key=api_key)

gemini_skill = skills.get("gemini-api")

# 3. Instantiate ADK Agent grounded in validated skill instructions
agent = Agent(
    name="retail-assistant",
    model="gemini-2.0-flash",
    instructions=gemini_skill.instructions,
    tools=gemini_skill.to_adk_tools(),
)
```


---

## Recommended Python Standards

1. **Virtual Environment Isolation**: Always manage dependencies using `uv` and run scripts via `uv run python main.py`. Never run global `pip install`.
2. **Type Safety**: Apply strict type hints across custom tools (`def my_tool(param: str) -> dict[str, Any]:`).
3. **Async Parameter Handling**: Await asynchronous parameter resolution in Next.js / FastAPI API route handlers.
