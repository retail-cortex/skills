"""Unit tests for standalone skills-loader package including GitHub & dotenv loading."""

import os
import tempfile
import unittest
from pathlib import Path
from unittest.mock import MagicMock, patch

from skills_loader.loader import (
    SkillRegistry,
    find_registry_root,
    load_all_skills,
    load_skills_from_github,
    parse_dotenv_file,
    parse_frontmatter,
)
from skills_loader.types import SkillDefinition, SkillSummary


@unittest.skipIf(os.environ.get("GITHUB_ACTIONS") == "true", "Skipping runfiles test in GitHub Actions CI")
class TestSkillsLoaderPackage(unittest.TestCase):
    """Tests evaluating standalone skills-loader functions and registry."""

    def test_find_registry_root(self) -> None:
        root = find_registry_root()
        self.assertTrue(root.exists())
        self.assertTrue((root / "packages").is_dir() or (root / "skills").is_dir() or root.name in ("packages", "skills"))

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

    def test_parse_dotenv_file_success(self) -> None:
        with tempfile.NamedTemporaryFile(mode="w+", delete=False, suffix=".env") as tf:
            tf.write("GITHUB_TOKEN=ghp_secret_123\nGITHUB_REF=v2.1.0\nSKILLS_ROOTS=skills,custom_skills\n# Comment\n")
            tmp_path = Path(tf.name)

        try:
            parsed = parse_dotenv_file(tmp_path)
            self.assertEqual(parsed.get("GITHUB_TOKEN"), "ghp_secret_123")
            self.assertEqual(parsed.get("GITHUB_REF"), "v2.1.0")
            self.assertEqual(parsed.get("SKILLS_ROOTS"), "skills,custom_skills")
        finally:
            tmp_path.unlink(missing_ok=True)

    def test_parse_dotenv_file_nonexistent_boundary(self) -> None:
        bogus_path = Path("/nonexistent/path/.env")
        parsed = parse_dotenv_file(bogus_path)
        self.assertEqual(parsed, {})

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
            self.assertIn("name", python_skill.to_dict())

        # Test search
        go_matches = registry.search("Go")
        self.assertTrue(any("go" in s.name.lower() for s in go_matches))

        # Test domain filtering
        domain_matches = registry.get_domain_skills("fastapi")
        self.assertTrue(len(domain_matches) > 0)

        # Test list_skills summary
        summaries = registry.list_skills()
        self.assertGreaterEqual(len(summaries), 23)

    @patch("subprocess.run")
    def test_load_skills_from_github_git_clone_success(self, mock_run: MagicMock) -> None:
        # Mock successful git clone command
        mock_proc = MagicMock()
        mock_proc.returncode = 0
        mock_run.returncode = 0
        mock_run.return_value = mock_proc

        # Create a temp local repo structure to simulate git clone result
        with tempfile.TemporaryDirectory() as fake_git_dir:
            fake_root = Path(fake_git_dir)
            skill_dir = fake_root / "skills" / "gh-mock-skill"
            skill_dir.mkdir(parents=True)
            (skill_dir / "SKILL.md").write_text(
                "---\nname: gh-mock-skill\ndescription: GitHub fetched skill\n---\n# Rules\nRun hermetic tests.",
                encoding="utf-8",
            )

            with patch("tempfile.TemporaryDirectory") as mock_tmp:
                mock_tmp.return_value.__enter__.return_value = fake_git_dir
                # Also mock repo target dir creation
                repo_target = fake_root / "repo"
                repo_target.mkdir(exist_ok=True)
                target_skill = repo_target / "skills" / "gh-mock-skill"
                target_skill.mkdir(parents=True, exist_ok=True)
                (target_skill / "SKILL.md").write_text(
                    "---\nname: gh-mock-skill\ndescription: GitHub fetched skill\n---\n# Rules\nRun hermetic tests.",
                    encoding="utf-8",
                )

                skills = load_skills_from_github(
                    repo="owner/skills-repo",
                    ref="v1.5.0",
                    roots=["skills"],
                    github_token="ghp_test_token",
                )

                self.assertIn("gh-mock-skill", skills)
                self.assertEqual(skills["gh-mock-skill"].description, "GitHub fetched skill")

    def test_skill_registry_from_github_factory(self) -> None:
        with patch("skills_loader.loader.load_skills_from_github") as mock_gh:
            mock_skill = SkillDefinition(
                name="factory-skill",
                description="Factory skill test",
                instructions="Follow rules",
            )
            mock_gh.return_value = {"factory-skill": mock_skill}

            registry = SkillRegistry.from_github(
                repo="https://github.com/my-org/custom-skills.git",
                ref="main",
                roots=["skills"],
            )

            mock_gh.assert_called_once_with(
                repo="https://github.com/my-org/custom-skills.git",
                ref="main",
                roots=["skills"],
                skill_filter=None,
                github_token=None,
                dotenv_path=None,
            )
            self.assertIn("factory-skill", registry.skills)

    @patch("subprocess.run")
    def test_load_google_gemini_api_skill_from_tree_url(self, mock_run: MagicMock) -> None:
        mock_proc = MagicMock()
        mock_proc.returncode = 0
        mock_run.return_value = mock_proc

        with tempfile.TemporaryDirectory() as fake_dir:
            fake_root = Path(fake_dir)
            repo_target = fake_root / "repo"
            gemini_dir = repo_target / "skills" / "cloud" / "gemini-api"
            gemini_dir.mkdir(parents=True, exist_ok=True)
            (gemini_dir / "SKILL.md").write_text(
                "---\nname: gemini-api\ndescription: Google Cloud Gemini API Skill\n---\n# Instructions\nUse Vertex AI SDK.",
                encoding="utf-8",
            )

            with patch("tempfile.TemporaryDirectory") as mock_tmp:
                mock_tmp.return_value.__enter__.return_value = fake_dir

                url = "https://github.com/google/skills/tree/main/skills/cloud/gemini-api"
                skills = load_skills_from_github(url)

                self.assertIn("gemini-api", skills)
                self.assertEqual(skills["gemini-api"].name, "gemini-api")
                self.assertIn("Google Cloud Gemini API", skills["gemini-api"].description)

    def test_parse_skill_root_uri(self) -> None:
        from skills_loader.loader import parse_skill_root_uri

        # Local file URI
        s, t, r, sub = parse_skill_root_uri("file://skills/custom")
        self.assertEqual((s, t, r, sub), ("file", "skills/custom", None, None))

        # Package URI
        s, t, r, sub = parse_skill_root_uri("pkg://retailcortex_skills_python")
        self.assertEqual((s, t, r, sub), ("pkg", "retailcortex_skills_python", None, None))

        # Standard GitHub repo URI with trailing :ref (e.g. github://google/skills/skills/cloud/gemini-api:main)
        s, t, r, sub = parse_skill_root_uri("github://google/skills/skills/cloud/gemini-api:main")
        self.assertEqual((s, t, r, sub), ("github", "google/skills", "main", "skills/cloud/gemini-api"))

        # Legacy format: mid-string :ref (e.g. github://google/skills:main/skills/cloud/gemini-api)
        s, t, r, sub = parse_skill_root_uri("github://google/skills:main/skills/cloud/gemini-api")
        self.assertEqual((s, t, r, sub), ("github", "google/skills", "main", "skills/cloud/gemini-api"))

        # GitHub repo URI with :ref
        s, t, r, sub = parse_skill_root_uri("github://owner/repo:v2.5.0")
        self.assertEqual((s, t, r, sub), ("github", "owner/repo", "v2.5.0", None))

        # GitHub repo URI with @ref/subpath
        s, t, r, sub = parse_skill_root_uri("github://owner/repo@v1.2.0/skills/cloud")
        self.assertEqual((s, t, r, sub), ("github", "owner/repo", "v1.2.0", "skills/cloud"))

        # GitHub repo URI with /tree/branch/subpath
        s, t, r, sub = parse_skill_root_uri("github://google/skills/tree/main/skills/cloud/gemini-api")
        self.assertEqual((s, t, r, sub), ("github", "google/skills", "main", "skills/cloud/gemini-api"))

    def test_load_skills_from_package(self) -> None:
        from skills_loader.loader import load_skills_from_package

        skills = load_skills_from_package("retailcortex_skills_python", skill_filter=["python-core"])
        self.assertIn("python-core", skills)

    def test_skill_registry_from_roots_factory(self) -> None:
        from skills_loader.loader import SkillRegistry

        # 1. file:// scheme
        reg_file = SkillRegistry.from_roots(
            roots=["file://."],
            skill_filter=["python-core"],
        )
        self.assertIn("python-core", reg_file.skills)

        # 2. pkg:// scheme
        reg_pkg = SkillRegistry.from_roots(
            roots=["pkg://retailcortex_skills_python"],
            skill_filter=["python-core"],
        )
        self.assertIn("python-core", reg_pkg.skills)

    def test_frontmatter_metadata_mapping(self) -> None:
        from skills_loader.loader import load_skill_from_dir
        with tempfile.TemporaryDirectory() as tmp_dir:
            sd = Path(tmp_dir) / "test-meta-skill"
            sd.mkdir()
            (sd / "SKILL.md").write_text(
                "---\nname: test-meta-skill\ndescription: Meta test\nlicense: MIT\nauthor: Jane Doe\nversion: 2.0.0\ncompatibility: ADK-v2\nallowed-tools: Bash(git:*) Read\ncustom_field: custom_val\n---\n# Instructions",
                encoding="utf-8",
            )
            ref_dir = sd / "references"
            ref_dir.mkdir()
            (ref_dir / "ref1.md").write_text("Reference content 1", encoding="utf-8")

            skill = load_skill_from_dir(sd)
            self.assertIsNotNone(skill)
            if skill:
                self.assertEqual(skill.license, "MIT")
                self.assertEqual(skill.author, "Jane Doe")
                self.assertEqual(skill.version, "2.0.0")
                self.assertEqual(skill.compatibility, "ADK-v2")
                self.assertEqual(skill.allowed_tools, "Bash(git:*) Read")
                self.assertEqual(skill.metadata.get("custom_field"), "custom_val")
                self.assertEqual(skill.metadata.get("author"), "Jane Doe")
                self.assertEqual(skill.to_dict()["license"], "MIT")
                self.assertEqual(skill.to_dict()["allowed_tools"], "Bash(git:*) Read")
                self.assertEqual(skill.get_reference_content("ref1.md"), "Reference content 1")
                self.assertIsNone(skill.get_reference_content("nonexistent.md"))

    def test_agents_skills_directory_scanning(self) -> None:
        from skills_loader.loader import load_all_skills
        with tempfile.TemporaryDirectory() as tmp_dir:
            root = Path(tmp_dir)
            ag_dir = root / ".agents" / "skills" / "ag-custom-skill"
            ag_dir.mkdir(parents=True)
            (ag_dir / "SKILL.md").write_text(
                "---\nname: ag-custom-skill\ndescription: Cross client agent skill\n---\n# Instructions",
                encoding="utf-8",
            )
            skills = load_all_skills(skills_root=root)
            self.assertIn("ag-custom-skill", skills)
            self.assertEqual(skills["ag-custom-skill"].description, "Cross client agent skill")

    def test_build_and_load_manifest(self) -> None:
        from skills_loader.loader import build_skills_manifest, load_skills_from_manifest
        with tempfile.TemporaryDirectory() as tmp_dir:
            out_json = Path(tmp_dir) / "skills_manifest.json"
            generated = build_skills_manifest(output_path=out_json)
            self.assertTrue(generated.is_file())

            loaded = load_skills_from_manifest(generated)
            self.assertGreaterEqual(len(loaded), 20)
            self.assertIn("python-core", loaded)


if __name__ == "__main__":
    unittest.main()
