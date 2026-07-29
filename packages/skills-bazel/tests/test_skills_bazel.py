import unittest
from skills_loader.loader import SkillRegistry


class TestSkillsBazelPackage(unittest.TestCase):
    """Unit test suite for skills-bazel package."""

    def test_bazel_modules_skill_loaded(self) -> None:
        registry = SkillRegistry()
        skill = registry.get("bazel-modules")
        self.assertIsNotNone(skill)
        if skill:
            self.assertEqual(skill.name, "bazel-modules")
            self.assertTrue(len(skill.instructions) > 0)


if __name__ == "__main__":
    unittest.main()
