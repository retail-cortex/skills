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

"""Skill Compiler module for enterprise AI agent skills.

Strips verbose natural language markdown and developer comments, produces strict or
permissive JSON Schemas, and generates immutable SHA-256 cryptographic digests.
"""

import hashlib
import json
import re
from typing import Any, Dict, List, Optional, Tuple

from .types import CompiledSkillReference, HITLPolicyTier, SkillDefinition


class SkillCompiler:
    """Compiles raw SkillDefinitions into lightweight, cryptographically locked pointers."""

    def __init__(
        self,
        strict_schemas: bool = True,
        allow_additional_properties: Optional[List[str]] = None,
    ):
        self.strict_schemas = strict_schemas
        self.allow_additional_properties = allow_additional_properties or []

    @staticmethod
    def compute_sha256(content: str) -> str:
        """Computes SHA-256 hexadecimal digest of raw content string."""
        return hashlib.sha256(content.encode("utf-8")).hexdigest()

    @staticmethod
    def strip_comments_and_formatting(raw_text: str) -> str:
        """Strips HTML/Markdown comments and redundant whitespace."""
        # Remove HTML comments
        clean = re.sub(r"<!--.*?-->", "", raw_text, flags=re.DOTALL)
        # Remove empty lines / consecutive space
        clean = re.sub(r"\n\s*\n", "\n", clean)
        return clean.strip()

    def build_json_schema(
        self, skill: SkillDefinition
    ) -> Dict[str, Any]:
        """Generates a JSON Schema for invocation of the skill parameters."""
        properties: Dict[str, Any] = {
            "skill_name": {
                "type": "string",
                "const": skill.name,
                "description": f"Target skill identifier: {skill.name}",
            },
            "intent": {
                "type": "string",
                "description": f"Execution intent matching: {skill.description[:100]}",
            },
        }

        # Add parameters based on tool requirements if present
        for tool_req in skill.tool_requirements:
            tool_name = tool_req.get("name", "tool")
            if isinstance(tool_name, str):
                safe_key = re.sub(r"[^a-zA-Z0-9_]", "_", tool_name).lower()
                properties[f"use_{safe_key}"] = {
                    "type": "boolean",
                    "description": f"Enable execution of {tool_name}",
                    "default": True,
                }

        schema: Dict[str, Any] = {
            "type": "object",
            "properties": properties,
            "required": ["skill_name", "intent"],
        }

        if self.strict_schemas and not self.allow_additional_properties:
            schema["additionalProperties"] = False
        elif self.allow_additional_properties:
            schema["additionalProperties"] = {
                "type": "string"
            }
            schema["patternProperties"] = {
                f"^({ '|'.join(map(re.escape, self.allow_additional_properties)) })$": {}
            }
        else:
            schema["additionalProperties"] = True

        return schema

    def determine_hitl_tier(self, skill: SkillDefinition) -> HITLPolicyTier:
        """Determines security policy tier from execution hints."""
        hints = skill.execution_hints or {}
        explicit_tier = hints.get("hitl_tier")
        if explicit_tier:
            try:
                return HITLPolicyTier(str(explicit_tier))
            except ValueError:
                pass

        requires_approval = hints.get("requires_human_approval", False)
        if requires_approval:
            return HITLPolicyTier.TIER_3_MANDATORY_APPROVAL

        # Fall back to category / name heuristics
        name_lower = skill.name.lower()
        if any(w in name_lower for w in ["drop", "delete", "destroy", "deploy", "secret", "publish"]):
            return HITLPolicyTier.TIER_3_MANDATORY_APPROVAL
        elif any(w in name_lower for w in ["write", "update", "create", "modify", "setup"]):
            return HITLPolicyTier.TIER_2_AUDITED_WRITE
        else:
            return HITLPolicyTier.TIER_1_AUTO_READ

    def compile(self, skill: SkillDefinition) -> CompiledSkillReference:
        """Compiles a SkillDefinition into a CompiledSkillReference."""
        raw_payload = f"{skill.name}:{skill.version}:{skill.instructions}:{skill.description}"
        sha256_hash = self.compute_sha256(raw_payload)

        json_schema = self.build_json_schema(skill)
        hitl_tier = self.determine_hitl_tier(skill)

        # Estimate tokens: ~4 chars per token for description + schema json
        schema_json_str = json.dumps(json_schema)
        estimated_tokens = max(30, (len(skill.description) + len(schema_json_str)) // 4)

        ref = CompiledSkillReference(
            skill_id=f"ref-{sha256_hash[:12]}",
            name=skill.name,
            description=skill.description,
            sha256_hash=sha256_hash,
            json_schema=json_schema,
            strict_schema=self.strict_schemas,
            allowed_properties=self.allow_additional_properties,
            estimated_tokens=estimated_tokens,
            hitl_tier=hitl_tier,
        )

        skill.compiled_reference = ref
        skill.sha256_hash = sha256_hash
        return ref
