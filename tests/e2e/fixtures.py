"""Fixtures and setup helpers for skills loader E2E tests."""

import tempfile
from contextlib import contextmanager
from pathlib import Path
from typing import Generator, List, Set, Tuple
from unittest.mock import MagicMock, patch

from skills_loader.loader import find_registry_root


def get_workspace_root() -> Path:
    """Returns the discovered workspace root path."""
    return find_registry_root()


def get_multi_skill_file_roots(root: Path | None = None) -> List[str]:
    """Returns file:// URIs pointing to distinct skills across python, go, and bazel packages."""
    ws_root = root or get_workspace_root()
    python_skill_path = ws_root / "packages/skills-python/src/retailcortex_skills_python/skills/python-core"
    go_skill_path = ws_root / "packages/skills-go/src/retailcortex_skills_go/skills/go-lang"
    bazel_skill_path = ws_root / "packages/skills-bazel/src/retailcortex_skills_bazel/skills/bazel-modules"

    return [
        f"file://{python_skill_path}",
        f"file://{go_skill_path}",
        f"file://{bazel_skill_path}",
    ]


def get_expected_package_skills(package_name: str = "retailcortex_skills_python") -> Set[str]:
    """Returns the set of expected skill names for a given package."""
    package_map = {
        "retailcortex_skills_python": {
            "python-core",
            "python-project-setup",
            "python-fastapi",
            "python-adk-fastapi",
            "python-fastmcp",
        },
    }
    return package_map.get(package_name, set())


def get_github_test_uri_spec() -> Tuple[str, str, str, str, str]:
    """Returns (uri, expected_scheme, expected_target, expected_ref, expected_subpath)."""
    return (
        "github://google/skills/skills/cloud/gemini-api:main",
        "github",
        "google/skills",
        "main",
        "skills/cloud/gemini-api",
    )


@contextmanager
def mock_github_skill_environment() -> Generator[Tuple[str, str, List[str]], None, None]:
    """Context manager setting up a hermetic mocked GitHub skill directory structure and subprocess patches."""
    with tempfile.TemporaryDirectory() as tmp_dir:
        fake_root = Path(tmp_dir)
        repo_target = fake_root / "repo"
        github_skill_dir = repo_target / "skills" / "cloud" / "gemini-api"
        github_skill_dir.mkdir(parents=True, exist_ok=True)

        (github_skill_dir / "SKILL.md").write_text(
            "---\nname: gemini-api\ndescription: Google Cloud Gemini API Skill\nauthor: Google DeepMind\nversion: 1.0.0\n---\n# Instructions\nUse Google GenAI SDK for Gemini model integration.",
            encoding="utf-8",
        )
        ref_dir = github_skill_dir / "references"
        ref_dir.mkdir(exist_ok=True)
        (ref_dir / "api_reference.md").write_text("# Gemini API Reference\nUse genai.Client()", encoding="utf-8")

        with patch("subprocess.run") as mock_run, patch("tempfile.TemporaryDirectory") as mock_tmp:
            mock_proc = MagicMock()
            mock_proc.returncode = 0
            mock_run.return_value = mock_proc
            mock_tmp.return_value.__enter__.return_value = tmp_dir

            yield ("google/skills", "main", ["skills/cloud/gemini-api"])


@contextmanager
def create_temp_manifest_dir() -> Generator[Path, None, None]:
    """Context manager yielding a temporary directory path for manifest generation."""
    with tempfile.TemporaryDirectory() as tmp_dir:
        yield Path(tmp_dir) / "skills_manifest.json"
