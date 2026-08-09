# Copyright 2026 Ryan McGuinness
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""End-to-end integration test for enterprise skills loader.

Verifies loading multiple skills, all skills from one package,
GitHub skill loading, directory structure verification,
and compatibility with uv and Bazel runners.
"""

import os
import unittest
from pathlib import Path

try:
    from tests.e2e.fixtures import (
        create_temp_manifest_dir,
        get_expected_package_skills,
        get_github_test_uri_spec,
        get_multi_skill_file_roots,
        get_workspace_root,
        mock_github_skill_environment,
    )
except ImportError:
    from fixtures import (
        create_temp_manifest_dir,
        get_expected_package_skills,
        get_github_test_uri_spec,
        get_multi_skill_file_roots,
        get_workspace_root,
        mock_github_skill_environment,
    )
from loader.loader import (
    SkillRegistry,
    build_skills_manifest,
    get_loader_skills_dir,
    load_all_skills,
    load_skills_from_github,
    load_skills_from_manifest,
    load_skills_from_package,
    parse_frontmatter,
    parse_skill_root_uri,
)


class TestSkillsE2E(unittest.TestCase):
    """End-to-end test suite for skill-builder dynamic skill loader."""

    def setUp(self) -> None:
        self.root = get_workspace_root()

    def test_registry_root_discovery(self) -> None:
        """Verify workspace root discovery under Bazel and uv execution."""
        self.assertTrue(self.root.exists(), f"Registry root does not exist: {self.root}")
        self.assertTrue(
            (self.root / "clients").is_dir() or (self.root / "apps").is_dir() or (self.root / "pkg").is_dir() or (self.root / "skills").is_dir() or self.root.name in ("clients", "apps", "pkg", "skills"),
            f"Invalid registry root structure at {self.root}",
        )

    def test_load_multiple_skills_from_roots(self) -> None:
        """Verify loading multiple distinct skills across different packages using file URIs."""
        roots = get_multi_skill_file_roots(self.root)
        registry = SkillRegistry.from_roots(roots=roots)
        loaded_skills = registry.skills

        self.assertGreaterEqual(len(loaded_skills), 3)
        self.assertIn("python-core", loaded_skills)
        self.assertIn("go-lang", loaded_skills)
        self.assertIn("bazel-modules", loaded_skills)

        # Verify individual skill properties
        py_skill = registry.get("python-core")
        self.assertIsNotNone(py_skill)
        if py_skill:
            self.assertEqual(py_skill.name, "python-core")
            self.assertTrue(len(py_skill.instructions) > 0)

        go_skill = registry.get("go-lang")
        self.assertIsNotNone(go_skill)
        if go_skill:
            self.assertEqual(go_skill.name, "go-lang")

        bazel_skill = registry.get("bazel-modules")
        self.assertIsNotNone(bazel_skill)
        if bazel_skill:
            self.assertEqual(bazel_skill.name, "bazel-modules")

    def test_load_all_skills_from_one_package(self) -> None:
        """Verify loading all skills contained within a single python package (skills-python)."""
        pkg_name = "retailcortex_skills_python"
        skills = load_skills_from_package(pkg_name)
        expected_python_skills = get_expected_package_skills(pkg_name)

        for expected_name in expected_python_skills:
            self.assertIn(expected_name, skills, f"Skill '{expected_name}' missing from '{pkg_name}'")

        pkg_registry = SkillRegistry.from_roots(roots=[f"pkg://{pkg_name}"])
        self.assertGreaterEqual(len(pkg_registry.skills), len(expected_python_skills))
        for expected_name in expected_python_skills:
            self.assertIn(expected_name, pkg_registry.skills)

    def test_load_github_skill(self) -> None:
        """Verify parsing and loading GitHub skills via github:// URI scheme and loader API."""
        uri, exp_scheme, exp_target, exp_ref, exp_subpath = get_github_test_uri_spec()
        scheme, target, ref, subpath = parse_skill_root_uri(uri)
        self.assertEqual((scheme, target, ref, subpath), (exp_scheme, exp_target, exp_ref, exp_subpath))

        with mock_github_skill_environment() as (repo, git_ref, roots):
            skills = load_skills_from_github(repo=repo, ref=git_ref, roots=roots)
            self.assertIn("gemini-api", skills)
            gh_skill = skills["gemini-api"]
            self.assertEqual(gh_skill.name, "gemini-api")
            self.assertEqual(gh_skill.author, "Google DeepMind")
            self.assertIn("api_reference.md", gh_skill.references)

    def test_directory_structure_verification(self) -> None:
        """Verify file and folder directory structure for all workspace skills."""
        all_skills = load_all_skills(self.root)
        self.assertGreater(len(all_skills), 0, "No skills found in workspace root")

        for skill_name, skill in all_skills.items():
            skill_dir = Path(skill.path)
            self.assertTrue(skill_dir.is_dir(), f"Skill '{skill_name}' path is not a directory: {skill_dir}")

            skill_md = skill_dir / "SKILL.md"
            self.assertTrue(skill_md.is_file(), f"Skill '{skill_name}' missing SKILL.md")

            content = skill_md.read_text(encoding="utf-8")
            _, body = parse_frontmatter(content)
            self.assertTrue(len(body) > 0, f"Skill '{skill_name}' body is empty")

            ref_dir = skill_dir / "references"
            if ref_dir.is_dir():
                ref_files = list(ref_dir.glob("*.md"))
                self.assertEqual(len(ref_files), len(skill.references))
                for ref_file in ref_files:
                    self.assertIn(ref_file.name, skill.references)

            ex_dir = skill_dir / "examples"
            if ex_dir.is_dir():
                ex_files = [f for f in ex_dir.iterdir() if f.is_file() and not f.name.startswith(".")]
                self.assertEqual(len(ex_files), len(skill.examples))
                for ex_file in ex_files:
                    self.assertIn(ex_file.name, skill.examples)

        loader_dir = get_loader_skills_dir()
        self.assertTrue(loader_dir.exists() and loader_dir.is_dir())

        with create_temp_manifest_dir() as manifest_file:
            generated_path = build_skills_manifest(skills_root=self.root, output_path=manifest_file)
            self.assertTrue(generated_path.is_file())

            manifest_skills = load_skills_from_manifest(generated_path)
            self.assertEqual(len(manifest_skills), len(all_skills))
            self.assertIn("python-core", manifest_skills)

    def test_uv_and_bazel_environment_integration(self) -> None:
        """Verify seamless execution integration under both uv and Bazel test environments."""
        is_bazel = "TEST_SRCDIR" in os.environ or "BUILD_WORKSPACE_DIRECTORY" in os.environ
        registry = SkillRegistry()
        self.assertGreaterEqual(
            len(registry.skills),
            20,
            f"SkillRegistry should load all workspace skills under {'Bazel' if is_bazel else 'uv/Python'}",
        )

        summaries = registry.list_skills()
        self.assertGreaterEqual(len(summaries), 20)

        python_skills = registry.get_domain_skills("python")
        self.assertTrue(len(python_skills) > 0)


if __name__ == "__main__":
    unittest.main()
