"""Unit tests for the PEP 517 build wrapper and dependency downloader."""

import os
import shutil
import tempfile
import tomllib
import unittest
from pathlib import Path
from unittest.mock import patch, MagicMock

from loader.build_meta import download_build_dependencies


class TestBuildMeta(unittest.TestCase):

    def setUp(self):
        self.temp_dir = tempfile.mkdtemp()
        self.workspace_dir = Path(self.temp_dir)
        
        # Create a mock skill directory to simulate a downloaded skill
        self.mock_skill_source = self.workspace_dir / "mock_skills" / "sample_skill"
        self.mock_skill_source.mkdir(parents=True)
        (self.mock_skill_source / "SKILL.md").write_text(
            "---\nname: sample_skill\ndescription: Test skill\n---\nBody", encoding="utf-8"
        )
        
        # Mock load_skills_from_roots response
        mock_skill_def = MagicMock()
        mock_skill_def.name = "sample_skill"
        mock_skill_def.path = str(self.mock_skill_source)
        
        self.mock_skills_map = {"sample_skill": mock_skill_def}

    def tearDown(self):
        shutil.rmtree(self.temp_dir)

    @patch("loader.build_meta.load_skills_from_roots")
    def test_download_build_dependencies_default_dest(self, mock_load):
        mock_load.return_value = self.mock_skills_map
        
        # Create pyproject.toml without dest
        pyproject_path = self.workspace_dir / "pyproject.toml"
        pyproject_content = """
        [tool.retailcortex-loader]
        dependencies = ["github://mock/repo:main"]
        """
        pyproject_path.write_text(pyproject_content, encoding="utf-8")
        
        original_cwd = os.getcwd()
        os.chdir(self.workspace_dir)
        try:
            download_build_dependencies(config_path="pyproject.toml")
        finally:
            os.chdir(original_cwd)
            
        mock_load.assert_called_once_with(roots=["github://mock/repo:main"])
        
        # Verify the skill was copied to the default dest (.skills)
        expected_dest = self.workspace_dir / ".skills" / "sample_skill"
        self.assertTrue(expected_dest.exists())
        self.assertTrue((expected_dest / "SKILL.md").exists())

    @patch("loader.build_meta.load_skills_from_roots")
    def test_download_build_dependencies_custom_dest(self, mock_load):
        mock_load.return_value = self.mock_skills_map
        
        # Create pyproject.toml with custom dest
        pyproject_path = self.workspace_dir / "pyproject.toml"
        pyproject_content = """
        [tool.retailcortex-loader]
        dest = "src/my_package/skills"
        dependencies = ["github://mock/repo:main"]
        """
        pyproject_path.write_text(pyproject_content, encoding="utf-8")
        
        original_cwd = os.getcwd()
        os.chdir(self.workspace_dir)
        try:
            download_build_dependencies(config_path="pyproject.toml")
        finally:
            os.chdir(original_cwd)
            
        # Verify the skill was copied to the custom dest
        expected_dest = self.workspace_dir / "src" / "my_package" / "skills" / "sample_skill"
        self.assertTrue(expected_dest.exists())
        self.assertTrue((expected_dest / "SKILL.md").exists())

    def test_download_build_dependencies_no_config(self):
        # Should exit gracefully if pyproject.toml is missing or has no config
        pyproject_path = self.workspace_dir / "pyproject.toml"
        pyproject_path.write_text("[project]\nname='test'", encoding="utf-8")
        
        original_cwd = os.getcwd()
        os.chdir(self.workspace_dir)
        try:
            # Won't raise an exception
            download_build_dependencies(config_path="pyproject.toml")
        finally:
            os.chdir(original_cwd)
            
        self.assertFalse((self.workspace_dir / ".skills").exists())

if __name__ == "__main__":
    unittest.main()
