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
        if (runfiles_root / "packages").is_dir() or (runfiles_root / "skills").is_dir():
            return runfiles_root

    current: Path = Path(__file__).resolve().parent
    for parent in [current] + list(current.parents):
        if (parent / "packages").is_dir() or (parent / "skills").is_dir():
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
        target_skill: Path = registry_root / "examples" / "skills" / "python" / "python-core"
        if not target_skill.exists():
            target_skill = registry_root / "examples" / "skills-python" / "src" / "retailcortex_skills_python" / "skills" / "python-core"
        result: SkillAuditResult = audit_skill_directory(target_skill)
        self.assertTrue(result.frontmatter_valid)
        self.assertTrue(result.l3_tree_valid)
        self.assertTrue(result.cwe_security_valid)
        self.assertTrue(result.rate_limit_429_valid)
        self.assertTrue(result.clickable_links_valid)
        self.assertTrue(result.passed)

    def test_audit_all_skills_suite(self) -> None:
        registry_root: Path = find_registry_root()
        summary: AuditSummary = audit_all_skills(registry_root / "examples" / "skills")
        if summary.failed_skills > 0:
            print(f"Audit failures: {[(r.skill_name, r.errors) for r in summary.results if not r.passed]}")
        self.assertGreaterEqual(summary.total_skills, 20)
        self.assertEqual(summary.passed_skills, summary.total_skills)
        self.assertEqual(summary.failed_skills, 0)

    def test_schema_to_json(self) -> None:
        summary = AuditSummary(
            total_skills=1,
            passed_skills=1,
            failed_skills=0,
            results=[SkillAuditResult(
                skill_name="test",
                directory_path="/path",
                frontmatter_valid=True,
                l3_tree_valid=True,
                cwe_security_valid=True,
                rate_limit_429_valid=True,
                clickable_links_valid=True,
            )]
        )
        data = summary.to_dict()
        self.assertEqual(data["total_skills"], 1)
        self.assertEqual(data["results"][0]["skill_name"], "test")

    def test_audit_directory_failures(self) -> None:
        import tempfile
        with tempfile.TemporaryDirectory() as tmp_dir:
            tmp_path = Path(tmp_dir)
            # Missing SKILL.md
            res = audit_skill_directory(tmp_path)
            self.assertFalse(res.passed)
            self.assertFalse(res.frontmatter_valid)

            # Invalid frontmatter
            skill_dir = tmp_path / "bad-skill"
            skill_dir.mkdir()
            (skill_dir / "references").mkdir()
            (skill_dir / "examples").mkdir()
            (skill_dir / "references" / "ref.md").write_text("ref")
            (skill_dir / "examples" / "ex.md").write_text("ex")
            (skill_dir / "SKILL.md").write_text("---\nname: INVALID_NAME\n---\nNo checkpoints")

            res_bad = audit_skill_directory(skill_dir)
            self.assertFalse(res_bad.passed)

    def test_cli_execution(self) -> None:
        from validator.cli import main
        from io import StringIO
        from unittest.mock import patch

        registry_root: Path = find_registry_root()
        target_skill = str(registry_root / "skills" / "python" / "python-core")

        with patch("sys.argv", ["skm-audit", target_skill]), patch("sys.exit"):
            main()

        with patch("sys.argv", ["skm-audit", "-r", "--json", str(registry_root / "skills")]), patch("sys.exit"):
            out = StringIO()
            with patch("sys.stdout", out):
                main()
                self.assertIn("Total Skills:", out.getvalue())

if __name__ == "__main__":
    unittest.main()
