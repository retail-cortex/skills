---
title: "Scaffolding & SDLC Audit"
weight: 50
---

# Scaffolding & Quality Audit

`skm` provides built-in tools for local skill development, scaffolding, and 5-point Enterprise SDLC quality auditing.

---

## Scaffolding New Skills (`skm init`)

Scaffold a new valid skill directory structure satisfying all SDLC invariants immediately:

```bash
skm init my-custom-skill -d ./skills
```

This creates:
```
my-custom-skill/
├── SKILL.md
├── references/
│   └── README.md
└── examples/
    └── README.md
```

---

## 5-Point SDLC Quality Audit (`skm validate`)

Run automated compliance checks against YAML frontmatter, CWE security rules, rate-limit resilience, and file links:

```bash
# Validate single skill
skm validate ./skills/my-skill

# Recursively audit all skills in directory and export JSON
skm validate -r ./packages --json
```

### Audit Verification Rules

1. **Frontmatter Invariants**: Validates `name`, `description`, `version`, `author`, `license`.
2. **Directory Structure**: Requires `references/` and `examples/` subdirectories.
3. **Security Rules (CWE)**: Checks for prompt injection and input sanitization safeguards.
4. **Resilience & Rate-Limiting**: Ensures HTTP 429 backoff/retry handling rules exist.
5. **File Link Integrity**: Verifies all Markdown `file://` link references point to valid local targets.
