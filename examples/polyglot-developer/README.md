# Polyglot Developer Agent (`examples/polyglot-developer`)

Custom Google ADK agent and CLI tool for scaffolding enterprise polyglot Bazel monorepo projects using domain skills:
- **`skills-bazel`**: Hermetic Bazel build targets and MODULE.bazel resolution.
- **`skills-go`**: Go microservice architecture and gRPC client/server patterns.
- **`skills-java`**: Java enterprise services and Maven/Bazel integration.
- **`skills-protobuf`**: Standardized Protocol Buffer schemas and gRPC contracts.
- **`skills-python`**: Python 3.13 FastAPI, ADK agent, and package setups.
- **`skills-frontend`**: React, Vite, and module federation frontend architecture.

## Execution

Using `uv`:

```bash
uv run polyglot-developer --help
```

Or run directly:

```bash
uv run python examples/polyglot-developer/main.py --target-dir ./my-polyglot-app
```

## Features

1. **Skill-Guided Scaffolding**: Uses loaded skills to inject architecture standards into generated Bazel build files and source packages.
2. **Multi-Language Service Generation**:
   - `proto/`: Shared Protobuf contracts (`protobuf-grpc`).
   - `services/go-backend/`: Go microservice (`go-lang`).
   - `services/java-backend/`: Java enterprise service (`java-enterprise`).
   - `services/python-backend/`: Python FastAPI / ADK agent service (`python-adk-fastapi`).
   - `apps/web-ui/`: React + Vite web dashboard (`react-vite`).
3. **Hermetic Bazel Module Federation**: Generates a unified `MODULE.bazel` binding toolchains for all languages.
