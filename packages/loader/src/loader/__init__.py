"""Standalone reusable skill scanner and loader package for Google ADK agents."""

from loader.loader import (
    SkillRegistry,
    build_skills_manifest,
    find_registry_root,
    load_all_skills,
    load_skill_from_dir,
    load_skills_from_entry_points,
    load_skills_from_github,
    load_skills_from_manifest,
    load_skills_from_package,
    load_skills_from_roots,
    parse_dotenv_file,
    parse_frontmatter,
    parse_skill_root_uri,
)
from loader.types import SkillDefinition, SkillSummary

__all__ = [
    "SkillDefinition",
    "SkillSummary",
    "SkillRegistry",
    "find_registry_root",
    "get_loader_skills_dir",
    "parse_frontmatter",
    "parse_dotenv_file",
    "parse_skill_root_uri",
    "load_skill_from_dir",
    "load_all_skills",
    "load_skills_from_entry_points",
    "load_skills_from_github",
    "load_skills_from_package",
    "load_skills_from_roots",
    "build_skills_manifest",
    "load_skills_from_manifest",
]
