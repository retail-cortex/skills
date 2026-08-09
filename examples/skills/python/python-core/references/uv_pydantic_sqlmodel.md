# Python uv, SQLModel & Google Cloud SDK Architecture

## Dependency Specifications (`pyproject.toml`)

Modern Python projects require `hatchling` or `flit` build backends with `uv` source mapping for monorepos:

```toml
[build-system]
requires = ["hatchling"]
build-backend = "hatchling.build"

[project]
name = "enterprise-service"
version = "0.1.0"
requires-python = ">=3.13"
dependencies = [
    "pydantic>=2.13.0",
    "sqlmodel>=0.0.22",
    "google-genai>=1.0.0",
    "google-cloud-logging>=3.12.0",
    "google-cloud-bigquery>=3.25.0",
]

[tool.ruff]
line-length = 100
target-version = "py313"

[tool.ruff.lint]
select = ["E", "F", "W", "I", "UP", "RUF"]
```

## Google GenAI SDK Usage (`google-genai`)

Always initialize GenAI clients targeting Vertex AI backends:

```python
from google import genai
from google.genai import types

client = genai.Client(
    project="enterprise-gcp-project",
    location="us-central1",
    backend=genai.BackendVertexAI,
)

response = client.models.generate_content(
    model="gemini-2.5-flash",
    contents="Analyze enterprise customer dataset schema.",
    config=types.GenerateContentConfig(
        temperature=0.2,
        max_output_tokens=1024,
    ),
)
```
