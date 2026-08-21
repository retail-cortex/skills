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

"""Backward-compatible re-export module delegating to the standalone skills-loader package."""

from castor_client import (
    SkillDefinition,
    SkillRegistry,
    SkillSummary,
    find_registry_root,
    load_all_skills,
    load_skill_from_dir,
    load_skills_from_github,
    load_skills_from_roots,
    parse_dotenv_file,
    parse_frontmatter,
    parse_skill_root_uri,
)

__all__ = [
    "SkillDefinition",
    "SkillSummary",
    "SkillRegistry",
    "find_registry_root",
    "parse_frontmatter",
    "parse_dotenv_file",
    "parse_skill_root_uri",
    "load_skill_from_dir",
    "load_all_skills",
    "load_skills_from_github",
    "load_skills_from_roots",
]
