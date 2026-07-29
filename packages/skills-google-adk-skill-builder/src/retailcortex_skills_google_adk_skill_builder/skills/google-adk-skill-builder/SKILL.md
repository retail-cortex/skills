---
name: google-adk-skill-builder
description: Meta-skill factory SDLC for Google ADK. Covers ADK evaluator TDD, HTTP 429 rate limit backoff, GitHub Actions schema validation, human-in-the-loop security reviews (CWE-94), and SemVer.
license: Apache-2.0
metadata:
  author: Ryan McGuinness
  version: "1.0"
---

# ADK Meta-Skill Factory SDLC Skill

This skill follows the **Skill Factory Pattern** from the Google ADK Progressive Disclosure architecture, enforcing full SDLC rigor (TDD with ADK evaluators, CI/CD validation, Human-in-the-loop security, HTTP 429 resilience, and SemVer).

## Prerequisites & Pre-Flight Checklist

1. Google ADK (`google-adk>=1.16.0`) installed.
2. `agentskills.io` universal specification loaded in `references/`.

## HTTP 429 Rate Limit & Quota Backoff Invariants

- Skill generation engines querying LLMs must implement exponential backoff with full randomized jitter to survive HTTP 429 quota exhaustion.

## Security Checkpoints & CWE Invariants

- **CWE-94 (Code Injection)**: Sanitize all text instructions to prevent prompt injection and unauthorized privilege escalation in AI agents.
- **CWE-73 (External Control of File Name or Path)**: Enforce strict kebab-case naming (<= 64 chars) to prevent path traversal when writing new `SKILL.md` files.
- **Human-in-the-Loop Gate**: Never deploy generated skills to production registries without explicit human verification.

## 3-Phase Execution Protocol

### Phase 1: Read Spec & Determine Requirements
Identify skill name (kebab-case), target domain task, and required L3 references.

### Phase 2: Generate Spec-Compliant Skill & Run Evaluator TDD
Generate `SKILL.md` with valid YAML frontmatter and run ADK evaluation suites:
```bash
uv run python -m adk.evaluate --skill-dir skills/my-new-skill
```

### Phase 3: Human Review & SemVer Catalog Registration
Verify skill safety and register in the central skill registry.

## Progressive Disclosure References

- **Agent Skills Specification**: Read [`references/agentskills_spec.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/adk-skill-factory/references/agentskills_spec.md).
- **Reference Generator**: View [`examples/skill_generator.py`](file:///Users/rmcguinness/Projects/skill-builder/skills/adk-skill-factory/examples/skill_generator.py).
