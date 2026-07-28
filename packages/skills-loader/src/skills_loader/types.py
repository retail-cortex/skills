"""Type definitions and data structures for enterprise skill definitions."""

from dataclasses import dataclass, field
from typing import Dict, List, Optional


@dataclass
class SkillDefinition:
    """Represents a loaded enterprise skill definition."""

    name: str
    description: str
    instructions: str
    references: Dict[str, str] = field(default_factory=dict)
    examples: Dict[str, str] = field(default_factory=dict)
    path: str = ""

    def to_dict(self) -> Dict[str, object]:
        """Serializes skill definition to dictionary format."""
        return {
            "name": self.name,
            "description": self.description,
            "instructions": self.instructions,
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
