---
name: configuration-modenv
description: Elite hierarchical TOML configuration SDLC. Covers TDD unit testing for cascading precedence, HTTP 429 backoff settings, GitHub Actions CI schema validation, and XOR cipher security (CWE-798).
license: Apache-2.0
metadata:
  author: Ryan McGuinness
  version: "1.0"
---

# Hierarchical TOML Configuration SDLC Skill (modenv)

This skill prescribes best practices for application configuration across all programming languages using **hierarchical TOML cascading configuration** and the **modenv** architecture.

## Prerequisites & Pre-Flight Checklist

1. `modenv` Go library (`github.com/rrmcguinness/modenv`) or Python TOML parser installed.
2. `MODENV_KEY` environment variable generated for XOR secret decryption.

## HTTP 429 Rate Limit Configuration Invariants

- Centralize rate limiting thresholds, token bucket sizes, and exponential backoff parameters in `.env.toml` to dynamically adjust to HTTP 429 quotas across environments.

## Security Checkpoints & CWE Invariants

- **CWE-798 (Hardcoded Credentials)**: Passwords and secrets starting with `xor:` MUST be decrypted strictly in memory at load time. Unencrypted plain-text secrets in git are strictly prohibited.
- **Defensive Copying**: Always return cloned structs to prevent pointer corruption across concurrent agent tasks.

## 3-Phase Execution Protocol

### Phase 1: Initialize TOML Hierarchy
Create `.env.toml` (base defaults), `.env.${RUNTIME}.toml` (staging/prod), and `.env.local.toml` (local developer overrides, gitignored).

### Phase 2: Run Configuration TDD Suite
Write unit tests asserting cascading precedence (`.env.local.toml` overrides `.env.dev.toml` overrides `.env.toml`).

### Phase 3: Validate Resolved Config & Decrypt XOR Secrets
```bash
go test -v ./pkg/modenv/...
MODENV_RUNTIME=production modenv read
```

## Progressive Disclosure References

- **Modenv Specification**: Read [`references/modenv_spec.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/configuration-modenv/references/modenv_spec.md).
- **Reference TOML Config**: View [`examples/.env.toml`](file:///Users/rmcguinness/Projects/skill-builder/skills/configuration-modenv/examples/.env.toml).
