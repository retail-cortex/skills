"""Type definitions and data structures for enterprise skill definitions."""

from dataclasses import dataclass, field
from typing import Dict, List, Optional


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
    metadata: Dict[str, str] = field(default_factory=dict)
    references: Dict[str, str] = field(default_factory=dict)
    examples: Dict[str, str] = field(default_factory=dict)
    path: str = ""

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
            "version": self.version,
            "compatibility": self.compatibility,
            "allowed_tools": self.allowed_tools,
            "metadata": self.metadata,
            "references": list(self.references.keys()),
            "examples": list(self.examples.keys()),
            "path": self.path,
        }


@dataclass
class SkillSummary:
    """High-level summary of a registered skill."""

    name: str
    description: str
    reference_count: int
    example_count: int
    path: str
