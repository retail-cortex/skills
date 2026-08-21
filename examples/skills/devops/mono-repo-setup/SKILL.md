---
name: mono-repo-setup
description: Elite meta-skill for scaffolding enterprise polyglot monorepos using Bazel 8/9 Bzlmod. Configures multi-language hermetic toolchains, GCP Terraform IaC, cascading TOML, HTTP 429 rate limit backoff, and 3-phase SDLC verification.
license: Apache-2.0
metadata:
  author: Ryan McGuinness
  version: "1.0"
authors:
  - name: Retail Cortex Engineering
    url: https://github.com/retail-cortex/castor
category: devops
tags:
  - monorepo
  - bazel
  - polyglot
  - scaffolding
trigger_phrases:
  - "scaffold Bazel monorepo"
  - "setup polyglot monorepo"
  - "Bzlmod root setup"
execution_hints:
  preferred_model: "gemini-3.1-pro"
  requires_human_approval: false
  environment_variables: []
  timeout_seconds: 300
---
# Monorepo Setup Meta-Skill (Bazel Bzlmod Standard)

This meta-skill provides automated, security-hardened scaffolding instructions and templates for establishing an enterprise **Polyglot Monorepo** powered by **Bazel 8/9 (Bzlmod)**.

## Prerequisites & Pre-Flight Checklist

1. Bazelisk or Bazel 8.0.0+ installed.
2. Git initialized in the repository root.
3. Access to Google Cloud SDK (`gcloud`) for Terraform remote state backend provisioning.

## Security Checkpoints & CWE Invariants

- **CWE-829 (Unverified External Imports)**: All dependencies MUST be locked via `MODULE.bazel.lock` and `pnpm-lock.yaml` / `uv.lock`. Direct unpinned downloads are strictly prohibited.
- **CWE-798 (Hardcoded Credentials)**: All secret configurations in `.env.toml` MUST use in-memory XOR encryption (`xor:...`) and local files (`.env.local.toml`) MUST be added to `.gitignore`.

## HTTP 429 Rate Limit & Backoff Invariants

- Outbound API calls across polyglot microservices must implement exponential backoff with full randomized jitter to prevent HTTP 429 quota exhaustion.

## 3-Phase Execution Protocol

### Phase 1: Validate & Initialize Directory Layout
```bash
mkdir -p apps libs api/proto configs/terraform build/ci .github/workflows docs
echo "9.0.0" > .bazelversion
```

### Phase 2: Configure Hermetic Bzlmod & Toolchains
Populate `MODULE.bazel` with hermetic toolchain definitions (Go 1.26+, Python 3.13 uv, Java 25 Maven, React 19 pnpm).

### Phase 3: Verify, TDD Test & Publish Docs
```bash
bazel run //:gazelle
bazel test //... --test_output=errors
bazel coverage //... --combined_report=lcov
bazel build //docs:hugo_site
```

## Progressive Disclosure References

- **Monorepo Scaffold Guide**: Read [`references/monorepo_scaffold_guide.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/mono-repo-setup/references/monorepo_scaffold_guide.md).
- **Reference Module File**: View [`examples/MODULE.bazel`](file:///Users/rmcguinness/Projects/skill-builder/skills/mono-repo-setup/examples/MODULE.bazel).
- **Reference Root Build**: View [`examples/BUILD.bazel`](file:///Users/rmcguinness/Projects/skill-builder/skills/mono-repo-setup/examples/BUILD.bazel).
- **Reference Terraform Config**: View [`examples/configs/terraform/main.tf`](file:///Users/rmcguinness/Projects/skill-builder/skills/mono-repo-setup/examples/configs/terraform/main.tf).
- **Reference Base Config**: View [`examples/configs/.env.toml`](file:///Users/rmcguinness/Projects/skill-builder/skills/mono-repo-setup/examples/configs/.env.toml).
