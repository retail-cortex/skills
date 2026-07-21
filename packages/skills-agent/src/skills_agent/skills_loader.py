"""Dynamic skill scanner and loader for enterprise AI agent skills."""

import os
import re
from pathlib import Path
from typing import Dict, List, Optional, Tuple

from skills_agent.types import SkillDefinition, SkillSummary


def find_registry_root() -> Path:
    """Discovers the root workspace directory containing the skills folder."""
    if "BUILD_WORKSPACE_DIRECTORY" in os.environ:
        workspace = Path(os.environ["BUILD_WORKSPACE_DIRECTORY"])
        if (workspace / "skills").is_dir():
            return workspace

    if "TEST_SRCDIR" in os.environ and "TEST_WORKSPACE" in os.environ:
        runfiles_root = Path(os.environ["TEST_SRCDIR"]) / os.environ["TEST_WORKSPACE"]
        if (runfiles_root / "skills").is_dir():
            return runfiles_root

    current: Path = Path(__file__).resolve().parent
    for parent in [current] + list(current.parents):
        if (parent / "skills").is_dir():
            return parent

    runfiles = os.environ.get("PYTHON_RUNFILES", "")
    if runfiles:
        rf_path = Path(runfiles)
        for cand in rf_path.rglob("skills"):
            if cand.is_dir():
                return cand.parent

    return Path.cwd()


def parse_frontmatter(content: str) -> Tuple[Dict[str, str], str]:
    """Parses YAML frontmatter block from SKILL.md content."""
    pattern = r"^---\s*\n(.*?)\n---\s*\n(.*)$"
    match = re.search(pattern, content, re.DOTALL)
    if not match:
        return {}, content

    yaml_text = match.group(1)
    body = match.group(2)

    data: Dict[str, str] = {}
    for line in yaml_text.splitlines():
        line = line.strip()
        if not line or line.startswith("#") or ":" not in line:
            continue
        key, val = line.split(":", 1)
        data[key.strip()] = val.strip().strip("'\"")

    return data, body


def load_skill_from_dir(skill_dir: Path) -> Optional[SkillDefinition]:
    """Loads a single skill definition from its directory."""
    skill_md = skill_dir / "SKILL.md"
    if not skill_md.is_file():
        return None

    try:
        content = skill_md.read_text(encoding="utf-8")
        fm_data, body = parse_frontmatter(content)
        name = fm_data.get("name", skill_dir.name)
        description = fm_data.get("description", f"Enterprise skill for {name}")

        references: Dict[str, str] = {}
        ref_dir = skill_dir / "references"
        if ref_dir.is_dir():
            for ref_file in sorted(ref_dir.glob("*.md")):
                try:
                    references[ref_file.name] = ref_file.read_text(encoding="utf-8")
                except Exception:
                    pass

        examples: Dict[str, str] = {}
        ex_dir = skill_dir / "examples"
        if ex_dir.is_dir():
            for ex_file in sorted(ex_dir.iterdir()):
                if ex_file.is_file() and not ex_file.name.startswith("."):
                    try:
                        examples[ex_file.name] = ex_file.read_text(encoding="utf-8")
                    except Exception:
                        pass

        return SkillDefinition(
            name=name,
            description=description,
            instructions=body.strip(),
            references=references,
            examples=examples,
            path=str(skill_dir.resolve()),
        )
    except Exception:
        return None


def load_all_skills(skills_root: Optional[Path] = None) -> Dict[str, SkillDefinition]:
    """Scans and loads all skill directories in the repository."""
    root = skills_root or find_registry_root()
    skills_dir = root / "skills" if not root.name == "skills" else root

    loaded: Dict[str, SkillDefinition] = {}
    if not skills_dir.is_dir():
        return loaded

    ignored = {".git", ".bazel", "packages", "validator", "node_modules", "scratch", "build", "dist", ".venv"}

    for entry in sorted(skills_dir.iterdir(), key=lambda p: p.name):
        if entry.is_dir() and entry.name not in ignored and not entry.name.startswith("."):
            skill_def = load_skill_from_dir(entry)
            if skill_def:
                loaded[skill_def.name] = skill_def

    return loaded


class SkillRegistry:
    """High-performance registry for discovering and querying enterprise skills."""

    def __init__(self, skills_root: Optional[Path] = None) -> None:
        self.root: Path = skills_root or find_registry_root()
        self._skills: Dict[str, SkillDefinition] = load_all_skills(self.root)

    @property
    def skills(self) -> Dict[str, SkillDefinition]:
        """Returns map of loaded skill definitions."""
        return self._skills

    def get(self, name: str) -> Optional[SkillDefinition]:
        """Retrieves a skill definition by name."""
        return self._skills.get(name)

    def list_skills(self) -> List[SkillSummary]:
        """Returns summarized metadata for all registered skills."""
        summaries: List[SkillSummary] = []
        for s in self._skills.values():
            summaries.append(
                SkillSummary(
                    name=s.name,
                    description=s.description,
                    reference_count=len(s.references),
                    example_count=len(s.examples),
                    path=s.path,
                )
            )
        return summaries

    def search(self, query: str) -> List[SkillDefinition]:
        """Searches skills by keyword matching in name, description, instructions, or references."""
        q = query.lower()
        results: List[SkillDefinition] = []
        for s in self._skills.values():
            if (
                q in s.name.lower()
                or q in s.description.lower()
                or q in s.instructions.lower()
                or any(q in r_text.lower() for r_text in s.references.values())
            ):
                results.append(s)
        return results

    def get_domain_skills(self, domain: str) -> List[SkillDefinition]:
        """Retrieves skills matching a specific language or domain."""
        domain_norm = domain.lower()
        return [
            s for s in self._skills.values()
            if domain_norm in s.name.lower() or domain_norm in s.description.lower()
        ]
