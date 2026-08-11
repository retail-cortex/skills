---
title: "Scaffolding & SDLC Audit"
weight: 50
---

# Scaffolding & Quality Audit

`cstr` provides built-in capabilities for rapid skill development, scaffolding, and automated 5-point Enterprise SDLC quality auditing.

---

## 1. Scaffolding New Skills (`cstr init`)

Scaffold a new valid skill directory structure satisfying all SDLC invariants immediately:

```bash
# Scaffold in custom directory
cstr init my-custom-skill -d ./skills
```

This generates:
```
skills/my-custom-skill/
├── SKILL.md
├── references/
│   └── README.md
└── examples/
    └── README.md
```

### Generated `SKILL.md` Template
```markdown
---
name: my-custom-skill
description: Comprehensive enterprise skill guidelines for my-custom-skill.
license: Apache-2.0
version: 1.0.0
authors:
  - name: Enterprise Engineering Team
    email: engineering@company.com
category: enterprise
tags:
  - my-custom-skill
trigger_phrases:
  - "run my-custom-skill"
execution_hints:
  preferred_model: "gemini-2.0-flash"
---

# My Custom Skill Guidelines

## Overview
Detailed instructions for AI Agent execution...
```

---

## 2. 5-Point SDLC Quality Audit (`cstr validate`)

Run automated compliance checks against YAML frontmatter, directory hierarchies, CWE security rules, rate-limit resilience, and Markdown file links:

```bash
# Validate single skill directory
cstr validate ./skills/my-custom-skill

# Recursively audit all skills in directory
cstr validate -r ./skills

# Machine-readable JSON output for CI/CD pipelines
cstr validate -r ./skills --json
```

---

## 3. The 5 SDLC Audit Rules

| Rule | Category | Enforcement Details |
| :--- | :--- | :--- |
| **Rule 1** | **YAML Frontmatter** | Requires non-empty `name`, `description`, `version`, `license`, and `authors`. Name must conform to `^[a-z0-9-]+$`. |
| **Rule 2** | **Directory Structure** | Requires `references/` and `examples/` subdirectories with documentation. |
| **Rule 3** | **Security Rules (CWE)** | Checks for explicit security invariants preventing prompt injection (CWE-79/116), command injection (CWE-78), and unescaped shell commands. |
| **Rule 4** | **Resilience & Rate Limits** | Requires explicit instructions regarding HTTP 429 backoff handling, exponential jitter retry logic, or concurrency limits. |
| **Rule 5** | **File Link Integrity** | Verifies all Markdown links referencing local files (`file://` or relative paths) resolve to valid, existing targets. |

### Audit Exit Codes
* **`0`**: All skills passed 100% compliance checks.
* **`1`**: One or more validation violations detected (blocks CI/CD pipelines and `cstr register`).

