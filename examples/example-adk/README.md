# Native Google ADK Agent Example (Local, Remote GitHub & Dotenv)

This example demonstrates how to build a native **Google Agent Development Kit (ADK)** agent that dynamically loads and synthesizes AI Agent Skills from both local workspace directories and remote **GitHub repositories** using qualified URI schemes (`file://` and `github://`), per-root branch specifiers (`:branch`), and selective skill loading (`SKILLS_FILTER`).

## Architecture & Features

1. **Qualified Multi-Root URIs (`SKILLS_ROOTS`)**:
   - `file://skills`: Scans local enterprise skills directory.
   - `github://google/skills/skills/cloud/gemini-api:main`: Fetches remote skills directly from GitHub at specified branch/ref (`:main`) and subpath.

2. **Selective Skill Loading (`SKILLS_FILTER`)**:
   - Filter skills by name (`a2a,bigquery,python-core,gemini-api`) in `.env`, ignoring unselected skills to optimize memory and context windows.

3. **Unified SkillRegistry & ADK Toolset**:
   - A single `SkillRegistry(dotenv_path=".env")` loads all qualified local and GitHub skills into an `ADKSkillToolset`.

4. **ADK Session & Token Delegation**:
   - Manages state, conversational history, and OAuth2 Bearer token delegation via `Session` and `ToolContext`.

---

## Setup & Environment Configuration

1. Copy the environment template:
   ```bash
   cp examples/example-adk/.env.example examples/example-adk/.env
   ```

2. Inspect `.env` configuration:
   ```env
   # Qualified skill roots
   SKILLS_ROOTS=file://skills,github://google/skills/skills/cloud/gemini-api:main

   # Select specific skills to register
   SKILLS_FILTER=a2a,bigquery,python-core,gemini-api
   ```

---

## Running the Example

### 1. Script Runners Configured in `pyproject.toml`

Run the example script or web control plane via workspace entry points:

```bash
# Run native ADK example script
PYTHONPATH=packages/skills-loader/src:packages/skills-agent/src:. uv run python -m examples.example_adk

# Run ADK example web server control plane
PYTHONPATH=packages/skills-loader/src:packages/skills-agent/src:. uv run python -c "from examples.example_adk import web; web()"
```

### 2. Direct Execution via `main.py`

Run the native ADK agent script directly:

```bash
uv run python examples/example-adk/main.py
```

### 3. Interactive ADK CLI Interface (`start-agent`)

Launch the interactive REPL CLI powered by ADK:

```bash
PYTHONPATH=packages/skills-loader/src:packages/skills-agent/src uv run start-agent
```

Inside the `adk>` prompt, run commands or query skills:
```text
adk> skills
adk> search bigquery
adk> How do I build resilient BigQuery CAPI services with Gemini API skills?
```

### 4. Google ADK Web Server Control Plane

Spin up the Google ADK Web UI control plane server:

```bash
PYTHONPATH=packages/skills-loader/src:packages/skills-agent/src uv run adk web --port 8000
```
