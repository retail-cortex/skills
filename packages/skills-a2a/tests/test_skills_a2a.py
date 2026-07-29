import unittest
from skills_loader.loader import SkillRegistry


class TestSkillsA2APackage(unittest.TestCase):
    """Unit test suite for skills-a2a package."""

    def test_a2a_skill_loaded(self) -> None:
        registry = SkillRegistry()
        skill = registry.get("a2a")
        self.assertIsNotNone(skill)
        if skill:
            self.assertEqual(skill.name, "a2a")
            self.assertTrue(len(skill.instructions) > 0)


if __name__ == "__main__":
    unittest.main()
