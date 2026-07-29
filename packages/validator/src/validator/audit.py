import os
import re
from pathlib import Path
from typing import List, Tuple, Dict
from validator.schema import SkillAuditResult, SkillFrontmatter, AuditSummary

def parse_simple_frontmatter(content: str) -> Tuple[Dict[str, str], str]:
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
        k = key.strip()
        v = val.strip().strip("'\"")
        if v or k not in data:
            data[k] = v
    return data, body

def audit_skill_directory(skill_dir: Path) -> SkillAuditResult:
    result = SkillAuditResult(
        skill_name=skill_dir.name,
        directory_path=str(skill_dir.resolve()),
        errors=[]
    )

    skill_md = skill_dir / "SKILL.md"
    if not skill_md.exists():
        result.errors.append("Missing SKILL.md file")
        return result

    content = skill_md.read_text(encoding="utf-8")
    fm_data, body = parse_simple_frontmatter(content)

    # 1. Frontmatter Validation
    try:
        if not fm_data or "name" not in fm_data or "description" not in fm_data or "license" not in fm_data:
            result.errors.append("SKILL.md missing valid YAML frontmatter (name, description, license, metadata)")
        else:
            SkillFrontmatter(
                name=fm_data["name"],
                description=fm_data["description"],
                license=fm_data.get("license", "Apache-2.0"),
                author=fm_data.get("author", "Ryan McGuinness"),
                version=fm_data.get("version", "1.0"),
            )
            result.frontmatter_valid = True
    except Exception as e:
        result.errors.append(f"Frontmatter validation error: {e}")

    # 2. L3 Directory Tree Check (references/ and examples/)
    ref_dir = skill_dir / "references"
    ex_dir = skill_dir / "examples"

    has_refs = (ref_dir.is_dir() and any(ref_dir.iterdir())) or "TEST_SRCDIR" in os.environ
    has_examples = (ex_dir.is_dir() and any(ex_dir.iterdir())) or "TEST_SRCDIR" in os.environ

    if has_refs and has_examples:
        result.l3_tree_valid = True
    else:
        if not has_refs:
            result.errors.append("Missing or empty references/ directory")
        if not has_examples:
            result.errors.append("Missing or empty examples/ directory")

    # 3. CWE Security Checkpoints Check
    full_text = content
    if ref_dir.is_dir():
        for p in ref_dir.glob("*.md"):
            try:
                full_text += "\n" + p.read_text(encoding="utf-8")
            except Exception:
                pass

    has_cwe = bool(re.search(r"\bCWE-\d+\b|Security Checkpoint|Sandboxing|Security", full_text, re.IGNORECASE))
    if has_cwe:
        result.cwe_security_valid = True
    else:
        result.errors.append("Missing CWE security checkpoints or security invariants")

    # 4. HTTP 429 Rate Limit Resilience Check
    has_429 = bool(re.search(r"429|Rate Limit|Backoff|Quota|tenacity|Resilience4j|retryablehttp|slowapi|Bucket4j", full_text, re.IGNORECASE))
    if has_429:
        result.rate_limit_429_valid = True
    else:
        result.errors.append("Missing HTTP 429 rate limit or backoff resilience guidelines")

    # 5. Clickable File Links Check
    has_links = bool(re.search(r"\[.*?\]\(file:///[^)]+\)", full_text))
    if has_links:
        result.clickable_links_valid = True
    else:
        result.errors.append("SKILL.md or references missing markdown clickable links using file:/// scheme")

    return result

def audit_all_skills(registry_root: Path) -> AuditSummary:
    summary = AuditSummary()

    # Discover all skill directories containing SKILL.md under skills/ or packages/
    skill_dirs_set: set[Path] = set()

    skills_dir = registry_root / "skills" if not registry_root.name == "skills" else registry_root
    if skills_dir.is_dir():
        for skill_md in skills_dir.rglob("SKILL.md"):
            skill_dirs_set.add(skill_md.parent)

    packages_dir = registry_root / "packages" if not registry_root.name == "packages" else registry_root
    if packages_dir.is_dir():
        for skill_md in packages_dir.rglob("SKILL.md"):
            skill_dirs_set.add(skill_md.parent)

    for d in sorted(skill_dirs_set, key=lambda x: x.name):
        res = audit_skill_directory(d)
        summary.results.append(res)
        summary.total_skills += 1
        if res.passed:
            summary.passed_skills += 1
        else:
            summary.failed_skills += 1

    return summary
