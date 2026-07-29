import re
from dataclasses import dataclass, field
from typing import List

@dataclass
class SkillFrontmatter:
    name: str
    description: str
    license: str = "Apache-2.0"
    author: str = "Ryan McGuinness"
    version: str = "1.0"

    def __post_init__(self):
        if not self.name or len(self.name) > 64:
            raise ValueError("Skill name must be non-empty and <= 64 characters")
        if not re.match(r"^[a-z0-9]+(-[a-z0-9]+)*$", self.name):
            raise ValueError(f"Skill name '{self.name}' must be strictly kebab-case")
        if not self.description or len(self.description) > 1024:
            raise ValueError("Description must be non-empty and <= 1024 characters")
        if not self.license:
            raise ValueError("License must be non-empty")
        if not self.author:
            raise ValueError("Metadata author must be non-empty")
        if not self.version:
            raise ValueError("Metadata version must be non-empty")

@dataclass
class SkillAuditResult:
    skill_name: str
    directory_path: str
    frontmatter_valid: bool = False
    l3_tree_valid: bool = False
    cwe_security_valid: bool = False
    rate_limit_429_valid: bool = False
    clickable_links_valid: bool = False
    errors: List[str] = field(default_factory=list)

    @property
    def passed(self) -> bool:
        return (
            self.frontmatter_valid
            and self.l3_tree_valid
            and self.cwe_security_valid
            and self.rate_limit_429_valid
            and self.clickable_links_valid
            and len(self.errors) == 0
        )

@dataclass
class AuditSummary:
    total_skills: int = 0
    passed_skills: int = 0
    failed_skills: int = 0
    results: List[SkillAuditResult] = field(default_factory=list)

    def to_dict(self) -> dict:
        return {
            "total_skills": self.total_skills,
            "passed_skills": self.passed_skills,
            "failed_skills": self.failed_skills,
            "results": [
                {
                    "skill_name": r.skill_name,
                    "directory_path": r.directory_path,
                    "passed": r.passed,
                    "frontmatter_valid": r.frontmatter_valid,
                    "l3_tree_valid": r.l3_tree_valid,
                    "cwe_security_valid": r.cwe_security_valid,
                    "rate_limit_429_valid": r.rate_limit_429_valid,
                    "clickable_links_valid": r.clickable_links_valid,
                    "errors": r.errors,
                }
                for r in self.results
            ]
        }
