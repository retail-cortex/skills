---
name: nx-monorepo
description: Elite Polyglot Monorepo SDLC using NX. Covers affected TDD testing, GitHub Actions CI with remote caching, HTTP 429 cache backoff, and module boundary security rules (CWE-668).
---

# NX Polyglot Monorepo SDLC Skill

This skill prescribes best practices for managing enterprise polyglot repositories using **NX** orchestration alongside Python **uv** workspaces, TypeScript/React applications, Go microservices, and Bazel integration.

## Prerequisites & Pre-Flight Checklist

1. Node.js 22+ and `pnpm` installed.
2. NX CLI installed globally or via `npx nx`.

## HTTP 429 Rate Limit & Remote Cache Backoff

- Configure NX Cloud and GCS remote cache upload/download actions with exponential backoff retries to survive HTTP 429 quota spikes during large parallel CI runs.

## Security Checkpoints & CWE Invariants

- **CWE-668 (Exposure of Resource to Wrong Sphere)**: Enforce architectural boundaries between libraries and applications using `@nx/enforce-module-boundaries` ESLint rules.
- **Isolation Invariant**: Private domain libraries in `libs/` must never be directly imported by untrusted external client apps.

## 3-Phase Execution Protocol

### Phase 1: Dependency Graph Analysis
Inspect project relationships and define workspace tags in `project.json`.

### Phase 2: Affected TDD Suite & Remote Caching
Run unit and integration tests only on projects affected by changes:
```bash
npx nx affected -t lint test build --base=origin/main
```

### Phase 3: Version & Release via SemVer
```bash
npx nx release --specifier patch
```

## Progressive Disclosure References

- **NX Polyglot Guide**: Read [`references/nx_polyglot.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/nx-monorepo/references/nx_polyglot.md).
- **Reference NX Configuration**: View [`examples/nx.json`](file:///Users/rmcguinness/Projects/skill-builder/skills/nx-monorepo/examples/nx.json).
