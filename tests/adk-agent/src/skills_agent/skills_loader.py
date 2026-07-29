"""Backward-compatible re-export module delegating to the standalone skills-loader package."""

from skills_loader import (
    SkillDefinition,
    SkillRegistry,
    SkillSummary,
    find_registry_root,
    load_all_skills,
    load_skill_from_dir,
    load_skills_from_github,
    load_skills_from_roots,
    parse_dotenv_file,
    parse_frontmatter,
    parse_skill_root_uri,
)

__all__ = [
    "SkillDefinition",
    "SkillSummary",
    "SkillRegistry",
    "find_registry_root",
    "parse_frontmatter",
    "parse_dotenv_file",
    "parse_skill_root_uri",
    "load_skill_from_dir",
    "load_all_skills",
    "load_skills_from_github",
    "load_skills_from_roots",
]
