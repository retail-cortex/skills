---
name: multi-content-suite
description: Comprehensive multi-content validation skill testing binary encoding across PNG, JPEG, WebP, GIF, PDF, WASM, WAV audio, ZIP archives, SQLite databases, and Protobuf binary streams.
license: Apache-2.0
metadata:
  author: Ryan McGuinness
  version: "1.0.0"
authors:
  - name: Retail Cortex Engineering
    url: https://github.com/retail-cortex/skills
category: testing
tags:
  - testing
  - binary
  - images
  - pdf
  - wasm
  - audio
  - zip
  - sqlite
  - protobuf
trigger_phrases:
  - "test all binary formats"
  - "validate multi-content binary skill"
  - "test audio wasm pdf and images"
execution_hints:
  preferred_model: "gemini-2.0-flash"
  requires_human_approval: false
  environment_variables: []
  timeout_seconds: 120
---

# Multi-Content Suite Skill (Complete Binary & Text Matrix)

This skill provides comprehensive verification of the full spectrum of binary formats, raster/vector graphics, document formats, archive packages, compiled WASM modules, audio waveforms, and structured text definitions.

## Visual & Multimedia Binary Assets

### Raster Images (PNG, JPEG, WebP, GIF)
- ![PNG Canvas](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/references/canvas.png)
- ![JPEG Photo](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/references/photo.jpg)
- ![WebP Graphic](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/references/graphic.webp)
- ![GIF Animation](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/references/animation.gif)

### Vector Graphics (SVG)
- ![Architecture SVG](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/references/architecture_diagram.svg)

## Binary Documents, Modules & Archives
- [Whitepaper PDF Document](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/references/whitepaper.pdf)
- [WebAssembly WASM Module](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/references/calculator.wasm)
- [Audio WAV Chime](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/references/chime.wav)
- [ZIP Archive Package](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/references/archive.zip)
- [SQLite Database Fixture](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/references/fixture.sqlite)
- [Protobuf Binary Descriptor](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/references/descriptor.pb)

## Structured Text & Schemas
- [Specification Guide](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/references/specification_guide.md)
- [API Contract JSON](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/references/api_contract.json)
- [Deployment YAML](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/references/deployment_config.yaml)
- [Tabular Dataset CSV](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/references/raw_dataset.csv)
- [Schema DDL SQL](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/references/schema_ddl.sql)

## Security Checkpoints & CWE Invariants

- **CWE-20 (Improper Input Validation)**: Decoders must check binary magic bytes (PDF `%PDF-`, PNG `PNG`, WASM `asm`, ZIP `PK`, SQLite `SQLite format 3`).
- **CWE-798 (Hardcoded Credentials)**: All secret configurations must be decoded dynamically via `modenv` XOR cipher at runtime.
- **CWE-89 (SQL Injection)**: Database queries must execute through parameterized ORM or prepared statements.

## HTTP 429 Rate Limit & Resilience

All asset transmission channels must enforce token-bucket throttling and exponential backoff retry algorithms when encountering HTTP 429 rate limit responses.

## Polyglot Reference Implementation Files

- [Go Server Implementation](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/examples/service.go)
- [Python Agent Pipeline](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/examples/agent_pipeline.py)
- [React TypeScript Component](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/examples/Component.tsx)
- [Java Application](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/examples/Application.java)
- [Terraform Infrastructure](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/examples/infrastructure.tf)
- [Container Dockerfile](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/examples/Dockerfile)
- [Configuration TOML](file:///Users/rmcguinness/Projects/skill-builder/examples/skills/testing/multi-content-suite/examples/config.toml)
