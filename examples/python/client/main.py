"""Standalone Python Client Example for Enterprise AI Agent Skills."""

import os
import sys
from pathlib import Path

# Add python loader package to path for execution
pkg_root = Path(__file__).resolve().parents[3] / "clients/python/src"
if str(pkg_root) not in sys.path:
    sys.path.insert(0, str(pkg_root))

from loader import parse_skill_root_uri



def run() -> dict[str, str]:
    # 1. Load environment properties using python-dotenv with fallback
    env_path = Path(__file__).parent / ".env"
    try:
        from dotenv import load_dotenv
        load_dotenv(dotenv_path=env_path)
    except ImportError:
        if env_path.exists():
            for line in env_path.read_text().splitlines():
                if line and not line.startswith("#") and "=" in line:
                    k, v = line.split("=", 1)
                    os.environ.setdefault(k.strip(), v.strip())


    server_url = os.getenv("SKM_SERVER_URL", "http://localhost:8080")
    api_key = os.getenv("SKM_API_KEY", "")

    print(f"Loaded SKM Server URL from dotenv: {server_url}")

    # 2. Parse polyglot skill URI
    uri = "github://google/skills@main/tree/main/skills/cloud/gemini-api"
    scheme, target, ref, subpath = parse_skill_root_uri(uri)
    print(f"Parsed URI: scheme={scheme}, target={target}, ref={ref}, subpath={subpath}")

    return {
        "server_url": server_url,
        "api_key": api_key,
        "scheme": scheme,
        "target": target,
    }


if __name__ == "__main__":
    run()
