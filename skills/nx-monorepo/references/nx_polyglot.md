# NX Polyglot Workspace Integration (Python uv + React + Go)

## Combining Python uv Workspaces with NX

When Python packages depend on each other inside an NX workspace, map them in root `pyproject.toml` and configure `project.json` for NX task tracking:

```toml
# Root pyproject.toml
[tool.uv.workspace]
members = ["apps/*", "packages/*"]

[tool.uv.sources]
media-vault-env_config = { path = "packages/env_config" }
```

## Project JSON Task Definitions (`project.json`)

Define runnable tasks for each Python or TypeScript package:

```json
{
  "name": "media_vault_api",
  "projectType": "application",
  "targets": {
    "test": {
      "executor": "nx:run-commands",
      "options": {
        "command": "uv run pytest",
        "cwd": "apps/media_vault_api"
      }
    }
  }
}
```
