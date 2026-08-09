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

import tempfile
import unittest
from pathlib import Path
import sys

project_root = Path(__file__).resolve().parents[4]
examples_dir = project_root / "examples/python/polyglot"
if str(examples_dir) not in sys.path:
    sys.path.insert(0, str(examples_dir))


sys.modules.pop("main", None)
from main import PolyglotDeveloperAgent


class TestPolyglotDeveloperExample(unittest.TestCase):
    """Unit test suite for examples/polyglot-developer package."""

    def test_polyglot_developer_scaffold(self) -> None:
        agent = PolyglotDeveloperAgent()
        status = agent.verify_skills_available()

        self.assertTrue(status.get("bazel-modules", False))
        self.assertTrue(status.get("go-lang", False))
        self.assertTrue(status.get("java-enterprise", False))
        self.assertTrue(status.get("protobuf-grpc", False))
        self.assertTrue(status.get("python-core", False))
        self.assertTrue(status.get("react-vite", False))

        with tempfile.TemporaryDirectory() as tmp_dir:
            target_path = Path(tmp_dir) / "polyglot-app"
            created = agent.scaffold_polyglot_project(target_path)
            self.assertGreaterEqual(len(created), 7)
            self.assertTrue((target_path / "MODULE.bazel").is_file())
            self.assertTrue((target_path / "proto" / "user.proto").is_file())
            self.assertTrue((target_path / "services" / "go-service" / "main.go").is_file())
            self.assertTrue((target_path / "services" / "java-service" / "src" / "main" / "java" / "com" / "polyglot" / "Application.java").is_file())
            self.assertTrue((target_path / "services" / "python-service" / "main.py").is_file())
            self.assertTrue((target_path / "apps" / "web-dashboard" / "src" / "App.tsx").is_file())


if __name__ == "__main__":
    unittest.main()
