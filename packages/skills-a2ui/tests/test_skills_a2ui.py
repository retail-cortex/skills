import unittest
from skills_loader.loader import SkillRegistry


class TestSkillsA2UIPackage(unittest.TestCase):
    """Unit test suite for skills-a2ui package."""

    def test_a2ui_skill_loaded(self) -> None:
        registry = SkillRegistry()
        skill = registry.get("a2ui")
        self.assertIsNotNone(skill)
        if skill:
            self.assertEqual(skill.name, "a2ui")
            self.assertTrue(len(skill.instructions) > 0)


if __name__ == "__main__":
    unittest.main()
