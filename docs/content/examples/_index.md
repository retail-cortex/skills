---
title: Standalone Examples & Integrations
description: Comprehensive overview of standalone multi-language examples, native build system integrations, and Bazel test harness validation.
weight: 40
---

# Standalone Examples & Integration Demos

The `examples/` directory contains self-contained example applications and skill packages. These examples demonstrate how to integrate the Enterprise Skills Loader, build system pre-processors, and `skm` service endpoints directly into native Go, Python, and Java build pipelines as well as Google Agent Development Kit (ADK) applications.

---

## Language Examples Matrix

Each supported language ecosystem features a standalone client integration example built using its native build tool, coupled with language-specific skill collections and Bazel test wiring:

| Language | Standalone Client | Build System | Property Loader | Skill Collection | Bazel Test Target |
| :--- | :--- | :--- | :--- | :--- | :--- |
| **Go** | `examples/go/client` | Go Modules (`go.mod`) | `modenv` (`.env.toml`) | `examples/go/skills` | `//examples/go/client:test_go_client_example` |
| **Python** | `examples/python/client` | `uv` / `pyproject.toml` | `python-dotenv` (`.env`) | `examples/python/skills` | `//examples/python/client:test_python_client_example` |
| **Java** | `examples/java/client` | Maven (`pom.xml`) | System Properties | `examples/java/skills` | `//examples/java/client:src/test/java/com/company/example/ApplicationTest` |
| **Polyglot** | `examples/python/polyglot` | `uv` / `pyproject.toml` | Environment / `uv` | Cross-Language | `//examples/python/polyglot:test_polyglot_developer` |

---

## Core Integration Principles

1. **Native Build System Execution**:
   - Each client example operates as an independent project without direct Bazel runtime dependencies. Developers can navigate to `examples/<lang>/client` and run native test suites (`go test`, `python3 -m unittest`, `mvn test`).
2. **Native Property Loading**:
   - Client applications demonstrate cascading configuration resolution using idiomatic language property frameworks (`modenv` TOML cascading in Go, `.env` file loading via `python-dotenv` in Python, and Java `System.getProperty()` with environment variable fallbacks).
3. **Build Lifecycle Validation**:
   - Client integration hooks (such as the Maven pre-processor plugin `skills-loader-maven-plugin:generate-manifest`) run during native build phases (`generate-resources`) to pre-compile skills manifests before code execution.
4. **Hermetic Bazel Test Harness Wiring**:
   - Every example is wired into the root Bazel workspace build graph (`MODULE.bazel`), ensuring complete regression testing via `bazel test //...`.

---

## Detailed Example Documentation

Explore dedicated integration guides for each language ecosystem:

- 🐹 **[Go Client & Skills Examples](go.md)**: Standalone Go module integration with `modenv` property loading and Go skill collections.
- 🐍 **[Python Client, Polyglot & Skills Examples](python.md)**: `uv` and `python-dotenv` client setup, Polyglot Monorepo Scaffolding Agent, and Python skill packages.
- ☕ **[Java Client & Maven Plugin Examples](java.md)**: Maven POM configuration, `skills-loader-maven-plugin` execution, System Properties resolution, and Java skill packages.
- 📚 **[Enterprise Skills Registry Catalog](skills.md)**: Full index of 23 enterprise AI Agent Skills distributed across language skill packages and standalone markdown definitions.

---

## Directory Structure

```text
examples/
├── go/
│   ├── client/                  # Standalone Go client example (go.mod, main.go, main_test.go)
│   └── skills/                  # Go SDLC enterprise skills (src/retailcortex_skills_go)
├── python/
│   ├── client/                  # Standalone Python client example (pyproject.toml, main.py, .env)
│   ├── polyglot/                # Polyglot monorepo developer agent CLI
│   └── skills/                  # Python SDLC enterprise skills (src/retailcortex_skills_python)
├── java/
│   ├── client/                  # Standalone Java client example (pom.xml, Application.java)
│   └── skills/                  # Java SDLC enterprise skills (src/retailcortex_skills_java)
└── skills/                      # Standalone markdown skill definitions
```
