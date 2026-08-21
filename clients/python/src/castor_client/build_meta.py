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

"""PEP 517 build backend wrapper and build-time skill dependency downloader."""

import os
import shutil
import tomllib
from pathlib import Path
from typing import Any, Dict, List, Optional



from .loader import load_skills_from_roots


def _copy_skill_tree(src_dir: str, dest_dir: Path) -> None:
    """Copies a skill directory tree safely to the destination."""
    src = Path(src_dir)
    if not src.is_dir():
        return
    
    target = dest_dir / src.name
    if target.exists():
        shutil.rmtree(target)
    
    shutil.copytree(src, target)


def download_build_dependencies(config_path: str = "pyproject.toml") -> None:
    """Reads pyproject.toml, downloads skill dependencies, and stages them in the package directory.
    
    Can be called directly from a custom Poetry build.py or legacy setup.py script.
    """
    p = Path(config_path)
    if not p.is_file():
        return
        
    try:
        with open(p, "rb") as f:
            config = tomllib.load(f)
    except Exception:
        return
        
    loader_config = config.get("tool", {}).get("retailcortex-loader", {})
    if not loader_config:
        return
        
    dependencies = loader_config.get("dependencies", [])
    if not dependencies:
        return
        
    dest_str = loader_config.get("dest", ".skills")
    dest_path = Path(dest_str)
    dest_path.mkdir(parents=True, exist_ok=True)
    
    # Load and resolve all skills
    skills_map = load_skills_from_roots(roots=dependencies)
    
    # Stage them into the destination directory
    for name, skill_def in skills_map.items():
        _copy_skill_tree(skill_def.path, dest_path)


def _sm():
    import setuptools.build_meta as sm
    return sm

# --- PEP 517 Hooks ---

def build_wheel(
    wheel_directory: str,
    config_settings: Optional[Dict[str, Any]] = None,
    metadata_directory: Optional[str] = None,
) -> str:
    """PEP 517 hook: builds a wheel after staging skill dependencies."""
    download_build_dependencies()
    return _sm().build_wheel(
        wheel_directory, config_settings, metadata_directory
    )


def build_sdist(
    sdist_directory: str,
    config_settings: Optional[Dict[str, Any]] = None,
) -> str:
    """PEP 517 hook: builds an sdist after staging skill dependencies."""
    download_build_dependencies()
    return _sm().build_sdist(sdist_directory, config_settings)


def build_editable(
    wheel_directory: str,
    config_settings: Optional[Dict[str, Any]] = None,
    metadata_directory: Optional[str] = None,
) -> str:
    """PEP 517 hook: builds an editable wheel after staging skill dependencies."""
    download_build_dependencies()
    sm = _sm()
    # Support older setuptools versions that may not have build_editable
    if hasattr(sm, "build_editable"):
        return sm.build_editable(
            wheel_directory, config_settings, metadata_directory
        )
    return sm.build_wheel(
        wheel_directory, config_settings, metadata_directory
    )

def get_requires_for_build_wheel(
    config_settings: Optional[Dict[str, Any]] = None
) -> List[str]:
    """PEP 517 hook."""
    return _sm().get_requires_for_build_wheel(config_settings)

def get_requires_for_build_sdist(
    config_settings: Optional[Dict[str, Any]] = None
) -> List[str]:
    """PEP 517 hook."""
    return _sm().get_requires_for_build_sdist(config_settings)

def get_requires_for_build_editable(
    config_settings: Optional[Dict[str, Any]] = None
) -> List[str]:
    """PEP 517 hook."""
    sm = _sm()
    if hasattr(sm, "get_requires_for_build_editable"):
        return sm.get_requires_for_build_editable(config_settings)
    return []

def prepare_metadata_for_build_wheel(
    metadata_directory: str,
    config_settings: Optional[Dict[str, Any]] = None,
) -> str:
    """PEP 517 hook."""
    download_build_dependencies()
    return _sm().prepare_metadata_for_build_wheel(
        metadata_directory, config_settings
    )

def prepare_metadata_for_build_editable(
    metadata_directory: str,
    config_settings: Optional[Dict[str, Any]] = None,
) -> str:
    """PEP 517 hook."""
    download_build_dependencies()
    sm = _sm()
    if hasattr(sm, "prepare_metadata_for_build_editable"):
        return sm.prepare_metadata_for_build_editable(
            metadata_directory, config_settings
        )
    return sm.prepare_metadata_for_build_wheel(
        metadata_directory, config_settings
    )
