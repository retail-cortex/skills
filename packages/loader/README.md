# Retail Cortex Skills Loader (`retailcortex-skills-loader`)

[![PyPI Version](https://img.shields.io/pypi/v/retailcortex-skills-loader.svg)](https://pypi.org/project/retailcortex-skills-loader/)
[![Python Version](https://img.shields.io/badge/python-3.13%2B-blue.svg)](https://www.python.org/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

Reusable enterprise AI agent skill scanner, parser, and loader compatible with Google ADK (Agent Development Kit).

## Features

- **Polyglot URI Resolution**: Resolves skills seamlessly from `github://`, `mod://`/`go://`, `maven://`/`mvn://`, `pkg://`, and `file://` URIs.
- **Zero-I/O Manifest Loading**: Loads pre-compiled `skills_manifest.json` definitions directly for instant startup.
- **Local & Remote Discovery**: Scans local workspace directories or downloads remote skills from GitHub repositories, Go module caches, and Maven repositories.
- **Environment Context Parsing**: Parses `.env` key-value definitions and instructions into strongly-typed `SkillDefinition` objects.
- **Google ADK Integration**: Exposes standard `SkillRegistry` abstractions to equip Google ADK agents with enterprise capabilities.

## Installation

```bash
pip install retailcortex-skills-loader
```

Or using `uv`:

```bash
uv add retailcortex-skills-loader
```

## Quickstart

```python
from loader import load_all_skills, SkillRegistry

# Discover and load all skills in the workspace
skills = load_all_skills()

for name, skill in skills.items():
    print(f"Loaded skill: {name} - {skill.description}")

# Instantiate SkillRegistry using polyglot root URIs
registry = SkillRegistry.from_roots([
    "file://.",
    "pkg://retailcortex_skills_python",
    "mod://github.com/retail-cortex/skills@v1.0.0",
    "maven://com.retailcortex.skills:skills-java:1.0.0"
])
```

## Remote GitHub Skill Loading

```python
from loader import load_skills_from_github, SkillRegistry

# Load skills directly from GitHub
skills = load_skills_from_github("retail-cortex/skills", ref="main")

# Instantiate registry directly from remote repository
registry = SkillRegistry.from_github("retail-cortex/skills", ref="main")
```

## Zero-I/O Pre-Compiled Manifest Loading

```python
from pathlib import Path
from loader import load_skills_from_manifest, build_skills_manifest

# Compile manifest for zero-I/O startup
build_skills_manifest(out_path=Path("skills_manifest.json"))

# Load pre-compiled manifest
skills = load_skills_from_manifest(Path("skills_manifest.json"))
```

## Build-Lifecycle Plugin (PEP 517 / Poetry / Setup.py)

The loader can act as a build-time plugin to automatically download and stage skill dependencies before your package is built.

### For PEP 517 Backends (Hatchling, Setuptools, etc.)

Configure your `pyproject.toml` to use the loader's build wrapper:

```toml
[build-system]
requires = ["retailcortex-skills-loader", "setuptools>=61.0"]
build-backend = "loader.build_meta"

[tool.retailcortex-loader]
# Optional: Target directory to bundle the skills into for packaging (defaults to .skills)
dest = "src/my_package/skills"
dependencies = [
    "github://retail-cortex/skills@main/packages/skills-python"
]
```

### For Poetry or Legacy `setup.py`

You can achieve the exact same automated downloading by calling the exposed function directly in your `build.py` (Poetry) or `setup.py`:

```python
from loader import download_build_dependencies

# This will read [tool.retailcortex-loader] from pyproject.toml
download_build_dependencies()
```

## License

Apache License 2.0. See LICENSE for details.
