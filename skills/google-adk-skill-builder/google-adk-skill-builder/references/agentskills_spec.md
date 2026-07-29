# Agent Skills Specification (agentskills.io)

## Universal Directory Format

Every agent skill directory must follow the universal layout:

```
skills/<skill-name>/
├── SKILL.md           # L2: Instructions + L1 Frontmatter
├── references/        # L3: Domain knowledge loaded on demand
├── examples/          # L3: Concrete code and schema samples
└── scripts/           # L3: Helper CLI tools
```

## SKILL.md Structure

```markdown
---
name: your-skill-name
description: A clear, descriptive summary (<1024 chars) explaining what the skill does and when to activate it.
---

# Your Skill Title

Detailed step-by-step workflow instructions (<5,000 tokens).

## References
Point to specific files in `references/` for deep domain knowledge.
```

## Token Efficiency Rules

- Monolithic System Prompts: 10,000+ tokens on every LLM call.
- Progressive Disclosure: ~100 tokens (L1 metadata) on startup. L2 and L3 are loaded only when the agent explicitly activates the skill via `load_skill` or `load_skill_resource`.
