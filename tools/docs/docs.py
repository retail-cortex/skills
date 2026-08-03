"""Documentation runner entrypoints for MkDocs."""

import os
import sys


def _ensure_workspace_cwd() -> None:
    """Ensure working directory is set to the workspace root if executed via Bazel or subprocess."""
    workspace_dir = os.environ.get("BUILD_WORKSPACE_DIRECTORY")
    if workspace_dir:
        os.chdir(workspace_dir)


def serve() -> None:
    """Serve the MkDocs documentation site locally."""
    _ensure_workspace_cwd()
    try:
        from mkdocs.__main__ import cli
    except ImportError:
        # Fallback to local virtual environment if running in Bazel without hermetic PyPI rule
        venv_site_packages = os.path.join(
            os.environ.get("BUILD_WORKSPACE_DIRECTORY", os.getcwd()),
            ".venv",
            "lib",
            f"python{sys.version_info.major}.{sys.version_info.minor}",
            "site-packages",
        )
        if os.path.exists(venv_site_packages) and venv_site_packages not in sys.path:
            sys.path.insert(0, venv_site_packages)

        try:
            from mkdocs.__main__ import cli
        except ImportError:
            sys.stderr.write(
                "Error: mkdocs is not installed. Install with `uv add mkdocs mkdocs-material`.\n"
            )
            sys.exit(1)

    args = sys.argv[1:]
    if not args or args[0] not in ("serve", "build", "gh-deploy", "new", "get-deps"):
        sys.argv = [sys.argv[0], "serve"] + args
    cli()


def build() -> None:
    """Build the MkDocs documentation site."""
    _ensure_workspace_cwd()
    try:
        from mkdocs.__main__ import cli
    except ImportError:
        venv_site_packages = os.path.join(
            os.environ.get("BUILD_WORKSPACE_DIRECTORY", os.getcwd()),
            ".venv",
            "lib",
            f"python{sys.version_info.major}.{sys.version_info.minor}",
            "site-packages",
        )
        if os.path.exists(venv_site_packages) and venv_site_packages not in sys.path:
            sys.path.insert(0, venv_site_packages)

        try:
            from mkdocs.__main__ import cli
        except ImportError:
            sys.stderr.write(
                "Error: mkdocs is not installed. Install with `uv add mkdocs mkdocs-material`.\n"
            )
            sys.exit(1)

    args = [arg for arg in sys.argv[1:] if arg != "build"]
    sys.argv = [sys.argv[0], "build"] + args
    cli()


if __name__ == "__main__":
    serve()
