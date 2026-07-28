"""Runner entry points for example-adk CLI scripts."""

import asyncio
import os
import sys
from pathlib import Path

root = Path(__file__).resolve().parents[1]
loader_src = root / "packages/skills-loader/src"
agent_src = root / "packages/skills-agent/src"
example_dir = root / "examples/example-adk"

for p in [str(loader_src), str(agent_src), str(example_dir)]:
    if p not in sys.path:
        sys.path.insert(0, p)


def main() -> None:
    """Runs the native ADK example script."""
    # Ensure working directory is examples/example-adk so .env is read cleanly
    os.chdir(example_dir)
    from main import main as example_main
    asyncio.run(example_main())


def web() -> None:
    """Runs the native ADK example web server control plane."""
    os.chdir(example_dir)
    from skills_agent.server import start_server
    start_server(host="0.0.0.0", port=8000)


if __name__ == "__main__":
    main()
