# Retail Cortex Skill Validator (`retailcortex-skill-validator`)

[![PyPI Version](https://img.shields.io/pypi/v/retailcortex-skill-validator.svg)](https://pypi.org/project/retailcortex-skill-validator/)
[![Python Version](https://img.shields.io/badge/python-3.13%2B-blue.svg)](https://www.python.org/)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)

Automated 5-Point SDLC & Security Audit validator for `agentskills.io` compatible AI Agent Skills.

## 5-Point SDLC & Security Audit Rules

1. **YAML Frontmatter**: Validates `name`, `description`, `license`, `author`, and `version` metadata in `SKILL.md`.
2. **L3 Directory Structure**: Ensures `references/` and `examples/` subdirectories exist and contain non-empty documentation/assets.
3. **Security Checkpoint (CWE)**: Verifies presence of security guidelines, CWE vulnerability disclosures, or sandboxing checkpoints.
4. **HTTP 429 Resilience**: Validates rate-limiting, exponential backoff, or quota retry strategy documentation.
5. **Clickable File Links**: Enforces `[Link Text](file:///path/to/file)` format for external and internal reference links.

## Installation

```bash
pip install retailcortex-skill-validator
```

Or using `uv`:

```bash
uv add retailcortex-skill-validator
```

## CLI Usage

```bash
# Audit skills in a directory recursively
uv run python -m validator.cli audit -r ./skills

# Export audit summary to JSON report
uv run python -m validator.cli audit -r ./skills --json
```

## Programmatic Usage

```python
from pathlib import Path
from validator import audit_all_skills, audit_skill_directory

# Audit a single skill directory
result = audit_skill_directory(Path("./skills/python-core"))
print(f"Passed: {result.passed}, Errors: {result.errors}")

# Audit an entire registry recursively
summary = audit_all_skills(Path("./skills"), recursive=True)
print(f"Total: {summary.total_skills}, Passed: {summary.passed_skills}, Failed: {summary.failed_skills}")
```

## License

Apache License 2.0. See LICENSE for details.
