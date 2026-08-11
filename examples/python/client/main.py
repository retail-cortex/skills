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


    server_url = os.getenv("CASTOR_SERVER_URL", os.getenv("CSTR_SERVER_URL", os.getenv("SKM_SERVER_URL", "http://localhost:8080")))
    api_key = os.getenv("CASTOR_API_KEY", os.getenv("CSTR_API_KEY", os.getenv("SKM_API_KEY", "")))

    print(f"Loaded Castor Server URL from dotenv: {server_url}")

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
