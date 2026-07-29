# Standalone Examples & Demonstrations

The `examples/` directory contains standalone, self-contained example packages demonstrating how to integrate `skills-loader` into Google Agent Development Kit (ADK) projects and build agentic CLI developer tools.

---

## 1. Native ADK Agent Example (`examples/example-adk`)

Demonstrates loading skills from both local workspaces (`file://skills`) and remote GitHub repositories (`github://google/skills/skills/cloud/gemini-api:main`) using `.env` skill filtering and qualified URI resolution.

### Directory Layout

```text
examples/example-adk/
├── .env
├── .env.example
├── README.md
├── main.py
└── pyproject.toml
```

### Execution

```bash
# Direct execution via uv
uv run python examples/example-adk/main.py

# Or run within example directory
cd examples/example-adk && uv run python main.py
```

### Key Highlights

- Demonstrates qualified URI parsing (`file://` and `github://...:branch`) via `skills-loader`.
- Registers loaded `SkillDefinition` metadata and instructions as tools in an ADK runner context.
- Demonstrates prompt execution with grounded skill retrieval.

---

## 2. Polyglot Bazel Developer Agent (`examples/polyglot-developer`)

Demonstrates creating a custom agentic CLI tool that uses `skills-loader` to load `skills-bazel`, `skills-go`, `skills-java`, `skills-protobuf`, `skills-python`, and `skills-frontend` to scaffold a complete polyglot Bazel monorepo.

### Directory Layout

```text
examples/polyglot-developer/
├── .env.example
├── README.md
├── main.py
├── pyproject.toml
└── tests/
    └── test_polyglot_developer.py
```

### Execution

```bash
# Run CLI tool via uv
uv run python examples/polyglot-developer/main.py --target-dir ./my-polyglot-app

# Or execute unit tests for the example
uv run python -m unittest examples/polyglot-developer/tests/test_polyglot_developer.py
```

### Key Highlights

- Leverages `SkillRegistry` and `SkillDefinition` objects to ground generated code in loaded enterprise skill rules.
- Scaffolds a complete 7-file polyglot monorepo structure:
  - `MODULE.bazel` (Hermetic Bazel bzlmod resolution from `skills-bazel`).
  - `BUILD.bazel` (Root target filegroup).
  - `proto/user.proto` (Protobuf & gRPC contracts from `skills-protobuf`).
  - `services/go-service/main.go` (Go microservice from `skills-go`).
  - `services/java-service/.../Application.java` (Java enterprise service from `skills-java`).
  - `services/python-service/main.py` (Python FastAPI service from `skills-python`).
  - `apps/web-dashboard/src/App.tsx` (React/Vite dashboard from `skills-frontend`).

---

## Contributing Standalone Examples

When contributing new examples:

1. Create a dedicated directory under `examples/<name>/`.
2. Include a standalone `pyproject.toml` with `build-system` and declared dependencies.
3. Include a `README.md` with usage instructions and example CLI options.
4. Ensure all Python entry points directly instantiate `skills_loader` or `SkillRegistry`.
