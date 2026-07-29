# Python Enterprise Scaffolding & Configuration Architecture

## Scaffolding Commands

```bash
#!/usr/bin/env bash
set -euo pipefail

# 1. Initialize uv
uv init --app enterprise-python-service
cd enterprise-python-service

# 2. Directory structure
mkdir -p src/enterprise_service tests configs/terraform .github/workflows docs

# 3. Touch baseline files
touch src/enterprise_service/__init__.py src/enterprise_service/main.py
touch tests/__init__.py tests/test_service.py
touch configs/.env.toml configs/.env.local.toml
touch configs/terraform/main.tf configs/terraform/variables.tf
touch BUILD.bazel
```

## Configuration Loading (`src/enterprise_service/config.py`)

Read cascading TOML files from `configs/` using `tomllib` (built-in to Python 3.11+):

```python
import os
import tomllib
from pathlib import Path
from pydantic import BaseModel

class AppConfig(BaseModel):
    environment: str
    rest_port: int
    project_id: str

def load_config() -> AppConfig:
    config_dir = Path(__file__).parents[2] / "configs"
    base_file = config_dir / ".env.toml"
    local_file = config_dir / ".env.local.toml"

    data = {}
    if base_file.exists():
        with open(base_file, "rb") as f:
            data.update(tomllib.load(f))
    if local_file.exists():
        with open(local_file, "rb") as f:
            data.update(tomllib.load(f))

    return AppConfig(
        environment=data.get("environment", {}).get("name", "development"),
        rest_port=int(data.get("server", {}).get("rest_port", 8000)),
        project_id=data.get("environment", {}).get("project_id", ""),
    )
```
