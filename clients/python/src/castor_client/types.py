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

"""Type definitions and data structures for enterprise skill definitions."""

from dataclasses import dataclass, field
from enum import Enum
from typing import Any, Dict, List, Optional


class HITLPolicyTier(str, Enum):
    """HITL intervention policy tiers for agent execution safety."""

    TIER_0_BYPASS_ALL = "TIER_0_BYPASS_ALL"
    TIER_1_AUTO_READ = "TIER_1_AUTO_READ"
    TIER_2_AUDITED_WRITE = "TIER_2_AUDITED_WRITE"
    TIER_3_MANDATORY_APPROVAL = "TIER_3_MANDATORY_APPROVAL"


@dataclass
class CompiledSkillReference:
    """Minimized, cryptographically locked pointer to a compiled skill definition."""

    skill_id: str
    name: str
    description: str
    sha256_hash: str
    json_schema: Dict[str, Any]
    strict_schema: bool = True
    allowed_properties: List[str] = field(default_factory=list)
    estimated_tokens: int = 50
    hitl_tier: HITLPolicyTier = HITLPolicyTier.TIER_1_AUTO_READ

    def to_dict(self) -> Dict[str, Any]:
        return {
            "skill_id": self.skill_id,
            "name": self.name,
            "description": self.description,
            "sha256_hash": self.sha256_hash,
            "json_schema": self.json_schema,
            "strict_schema": self.strict_schema,
            "allowed_properties": self.allowed_properties,
            "estimated_tokens": self.estimated_tokens,
            "hitl_tier": self.hitl_tier.value if isinstance(self.hitl_tier, HITLPolicyTier) else str(self.hitl_tier),
        }


@dataclass
class HITLGateResult:
    """Result of a Human-in-the-Loop policy gate evaluation."""

    approved: bool
    tier: HITLPolicyTier
    reason: str
    bypassed: bool = False
    requires_user_input: bool = False
    audit_event_id: str = ""


@dataclass
class SkillDirectorySearchResult:
    """Structured result returned by the skill_directory_search meta-tool."""

    query_intent: str
    matches: List[CompiledSkillReference] = field(default_factory=list)
    total_found: int = 0


@dataclass
class SkillDefinition:
    """Represents a loaded enterprise skill definition."""

    name: str
    description: str
    instructions: str
    license: Optional[str] = None
    author: Optional[str] = None
    version: Optional[str] = None
    compatibility: Optional[str] = None
    allowed_tools: Optional[str] = None
    authors: List[Dict[str, str]] = field(default_factory=list)
    tool_requirements: List[Dict[str, object]] = field(default_factory=list)
    category: Optional[str] = None
    tags: List[str] = field(default_factory=list)
    trigger_phrases: List[str] = field(default_factory=list)
    execution_hints: Dict[str, object] = field(default_factory=dict)
    metadata: Dict[str, str] = field(default_factory=dict)
    references: Dict[str, str] = field(default_factory=dict)
    examples: Dict[str, str] = field(default_factory=dict)
    scripts: List[Dict[str, object]] = field(default_factory=list)
    resources: List[Dict[str, object]] = field(default_factory=list)
    path: str = ""
    source_uri: str = ""
    compiled_reference: Optional[CompiledSkillReference] = None
    sha256_hash: str = ""

    def get_reference_content(self, name: str) -> Optional[str]:
        """Lazy / on-demand retriever for reference file content."""
        if name in self.references:
            return self.references[name]
        return None

    def get_example_content(self, name: str) -> Optional[str]:
        """Lazy / on-demand retriever for example file content."""
        if name in self.examples:
            return self.examples[name]
        return None

    def to_dict(self) -> Dict[str, object]:
        """Serializes skill definition to dictionary format."""
        return {
            "name": self.name,
            "description": self.description,
            "instructions": self.instructions,
            "license": self.license,
            "author": self.author,
            "authors": self.authors,
            "version": self.version,
            "compatibility": self.compatibility,
            "allowed_tools": self.allowed_tools,
            "tool_requirements": self.tool_requirements,
            "category": self.category,
            "tags": self.tags,
            "trigger_phrases": self.trigger_phrases,
            "execution_hints": self.execution_hints,
            "metadata": self.metadata,
            "references": list(self.references.keys()),
            "examples": list(self.examples.keys()),
            "scripts": self.scripts,
            "resources": self.resources,
            "path": self.path,
            "source_uri": self.source_uri,
            "compiled_reference": self.compiled_reference.to_dict() if self.compiled_reference else None,
            "sha256_hash": self.sha256_hash,
        }


@dataclass
class SkillSummary:
    """High-level summary of a registered skill."""

    name: str
    description: str
    reference_count: int
    example_count: int
    path: str
    category: Optional[str] = None
    tags: List[str] = field(default_factory=list)
    trigger_phrases: List[str] = field(default_factory=list)
    sha256_hash: str = ""
    hitl_tier: HITLPolicyTier = HITLPolicyTier.TIER_1_AUTO_READ
    script_count: int = 0
    resource_count: int = 0


