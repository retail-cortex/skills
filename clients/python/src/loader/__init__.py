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

"""Standalone reusable skill scanner and loader package for Google ADK agents."""

from loader.compiler import SkillCompiler
from loader.discovery import SkillDiscoveryEngine, TFIDFVectorIndex
from loader.hitl import HITLEngine
from loader.loader import (
    SkillRegistry,
    build_skills_manifest,
    find_registry_root,
    get_loader_skills_dir,
    load_all_skills,
    load_skill_from_dir,
    load_skills_from_entry_points,
    load_skills_from_github,
    load_skills_from_manifest,
    load_skills_from_package,
    load_skills_from_roots,
    parse_dotenv_file,
    parse_frontmatter,
    parse_skill_root_uri,
    read_manifest_lock,
    update_manifest_lock,
    verify_manifest_lock,
    write_manifest_lock,
)
from loader.build_meta import download_build_dependencies
from loader.types import (
    CompiledSkillReference,
    HITLGateResult,
    HITLPolicyTier,
    SkillDefinition,
    SkillDirectorySearchResult,
    SkillSummary,
)

__all__ = [
    "SkillDefinition",
    "SkillSummary",
    "CompiledSkillReference",
    "HITLGateResult",
    "HITLPolicyTier",
    "SkillDirectorySearchResult",
    "SkillCompiler",
    "SkillDiscoveryEngine",
    "TFIDFVectorIndex",
    "HITLEngine",
    "SkillRegistry",
    "find_registry_root",
    "get_loader_skills_dir",
    "parse_frontmatter",
    "parse_dotenv_file",
    "parse_skill_root_uri",
    "load_skill_from_dir",
    "load_all_skills",
    "load_skills_from_entry_points",
    "load_skills_from_github",
    "load_skills_from_package",
    "load_skills_from_roots",
    "build_skills_manifest",
    "load_skills_from_manifest",
    "read_manifest_lock",
    "write_manifest_lock",
    "update_manifest_lock",
    "verify_manifest_lock",
    "download_build_dependencies",
]

