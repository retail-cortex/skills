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

"""Unit tests for the PEP 517 build wrapper and dependency downloader."""

import os
import shutil
import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch, MagicMock

from loader import build_meta
from loader.build_meta import (
    download_build_dependencies,
    build_wheel,
    build_sdist,
    build_editable,
    get_requires_for_build_wheel,
    get_requires_for_build_sdist,
    get_requires_for_build_editable,
    prepare_metadata_for_build_wheel,
    prepare_metadata_for_build_editable,
    _copy_skill_tree,
)


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

    def test_copy_skill_tree_non_existent(self):
        # Should return safely when src is invalid
        dest = self.workspace_dir / "dest"
        dest.mkdir(parents=True)
        _copy_skill_tree(str(self.workspace_dir / "non_existent"), dest)
        self.assertFalse((dest / "non_existent").exists())

    def test_copy_skill_tree_overwrite(self):
        dest = self.workspace_dir / "dest"
        dest.mkdir(parents=True)
        target = dest / "sample_skill"
        target.mkdir(parents=True)
        (target / "old.txt").write_text("old")
        
        _copy_skill_tree(str(self.mock_skill_source), dest)
        self.assertTrue((target / "SKILL.md").exists())
        self.assertFalse((target / "old.txt").exists())

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
        pyproject_path = self.workspace_dir / "pyproject.toml"
        pyproject_path.write_text("[project]\nname='test'", encoding="utf-8")
        
        original_cwd = os.getcwd()
        os.chdir(self.workspace_dir)
        try:
            download_build_dependencies(config_path="pyproject.toml")
        finally:
            os.chdir(original_cwd)
            
        self.assertFalse((self.workspace_dir / ".skills").exists())

    def test_download_build_dependencies_invalid_toml(self):
        pyproject_path = self.workspace_dir / "pyproject.toml"
        pyproject_path.write_text("invalid = [toml", encoding="utf-8")
        
        original_cwd = os.getcwd()
        os.chdir(self.workspace_dir)
        try:
            download_build_dependencies(config_path="pyproject.toml")
        finally:
            os.chdir(original_cwd)
            
        self.assertFalse((self.workspace_dir / ".skills").exists())

    @patch("loader.build_meta.download_build_dependencies")
    @patch("loader.build_meta._sm")
    def test_pep517_hooks(self, mock_sm_getter, mock_download):
        mock_sm = MagicMock()
        mock_sm.build_wheel.return_value = "wheel.whl"
        mock_sm.build_sdist.return_value = "sdist.tar.gz"
        mock_sm.build_editable.return_value = "editable.whl"
        mock_sm.get_requires_for_build_wheel.return_value = ["req1"]
        mock_sm.get_requires_for_build_sdist.return_value = ["req2"]
        mock_sm.get_requires_for_build_editable.return_value = ["req3"]
        mock_sm.prepare_metadata_for_build_wheel.return_value = "meta.dist-info"
        mock_sm.prepare_metadata_for_build_editable.return_value = "editable.dist-info"
        mock_sm_getter.return_value = mock_sm

        self.assertEqual(build_wheel("dist"), "wheel.whl")
        self.assertEqual(build_sdist("dist"), "sdist.tar.gz")
        self.assertEqual(build_editable("dist"), "editable.whl")
        self.assertEqual(get_requires_for_build_wheel(), ["req1"])
        self.assertEqual(get_requires_for_build_sdist(), ["req2"])
        self.assertEqual(get_requires_for_build_editable(), ["req3"])
        self.assertEqual(prepare_metadata_for_build_wheel("meta"), "meta.dist-info")
        self.assertEqual(prepare_metadata_for_build_editable("meta"), "editable.dist-info")


if __name__ == "__main__":
    unittest.main()
