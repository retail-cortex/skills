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
