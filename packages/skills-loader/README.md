# Retail Cortex Skills Loader (`retailcortex-skills-loader`)

[![PyPI Version](https://img.shields.io/pypi/v/retailcortex-skills-loader.svg)](https://pypi.org/project/retailcortex-skills-loader/)
[![Python Version](https://img.shields.io/badge/python-3.13%2B-blue.svg)](https://www.python.org/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

Reusable enterprise AI agent skill scanner, parser, and loader compatible with Google ADK (Agent Development Kit).

## Features

- **Local Skill Discovery**: Scans local workspace directories for `SKILL.md` definitions and frontmatter metadata.
- **GitHub Remote Loading**: Dynamically downloads and extracts skill packages directly from remote GitHub repositories or release tarballs/zips.
- **Environment Context Parsing**: Parses `.env` key-value definitions and instructions into typed `SkillDefinition` objects.
- **Google ADK Integration**: Exposes standard tools and registry abstractions (`SkillRegistry`) to equip Google ADK agents with enterprise capabilities.

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
from skills_loader import load_all_skills, SkillRegistry

# Discover and load all skills in the current workspace or search paths
skills = load_all_skills()

for skill in skills:
    print(f"Loaded skill: {skill.name} - {skill.description}")

# Create a SkillRegistry to query loaded skills
registry = SkillRegistry(skills)
```

## Remote GitHub Skill Loading

```python
from skills_loader import load_skills_from_github

skills = load_skills_from_github("retail-cortex/skills", ref="main")
```

## License

Apache License 2.0. See LICENSE for details.
