import os
import sys
import unittest
from pathlib import Path

# Add src to path
sys.path.insert(0, str(Path(__file__).parents[1] / "src"))

from validator.schema import SkillFrontmatter, SkillAuditResult, AuditSummary
from validator.audit import parse_simple_frontmatter, audit_skill_directory, audit_all_skills

def find_registry_root() -> Path:
    if "BUILD_WORKSPACE_DIRECTORY" in os.environ:
        return Path(os.environ["BUILD_WORKSPACE_DIRECTORY"])
    if "TEST_SRCDIR" in os.environ and "TEST_WORKSPACE" in os.environ:
        runfiles_root = Path(os.environ["TEST_SRCDIR"]) / os.environ["TEST_WORKSPACE"]
        if (runfiles_root / "skills").is_dir():
            return runfiles_root

    current: Path = Path(__file__).resolve().parent
    for parent in [current] + list(current.parents):
        if (parent / "skills").is_dir():
            return parent
    return Path(__file__).parents[3]

class TestSkillValidator(unittest.TestCase):

    def test_frontmatter_validation_success(self) -> None:
        fm: SkillFrontmatter = SkillFrontmatter(
            name="test-skill",
            description="A valid test skill description for LLM planners."
        )
        self.assertEqual(fm.name, "test-skill")
        self.assertTrue(len(fm.description) > 0)

    def test_frontmatter_validation_invalid_name(self) -> None:
        with self.assertRaises(ValueError):
            SkillFrontmatter(
                name="Invalid_Name_Skill",
                description="Testing invalid kebab-case name"
            )

    def test_parse_frontmatter_extraction(self) -> None:
        content: str = """---
name: sample-skill
description: Sample description
---
# Sample Body Content
"""
        data, body = parse_simple_frontmatter(content)
        self.assertEqual(data["name"], "sample-skill")
        self.assertIn("Sample Body Content", body)

    def test_audit_live_skill_directory(self) -> None:
        registry_root: Path = find_registry_root()
        target_skill: Path = registry_root / "skills" / "python-core"
        if target_skill.exists():
            result: SkillAuditResult = audit_skill_directory(target_skill)
            self.assertTrue(result.frontmatter_valid)
            self.assertTrue(result.l3_tree_valid)
            self.assertTrue(result.cwe_security_valid)
            self.assertTrue(result.rate_limit_429_valid)
            self.assertTrue(result.clickable_links_valid)
            self.assertTrue(result.passed)

    def test_audit_all_skills_suite(self) -> None:
        registry_root: Path = find_registry_root()
        summary: AuditSummary = audit_all_skills(registry_root)
        self.assertEqual(summary.total_skills, 23)
        self.assertEqual(summary.passed_skills, 23)
        self.assertEqual(summary.failed_skills, 0)

if __name__ == "__main__":
    unittest.main()
