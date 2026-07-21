"""Unit tests for skills loader and registry."""

import unittest
from pathlib import Path

from skills_agent.skills_loader import (
    SkillRegistry,
    find_registry_root,
    load_all_skills,
    parse_frontmatter,
)


class TestSkillsLoader(unittest.TestCase):
    """Tests for discovering and loading skills from the repository."""

    def test_find_registry_root(self) -> None:
        root = find_registry_root()
        self.assertTrue(root.exists())
        self.assertTrue((root / "skills").is_dir() or root.name == "skills")

    def test_parse_frontmatter(self) -> None:
        sample = "---\nname: test-skill\ndescription: A test description\n---\n# Instructions\nFollow these rules."
        fm, body = parse_frontmatter(sample)
        self.assertEqual(fm["name"], "test-skill")
        self.assertEqual(fm["description"], "A test description")
        self.assertIn("# Instructions", body)

    def test_parse_frontmatter_empty(self) -> None:
        sample = "# No Frontmatter\nJust body content."
        fm, body = parse_frontmatter(sample)
        self.assertEqual(fm, {})
        self.assertEqual(body, sample)

    def test_load_all_skills(self) -> None:
        skills = load_all_skills()
        self.assertGreaterEqual(len(skills), 23)
        self.assertIn("python-core", skills)
        self.assertIn("go-lang", skills)
        self.assertIn("python-adk-fastapi", skills)

    def test_skill_registry_queries(self) -> None:
        registry = SkillRegistry()
        self.assertGreaterEqual(len(registry.skills), 23)

        python_skill = registry.get("python-core")
        self.assertIsNotNone(python_skill)
        if python_skill:
            self.assertEqual(python_skill.name, "python-core")
            self.assertTrue(len(python_skill.instructions) > 0)

        # Test search
        go_matches = registry.search("Go")
        self.assertTrue(any("go" in s.name.lower() for s in go_matches))

        # Test domain filtering
        domain_matches = registry.get_domain_skills("fastapi")
        self.assertTrue(len(domain_matches) > 0)

        # Test list_skills summary
        summaries = registry.list_skills()
        self.assertGreaterEqual(len(summaries), 23)


if __name__ == "__main__":
    unittest.main()
