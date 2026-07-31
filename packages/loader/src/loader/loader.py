"""Dynamic skill scanner and loader for enterprise AI agent skills compatible with Google ADK."""

import hashlib
import importlib
import importlib.metadata
import importlib.resources
import json
import os
import re
import shutil
import subprocess
import tempfile
import urllib.request
import zipfile
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple, Union

from loader.compiler import SkillCompiler
from loader.discovery import SkillDiscoveryEngine
from loader.hitl import HITLEngine
from loader.types import (
    CompiledSkillReference,
    HITLGateResult,
    HITLPolicyTier,
    SkillDefinition,
    SkillDirectorySearchResult,
    SkillSummary,
)


def find_registry_root() -> Path:
    """Discovers the root workspace directory containing enterprise skill packages."""
    if "BUILD_WORKSPACE_DIRECTORY" in os.environ:
        workspace = Path(os.environ["BUILD_WORKSPACE_DIRECTORY"])
        if (workspace / "packages").exists() or (workspace / "skills").exists():
            return workspace

    # 1. Official rules_python Rlocation lookup
    try:
        from rules_python.python.runfiles import runfiles
        r = runfiles.Create()
        if r:
            for ws in [os.environ.get("TEST_WORKSPACE", ""), "_main", "skill_builder"]:
                if ws:
                    p = r.Rlocation(f"{ws}/packages/skills-python/src/retailcortex_skills_python/skills/python-core/SKILL.md")
                    if p and Path(p).exists():
                        return Path(p).parent.parent.parent.parent.parent.parent.parent
                    p2 = r.Rlocation(f"{ws}/skills/python-core/SKILL.md")
                    if p2 and Path(p2).exists():
                        return Path(p2).parent.parent.parent
    except Exception:
        pass

    # 2. Bazel TEST_SRCDIR runfiles tree discovery
    if "TEST_SRCDIR" in os.environ:
        test_srcdir = Path(os.environ["TEST_SRCDIR"])
        workspace_name = os.environ.get("TEST_WORKSPACE", "")

        candidates = []
        if workspace_name:
            candidates.append(test_srcdir / workspace_name)
        candidates.extend([
            test_srcdir / "_main",
            test_srcdir / "skill_builder",
            test_srcdir,
        ])

        for cand in candidates:
            if (cand / "packages").exists() or (cand / "skills").exists():
                return cand

        # Search for any SKILL.md under TEST_SRCDIR
        for cand in test_srcdir.rglob("SKILL.md"):
            for parent in [cand] + list(cand.parents):
                if (parent / "packages").exists() or (parent / "skills").exists():
                    return parent

    # 3. Source directory parent walk fallback
    current: Path = Path(__file__).resolve().parent
    for parent in [current] + list(current.parents):
        if (parent / "packages").exists() or (parent / "skills").exists():
            return parent.resolve()

    # 4. PYTHON_RUNFILES environment variable fallback
    runfiles_env = os.environ.get("PYTHON_RUNFILES", "")
    if runfiles_env:
        rf_path = Path(runfiles_env)
        for cand in rf_path.rglob("SKILL.md"):
            for parent in [cand] + list(cand.parents):
                if (parent / "packages").exists() or (parent / "skills").exists():
                    return parent

    return Path.cwd().resolve()


def get_loader_skills_dir() -> Path:
    """Returns the persistent workspace directory for cached downloaded skill trees."""
    loader_dir = find_registry_root() / ".loader_skills"
    loader_dir.mkdir(parents=True, exist_ok=True)
    return loader_dir


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
        k = key.strip()
        v = val.strip().strip("'\"")
        if v or k not in data:
            data[k] = v

    return data, body


def parse_dotenv_file(path: Union[Path, str]) -> Dict[str, str]:
    """Parses key-value environment variables from a .env or dotenv configuration file."""
    env_vars: Dict[str, str] = {}
    p = Path(path)
    if not p.is_file():
        return env_vars
    try:
        content = p.read_text(encoding="utf-8")
        for line in content.splitlines():
            line = line.strip()
            if not line or line.startswith("#") or "=" not in line:
                continue
            key, val = line.split("=", 1)
            env_vars[key.strip()] = val.strip().strip("'\"")
    except Exception:
        pass
    return env_vars


def parse_skill_root_uri(uri: str) -> Tuple[str, str, Optional[str], Optional[str]]:
    """Parses a qualified skill root URI into (scheme, target, ref, subpath).

    Supported URI formats:
    - file://path/to/skills (e.g. file://. or file://packages)
    - pkg://package_name (e.g. pkg://retailcortex_skills_python - scans 'skills/' sub-directory if present, or package root)
    - github://owner/repo/subpath:ref (Standard format with trailing :ref, e.g. github://google/skills/skills/cloud/gemini-api:main)
    - github://owner/repo:ref (Repo root with trailing :ref, e.g. github://owner/repo:v2.5.0)
    - github://owner/repo/tree/ref/subpath (GitHub web URL format)
    - Legacy formats supported for backwards compatibility: github://owner/repo:ref/subpath, github://owner/repo@ref/subpath
    """
    clean = uri.strip()
    if clean.startswith("file://"):
        target = clean[len("file://"):]
        return "file", target, None, None
    elif clean.startswith("pkg://") or clean.startswith("package://"):
        prefix = "pkg://" if clean.startswith("pkg://") else "package://"
        target = clean[len(prefix):]
        return "pkg", target, None, None
    elif clean.startswith("maven://") or clean.startswith("mvn://"):
        prefix = "maven://" if clean.startswith("maven://") else "mvn://"
        raw = clean[len(prefix):]
        parts = raw.split("/", 1)
        subpath = parts[1] if len(parts) > 1 else None
        coords = parts[0].split(":")
        if len(coords) >= 3:
            target = f"{coords[0]}:{coords[1]}"
            ref = coords[2]
        else:
            target = parts[0]
            ref = None
        scheme = "maven" if clean.startswith("maven://") else "mvn"
        return scheme, target, ref, subpath
    elif clean.startswith("mod://") or clean.startswith("go://"):
        prefix = "mod://" if clean.startswith("mod://") else "go://"
        raw = clean[len(prefix):]
        ref = None
        subpath = None
        if "@" in raw:
            mod_part, rest = raw.split("@", 1)
            target = mod_part
            if "/" in rest:
                ref_part, subpath_part = rest.split("/", 1)
                ref = ref_part
                subpath = subpath_part
            else:
                ref = rest
        else:
            target = raw
        scheme = "mod" if clean.startswith("mod://") else "go"
        return scheme, target, ref, subpath
    elif clean.startswith("github://") or clean.startswith("https://github.com/"):
        prefix = "github://" if clean.startswith("github://") else "https://github.com/"
        target = clean[len(prefix):].rstrip("/")

        ref: Optional[str] = None
        subpath: Optional[str] = None

        if "/tree/" in target:
            repo_part, tree_part = target.split("/tree/", 1)
            target = repo_part
            tree_bits = tree_part.split("/", 1)
            ref = tree_bits[0]
            if len(tree_bits) > 1:
                subpath = tree_bits[1]
        elif ":" in target:
            part1, part2 = target.rsplit(":", 1)
            if "/" not in part2:
                ref = part2
                target = part1
            else:
                repo_part, ref_sub = target.split(":", 1)
                target = repo_part
                if "/" in ref_sub:
                    ref, subpath = ref_sub.split("/", 1)
                else:
                    ref = ref_sub
        elif "@" in target:
            repo_part, ref_sub = target.split("@", 1)
            target = repo_part
            if "/" in ref_sub:
                ref, subpath = ref_sub.split("/", 1)
            else:
                ref = ref_sub

        if not subpath:
            parts = target.split("/")
            if len(parts) > 2:
                target = f"{parts[0]}/{parts[1]}"
                subpath = "/".join(parts[2:])

        return "github", target, ref or "main", subpath
    else:
        return "file", clean, None, None


def load_skill_from_dir(skill_dir: Path) -> Optional[SkillDefinition]:
    """Loads a single skill definition from its directory, enforcing symlink boundary checks."""
    skill_md = skill_dir / "SKILL.md"
    if not skill_md.is_file():
        return None

    try:
        content = skill_md.read_text(encoding="utf-8")
        fm_data, body = parse_frontmatter(content)
        name = fm_data.get("name", skill_dir.name)
        description = fm_data.get("description", f"Enterprise skill for {name}")
        license_val = fm_data.get("license")
        author_val = fm_data.get("author")
        authors_val = fm_data.get("authors") or []
        version_val = fm_data.get("version")
        compatibility_val = fm_data.get("compatibility")
        allowed_tools_val = fm_data.get("allowed-tools") or fm_data.get("allowed_tools")
        tool_reqs_val = fm_data.get("tool_requirements") or []
        category_val = fm_data.get("category")
        tags_val = fm_data.get("tags") or []
        triggers_val = fm_data.get("trigger_phrases") or []
        execution_hints_val = fm_data.get("execution_hints") or {}

        known_keys = {
            "name", "description", "license", "author", "authors", "version", "compatibility",
            "allowed-tools", "allowed_tools", "tool_requirements", "category", "tags",
            "trigger_phrases", "execution_hints"
        }
        meta_dict = {k: v for k, v in fm_data.items() if k not in known_keys}

        if not author_val and "author" in meta_dict:
            author_val = meta_dict["author"]
        if not version_val and "version" in meta_dict:
            version_val = meta_dict["version"]

        if author_val and "author" not in meta_dict:
            meta_dict["author"] = author_val
        if version_val and "version" not in meta_dict:
            meta_dict["version"] = version_val

        references: Dict[str, str] = {}
        ref_dir = skill_dir / "references"
        if ref_dir.is_dir():
            for ref_file in sorted(ref_dir.glob("*.md")):
                try:
                    if ref_file.is_relative_to(skill_dir):
                        references[ref_file.name] = ref_file.read_text(encoding="utf-8")
                        continue
                    resolved_ref = ref_file.resolve()
                    resolved_base = skill_dir.resolve()
                    if resolved_ref != resolved_base and resolved_base not in resolved_ref.parents:
                        try:
                            if not resolved_ref.is_relative_to(resolved_base):
                                continue
                        except ValueError:
                            continue
                    references[ref_file.name] = ref_file.read_text(encoding="utf-8")
                except Exception:
                    pass

        examples: Dict[str, str] = {}
        ex_dir = skill_dir / "examples"
        if ex_dir.is_dir():
            for ex_file in sorted(ex_dir.iterdir()):
                if ex_file.is_file() and not ex_file.name.startswith("."):
                    try:
                        if ex_file.is_relative_to(skill_dir):
                            examples[ex_file.name] = ex_file.read_text(encoding="utf-8")
                            continue
                        resolved_ex = ex_file.resolve()
                        resolved_base = skill_dir.resolve()
                        if resolved_ex != resolved_base and resolved_base not in resolved_ex.parents:
                            try:
                                if not resolved_ex.is_relative_to(resolved_base):
                                    continue
                            except ValueError:
                                continue
                        examples[ex_file.name] = ex_file.read_text(encoding="utf-8")
                    except Exception:
                        pass

        return SkillDefinition(
            name=name,
            description=description,
            instructions=body.strip(),
            license=license_val,
            author=author_val,
            authors=authors_val,
            version=version_val,
            compatibility=compatibility_val,
            allowed_tools=allowed_tools_val,
            tool_requirements=tool_reqs_val,
            category=category_val,
            tags=tags_val,
            trigger_phrases=triggers_val,
            execution_hints=execution_hints_val,
            metadata=meta_dict,
            references=references,
            examples=examples,
            path=str(skill_dir),
        )
    except Exception:
        return None


def load_all_skills(
    skills_root: Optional[Path] = None,
    skill_filter: Optional[List[str]] = None,
) -> Dict[str, SkillDefinition]:
    """Scans and loads skill directories in the workspace packages, entry points, and local folders."""
    root = skills_root or find_registry_root()
    filter_set = set(skill_filter) if skill_filter else None
    loaded: Dict[str, SkillDefinition] = {}

    if root.is_dir() and (root / "SKILL.md").is_file():
        single = load_skill_from_dir(root)
        if single:
            if not filter_set or single.name in filter_set:
                return {single.name: single}

    # 1. Scan workspace packages for skills
    packages_dir = root / "packages" if not root.name == "packages" else root
    if packages_dir.is_dir():
        for skill_dir in sorted(packages_dir.glob("skills-*/src/*/skills/*")):
            if skill_dir.is_dir() and not skill_dir.name.startswith("."):
                if filter_set and skill_dir.name not in filter_set:
                    continue
                skill_def = load_skill_from_dir(skill_dir)
                if skill_def:
                    if not filter_set or skill_def.name in filter_set:
                        loaded[skill_def.name] = skill_def

    # 2. Fallback scan for standalone skills directory (and subcategory folders)
    skills_dir = root / "skills" if not root.name == "skills" else root
    if skills_dir.is_dir():
        for skill_md in sorted(skills_dir.rglob("SKILL.md"), key=lambda p: p.parent.name):
            entry = skill_md.parent
            if entry.name in loaded:
                continue
            if filter_set and entry.name not in filter_set:
                continue
            skill_def = load_skill_from_dir(entry)
            if skill_def:
                if not filter_set or skill_def.name in filter_set:
                    loaded[skill_def.name] = skill_def

    # 3. Scan standard cross-client .agents/skills directories (project-level & user-level)
    agents_dirs = [
        root / ".agents" / "skills",
        Path.home() / ".agents" / "skills",
    ]
    for ag_dir in agents_dirs:
        if ag_dir.is_dir():
            for entry in sorted(ag_dir.iterdir(), key=lambda p: p.name):
                if entry.is_dir() and not entry.name.startswith("."):
                    if entry.name in loaded:
                        continue
                    if filter_set and entry.name not in filter_set:
                        continue
                    skill_def = load_skill_from_dir(entry)
                    if skill_def:
                        if not filter_set or skill_def.name in filter_set:
                            loaded[skill_def.name] = skill_def

    # 4. Load skills installed via Python entry points in site-packages (only at root workspace scan)
    if skills_root is None:
        ep_skills = load_skills_from_entry_points(skill_filter=skill_filter)
        for name, sdef in ep_skills.items():
            if name not in loaded:
                loaded[name] = sdef

    return loaded


def load_skills_from_entry_points(
    group: str = "retailcortex.skills",
    skill_filter: Optional[List[str]] = None,
) -> Dict[str, SkillDefinition]:
    """Discovers and loads enterprise skills installed via Python entry points in site-packages."""
    loaded: Dict[str, SkillDefinition] = {}
    try:
        eps = importlib.metadata.entry_points(group=group)
        for ep in eps:
            try:
                mod = importlib.import_module(ep.value)
                if hasattr(mod, "__file__") and mod.__file__:
                    mod_path = Path(mod.__file__).parent
                    skills_sub = mod_path / "skills"
                    target = skills_sub if skills_sub.is_dir() else mod_path
                    skills = load_all_skills(target, skill_filter=skill_filter)
                    loaded.update(skills)
            except Exception:
                pass
    except Exception:
        pass
    return loaded


def load_skills_from_package(
    package_name: str,
    skill_filter: Optional[List[str]] = None,
) -> Dict[str, SkillDefinition]:
    """Loads enterprise skills from an installed or importable Python package.

    Scans for a `skills/` sub-directory within the imported package directory first.
    If no `skills/` sub-directory exists, it falls back to scanning the package root directory directly.
    """
    clean_pkg = package_name.strip()
    if not clean_pkg:
        return {}

    loaded: Dict[str, SkillDefinition] = {}

    # 1. Try direct import
    try:
        mod = importlib.import_module(clean_pkg)
        if hasattr(mod, "__file__") and mod.__file__:
            mod_path = Path(mod.__file__).parent
            skills_sub = mod_path / "skills"
            target = skills_sub if skills_sub.is_dir() else mod_path
            skills = load_all_skills(target, skill_filter=skill_filter)
            if skills:
                return skills
    except Exception:
        pass

    # 2. Try entry points lookup if package_name matches group name
    if clean_pkg in ("retailcortex.skills", "skills"):
        return load_skills_from_entry_points(group=clean_pkg, skill_filter=skill_filter)

    # 3. Workspace package fallback discovery
    root = find_registry_root()
    packages_dir = root / "packages" if not root.name == "packages" else root
    if packages_dir.is_dir():
        for p in sorted(packages_dir.glob("skills-*")):
            src_dir = p / "src"
            if src_dir.is_dir():
                for pkg_dir in src_dir.iterdir():
                    if pkg_dir.name == clean_pkg and pkg_dir.is_dir():
                        skills_sub = pkg_dir / "skills"
                        target = skills_sub if skills_sub.is_dir() else pkg_dir
                        skills = load_all_skills(target, skill_filter=skill_filter)
                        loaded.update(skills)

    return loaded


def load_skills_from_github(
    repo: str,
    ref: Optional[str] = None,
    roots: Optional[List[str]] = None,
    skill_filter: Optional[List[str]] = None,
    github_token: Optional[str] = None,
    dotenv_path: Optional[Union[Path, str]] = None,
) -> Dict[str, SkillDefinition]:
    """Loads enterprise skills from a remote GitHub repository at a specific tag, branch, or revision commit SHA."""
    env_file = Path(dotenv_path) if dotenv_path else (Path.cwd() / ".env")
    dotenv_vars = parse_dotenv_file(env_file)

    clean_repo = repo.strip()
    if clean_repo.startswith("https://github.com/"):
        clean_repo = clean_repo[len("https://github.com/"):].rstrip("/")
    if clean_repo.endswith(".git"):
        clean_repo = clean_repo[:-4]

    url_roots: Optional[List[str]] = None
    if "/tree/" in clean_repo:
        parts = clean_repo.split("/tree/", 1)
        clean_repo = parts[0].rstrip("/")
        tree_parts = parts[1].split("/", 1)
        parsed_ref = tree_parts[0]
        if not ref:
            ref = parsed_ref
        if len(tree_parts) > 1 and tree_parts[1].strip():
            url_roots = [tree_parts[1].strip()]

    token = (
        github_token
        or os.environ.get("GITHUB_TOKEN")
        or os.environ.get("GH_TOKEN")
        or dotenv_vars.get("GITHUB_TOKEN")
        or dotenv_vars.get("GH_TOKEN")
    )
    git_ref = (
        ref
        or os.environ.get("GITHUB_REF")
        or dotenv_vars.get("GITHUB_REF")
        or "main"
    )

    if roots:
        root_paths = roots
    elif url_roots:
        root_paths = url_roots
    elif "SKILLS_ROOTS" in os.environ:
        root_paths = [r.strip() for r in os.environ["SKILLS_ROOTS"].split(",") if r.strip()]
    elif "SKILLS_ROOTS" in dotenv_vars:
        root_paths = [r.strip() for r in dotenv_vars["SKILLS_ROOTS"].split(",") if r.strip()]
    else:
        root_paths = ["skills", "."]

    if skill_filter:
        selected_skills = skill_filter
    elif "SKILLS_FILTER" in os.environ:
        selected_skills = [s.strip() for s in os.environ["SKILLS_FILTER"].split(",") if s.strip()]
    elif "SKILLS_FILTER" in dotenv_vars:
        selected_skills = [s.strip() for s in dotenv_vars["SKILLS_FILTER"].split(",") if s.strip()]
    else:
        selected_skills = None

    loaded_skills: Dict[str, SkillDefinition] = {}
    loader_base = get_loader_skills_dir()
    repo_slug = clean_repo.replace("/", "_")
    persistent_repo_dir = loader_base / "github" / repo_slug / git_ref
    persistent_repo_dir.mkdir(parents=True, exist_ok=True)

    with tempfile.TemporaryDirectory() as tmp_dir:
        tmp_path = Path(tmp_dir)

        cloned = False
        repo_dir = tmp_path / "repo"
        try:
            clone_url = f"https://github.com/{clean_repo}.git"
            if token:
                clone_url = f"https://x-access-token:{token}@github.com/{clean_repo}.git"

            cmd_clone = ["git", "clone", "--depth", "1", "--branch", git_ref, clone_url, str(repo_dir)]
            res = subprocess.run(cmd_clone, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
            if res.returncode == 0:
                cloned = True
            else:
                subprocess.run(["git", "init", str(repo_dir)], stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=False)
                subprocess.run(["git", "-C", str(repo_dir), "remote", "add", "origin", clone_url], check=False)
                res_fetch = subprocess.run(["git", "-C", str(repo_dir), "fetch", "--depth", "1", "origin", git_ref], stdout=subprocess.PIPE, stderr=subprocess.PIPE, check=False)
                if res_fetch.returncode == 0:
                    res_co = subprocess.run(["git", "-C", str(repo_dir), "checkout", "FETCH_HEAD"], check=False)
                    if res_co.returncode == 0:
                        cloned = True
        except Exception:
            cloned = False

        repo_target_dir = repo_dir if cloned else tmp_path

        if not cloned:
            archive_url = f"https://api.github.com/repos/{clean_repo}/zipball/{git_ref}"
            req = urllib.request.Request(archive_url)
            req.add_header("User-Agent", "skills-loader/1.0.0")
            if token:
                req.add_header("Authorization", f"token {token}")
            try:
                zip_path = tmp_path / "repo.zip"
                with urllib.request.urlopen(req) as resp, open(zip_path, "wb") as out:
                    out.write(resp.read())
                with zipfile.ZipFile(zip_path, "r") as zip_ref:
                    zip_ref.extractall(tmp_path)
                for item in tmp_path.iterdir():
                    if item.is_dir() and item.name != "repo":
                        repo_target_dir = item
                        break
            except Exception as e:
                if any(persistent_repo_dir.iterdir()):
                    repo_target_dir = persistent_repo_dir
                else:
                    raise RuntimeError(f"Failed to load GitHub repo '{clean_repo}' at revision '{git_ref}': {e}")

        # Mirror downloaded tree into .loader_skills directory for persistent reference
        if repo_target_dir.is_dir() and repo_target_dir != persistent_repo_dir:
            for item in repo_target_dir.iterdir():
                dest = persistent_repo_dir / item.name
                if item.is_dir():
                    if dest.exists():
                        shutil.rmtree(dest)
                    shutil.copytree(item, dest)
                else:
                    shutil.copy2(item, dest)
            repo_target_dir = persistent_repo_dir

        for root_rel in root_paths:
            candidate_dir = repo_target_dir / root_rel if root_rel != "." else repo_target_dir
            if candidate_dir.is_dir():
                single_skill = load_skill_from_dir(candidate_dir)
                if single_skill:
                    if not selected_skills or single_skill.name in selected_skills:
                        loaded_skills[single_skill.name] = single_skill
                else:
                    skills = load_all_skills(candidate_dir, skill_filter=selected_skills)
                    loaded_skills.update(skills)

    return loaded_skills


def load_skills_from_maven(
    coordinate: str,
    version: Optional[str] = None,
    roots: Optional[List[str]] = None,
    skill_filter: Optional[List[str]] = None,
) -> Dict[str, SkillDefinition]:
    """Resolves and loads skills from a Maven artifact coordinate in ~/.m2/repository."""
    loaded: Dict[str, SkillDefinition] = {}
    m2_repo = Path.home() / ".m2" / "repository"
    if not m2_repo.is_dir():
        return loaded

    parts = coordinate.split(":")
    if len(parts) < 2:
        return loaded

    group_path = parts[0].replace(".", "/")
    artifact = parts[1]
    ver = version or (parts[2] if len(parts) > 2 else "1.0.0")

    artifact_dir = m2_repo / group_path / artifact / ver
    if not artifact_dir.is_dir():
        return loaded

    sub_targets = roots or [""]
    for sub in sub_targets:
        target_dir = artifact_dir / sub if sub else artifact_dir
        if target_dir.is_dir():
            skills = load_all_skills(target_dir, skill_filter=skill_filter)
            loaded.update(skills)

    return loaded


def load_skills_from_go_module(
    module_path: str,
    version: Optional[str] = None,
    roots: Optional[List[str]] = None,
    skill_filter: Optional[List[str]] = None,
) -> Dict[str, SkillDefinition]:
    """Resolves and loads skills from a Go module in $GOPATH/pkg/mod."""
    loaded: Dict[str, SkillDefinition] = {}
    gopath = os.environ.get("GOPATH")
    if not gopath:
        gopath = str(Path.home() / "go")

    mod_dir = Path(gopath) / "pkg" / "mod"
    if not mod_dir.is_dir():
        return loaded

    ver = version or "latest"
    mod_target = mod_dir / f"{module_path.lower()}@{ver}"
    if not mod_target.is_dir():
        matches = sorted(mod_dir.glob(f"{module_path.lower()}@*"))
        if matches:
            mod_target = matches[-1]
        else:
            return loaded

    sub_targets = roots or [""]
    for sub in sub_targets:
        target_dir = mod_target / sub if sub else mod_target
        if target_dir.is_dir():
            skills = load_all_skills(target_dir, skill_filter=skill_filter)
            loaded.update(skills)

    return loaded


def load_skills_from_roots(
    roots: Optional[List[str]] = None,
    skill_filter: Optional[List[str]] = None,
    github_token: Optional[str] = None,
    dotenv_path: Optional[Union[Path, str]] = None,
) -> Dict[str, SkillDefinition]:
    """Loads enterprise skills across multiple qualified `file://`, `pkg://`, `github://`, `maven://`, and `mod://` root URIs."""
    env_file = Path(dotenv_path) if dotenv_path else (Path.cwd() / ".env")
    dotenv_vars = parse_dotenv_file(env_file)

    if roots is not None:
        root_uris = roots
    elif "SKILLS_ROOTS" in os.environ:
        root_uris = [r.strip() for r in os.environ["SKILLS_ROOTS"].split(",") if r.strip()]
    elif "SKILLS_ROOTS" in dotenv_vars:
        root_uris = [r.strip() for r in dotenv_vars["SKILLS_ROOTS"].split(",") if r.strip()]
    else:
        root_uris = ["file://."]

    if skill_filter:
        selected_skills = skill_filter
    elif "SKILLS_FILTER" in os.environ:
        selected_skills = [s.strip() for s in os.environ["SKILLS_FILTER"].split(",") if s.strip()]
    elif "SKILLS_FILTER" in dotenv_vars:
        selected_skills = [s.strip() for s in dotenv_vars["SKILLS_FILTER"].split(",") if s.strip()]
    else:
        selected_skills = None

    loaded: Dict[str, SkillDefinition] = {}

    for uri in root_uris:
        scheme, target, ref, subpath = parse_skill_root_uri(uri)
        if scheme == "file":
            p = Path(target)
            if not p.is_absolute():
                base_root = find_registry_root()
                p_cand = (base_root / target).resolve()
                p = p_cand if p_cand.exists() else p.resolve()
            if p.is_dir():
                single = load_skill_from_dir(p)
                if single:
                    if not selected_skills or single.name in selected_skills:
                        loaded[single.name] = single
                else:
                    skills = load_all_skills(p, skill_filter=selected_skills)
                    loaded.update(skills)
        elif scheme == "pkg":
            skills = load_skills_from_package(target, skill_filter=selected_skills)
            loaded.update(skills)
        elif scheme == "github":
            gh_roots = [subpath] if subpath else None
            skills = load_skills_from_github(
                repo=target,
                ref=ref,
                roots=gh_roots,
                skill_filter=selected_skills,
                github_token=github_token,
                dotenv_path=dotenv_path,
            )
            loaded.update(skills)
        elif scheme in ("maven", "mvn"):
            mvn_roots = [subpath] if subpath else None
            skills = load_skills_from_maven(
                coordinate=target,
                version=ref,
                roots=mvn_roots,
                skill_filter=selected_skills,
            )
            loaded.update(skills)
        elif scheme in ("mod", "go"):
            mod_roots = [subpath] if subpath else None
            skills = load_skills_from_go_module(
                module_path=target,
                version=ref,
                roots=mod_roots,
                skill_filter=selected_skills,
            )
            loaded.update(skills)

    return loaded


def build_skills_manifest(
    skills_root: Optional[Path] = None,
    output_path: Optional[Path] = None,
) -> Path:
    """Builds a pre-compiled JSON manifest of all skills for fast zero-I/O loading."""
    skills = load_all_skills(skills_root)
    out_file = output_path or (get_loader_skills_dir() / "skills_manifest.json")
    manifest_data = {
        name: {
            "name": s.name,
            "description": s.description,
            "instructions": s.instructions,
            "license": s.license,
            "author": s.author,
            "version": s.version,
            "compatibility": s.compatibility,
            "allowed_tools": s.allowed_tools,
            "metadata": s.metadata,
            "references": s.references,
            "examples": s.examples,
            "path": s.path,
        }
        for name, s in skills.items()
    }
    out_file.parent.mkdir(parents=True, exist_ok=True)
    out_file.write_text(json.dumps(manifest_data, indent=2), encoding="utf-8")
    return out_file


def load_skills_from_manifest(manifest_path: Path) -> Dict[str, SkillDefinition]:
    """Loads skill definitions directly from a pre-compiled JSON manifest file."""
    if not manifest_path.is_file():
        return {}
    try:
        content = json.loads(manifest_path.read_text(encoding="utf-8"))
        loaded: Dict[str, SkillDefinition] = {}
        for name, data in content.items():
            loaded[name] = SkillDefinition(
                name=data["name"],
                description=data["description"],
                instructions=data["instructions"],
                license=data.get("license"),
                author=data.get("author"),
                version=data.get("version"),
                compatibility=data.get("compatibility"),
                allowed_tools=data.get("allowed_tools"),
                metadata=data.get("metadata", {}),
                references=data.get("references", {}),
                examples=data.get("examples", {}),
                path=data.get("path", ""),
            )
        return loaded
    except Exception:
        return {}


class SkillRegistry:
    """High-performance registry for discovering and querying enterprise skills in Google ADK agents."""

    def __init__(
        self,
        skills_root: Optional[Union[Path, str]] = None,
        roots: Optional[List[str]] = None,
        skill_filter: Optional[List[str]] = None,
        dotenv_path: Optional[Union[Path, str]] = None,
    ) -> None:
        self.root: Path = Path(skills_root).resolve() if isinstance(skills_root, (Path, str)) else find_registry_root()

        if roots:
            root_list = roots
        elif skills_root:
            root_str = str(skills_root)
            root_list = [root_str]
        else:
            root_list = None

        self._skills: Dict[str, SkillDefinition] = load_skills_from_roots(
            roots=root_list,
            skill_filter=skill_filter,
            dotenv_path=dotenv_path,
        )
        self.compiler = SkillCompiler()
        self.discovery_engine = SkillDiscoveryEngine(compiler=self.compiler)
        self.hitl_engine = HITLEngine()
        self._search_cache: Dict[str, List[SkillDefinition]] = {}
        self._domain_cache: Dict[str, Dict[str, Any]] = {}

        # Automatically index loaded skills
        for s in self._skills.values():
            self.discovery_engine.register_skill(s)
        self.discovery_engine.build_index()

    @classmethod
    def from_github(
        cls,
        repo: str,
        ref: Optional[str] = None,
        roots: Optional[List[str]] = None,
        skill_filter: Optional[List[str]] = None,
        github_token: Optional[str] = None,
        dotenv_path: Optional[Union[Path, str]] = None,
    ) -> "SkillRegistry":
        """Instantiates a SkillRegistry populated with select skills fetched from a remote GitHub repository."""
        instance = cls.__new__(cls)
        instance.root = Path.cwd().resolve()
        instance._skills = load_skills_from_github(
            repo=repo,
            ref=ref,
            roots=roots,
            skill_filter=skill_filter,
            github_token=github_token,
            dotenv_path=dotenv_path,
        )
        instance._search_cache = {}
        instance._domain_cache = {}
        instance.compiler = SkillCompiler()
        instance.discovery_engine = SkillDiscoveryEngine(compiler=instance.compiler)
        instance.hitl_engine = HITLEngine()

        for s in instance._skills.values():
            instance.discovery_engine.register_skill(s)
        instance.discovery_engine.build_index()
        return instance

    @classmethod
    def from_roots(
        cls,
        roots: List[str],
        skill_filter: Optional[List[str]] = None,
        github_token: Optional[str] = None,
        dotenv_path: Optional[Union[Path, str]] = None,
    ) -> "SkillRegistry":
        """Instantiates a SkillRegistry populated from qualified file://, pkg://, and github:// root URIs."""
        instance = cls.__new__(cls)
        instance.root = Path.cwd().resolve()
        instance._skills = load_skills_from_roots(
            roots=roots,
            skill_filter=skill_filter,
            github_token=github_token,
            dotenv_path=dotenv_path,
        )
        instance._search_cache = {}
        instance._domain_cache = {}
        instance.compiler = SkillCompiler()
        instance.discovery_engine = SkillDiscoveryEngine(compiler=instance.compiler)
        instance.hitl_engine = HITLEngine()

        for s in instance._skills.values():
            instance.discovery_engine.register_skill(s)
        instance.discovery_engine.build_index()
        return instance

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
            compiled_ref = s.compiled_reference or self.compiler.compile(s)
            summaries.append(
                SkillSummary(
                    name=s.name,
                    description=s.description,
                    reference_count=len(s.references),
                    example_count=len(s.examples),
                    path=s.path,
                    category=s.category,
                    tags=s.tags,
                    trigger_phrases=s.trigger_phrases,
                    sha256_hash=compiled_ref.sha256_hash,
                    hitl_tier=compiled_ref.hitl_tier,
                )
            )
        return summaries

    def compile_all(
        self,
        strict_schemas: bool = True,
        allow_additional_properties: Optional[List[str]] = None,
    ) -> Dict[str, CompiledSkillReference]:
        """Compiles all registered skills using custom compiler options."""
        compiler = SkillCompiler(
            strict_schemas=strict_schemas,
            allow_additional_properties=allow_additional_properties,
        )
        compiled_map: Dict[str, CompiledSkillReference] = {}
        for s in self._skills.values():
            compiled_ref = compiler.compile(s)
            compiled_map[s.name] = compiled_ref
        return compiled_map

    def search_intent(self, intent: str, top_k: int = 5) -> SkillDirectorySearchResult:
        """Executes natural language JIT discovery search using local TF-IDF index."""
        return self.discovery_engine.search_skills(intent, top_k=top_k)

    def evaluate_execution_gate(
        self,
        skill_name: str,
        runtime_params: Optional[Dict[str, Any]] = None,
        force_skip_hitl: bool = False,
    ) -> HITLGateResult:
        """Evaluates HITL execution policy gate for a given skill."""
        skill = self.get(skill_name)
        if not skill:
            return HITLGateResult(
                approved=False,
                tier=HITLPolicyTier.TIER_3_MANDATORY_APPROVAL,
                reason=f"Skill '{skill_name}' not found in registry.",
            )
        ref = skill.compiled_reference or self.compiler.compile(skill)
        return self.hitl_engine.evaluate_gate(
            skill=skill,
            ref=ref,
            runtime_params=runtime_params,
            force_skip_hitl=force_skip_hitl,
        )

    def search(self, query: str) -> List[SkillDefinition]:
        """Searches skills by keyword matching in name, description, instructions, or references with memoization."""
        q = query.lower()
        if q in self._search_cache:
            return self._search_cache[q]

        results: List[SkillDefinition] = []
        for s in self._skills.values():
            if (
                q in s.name.lower()
                or q in s.description.lower()
                or q in s.instructions.lower()
                or any(q in r_text.lower() for r_text in s.references.values())
            ):
                results.append(s)
        self._search_cache[q] = results
        return results

    def get_domain_skills(self, domain: str) -> List[SkillDefinition]:
        """Retrieves skills matching a specific language or domain with memoization."""
        domain_norm = domain.lower()
        if domain_norm in self._domain_cache:
            return self._domain_cache[domain_norm]

        res = [
            s for s in self._skills.values()
            if domain_norm in s.name.lower() or domain_norm in s.description.lower()
        ]
        self._domain_cache[domain_norm] = res
        return res


def calculate_skill_checksum(skill_dir: Union[Path, str]) -> str:
    """Calculates a deterministic SHA256 checksum of a skill directory's contents."""
    s_dir = Path(skill_dir).resolve()
    if not s_dir.is_dir():
        raise ValueError(f"Skill directory does not exist: {s_dir}")

    hasher = hashlib.sha256()
    files: List[Tuple[str, Path]] = []
    for p in s_dir.rglob("*"):
        if p.is_file() and p.name not in (".DS_Store", ".manifest.lock"):
            rel = p.relative_to(s_dir).as_posix()
            files.append((rel, p))

    files.sort(key=lambda x: x[0])
    for rel, p in files:
        hasher.update(rel.encode("utf-8"))
        hasher.update(p.read_bytes())

    return hasher.hexdigest()


def read_manifest_lock(dest_dir: Union[Path, str]) -> Dict[str, Any]:
    """Reads the .manifest.lock file from a target destination directory."""
    d_dir = Path(dest_dir).resolve()
    lock_file = d_dir / ".manifest.lock"
    if not lock_file.is_file():
        raise FileNotFoundError(f".manifest.lock file not found in {d_dir}")

    try:
        content = json.loads(lock_file.read_text(encoding="utf-8"))
        if not isinstance(content, dict):
            return {"version": "1.0.0", "skills": {}}
        if "skills" not in content or not isinstance(content["skills"], dict):
            content["skills"] = {}
        return content
    except Exception as e:
        raise ValueError(f"Failed to parse .manifest.lock in {d_dir}: {e}")


def write_manifest_lock(dest_dir: Union[Path, str], lock_data: Dict[str, Any]) -> Path:
    """Writes the .manifest.lock file to a target destination directory."""
    d_dir = Path(dest_dir).resolve()
    d_dir.mkdir(parents=True, exist_ok=True)
    lock_file = d_dir / ".manifest.lock"

    if "version" not in lock_data:
        lock_data["version"] = "1.0.0"
    if "skills" not in lock_data or not isinstance(lock_data["skills"], dict):
        lock_data["skills"] = {}

    lock_file.write_text(json.dumps(lock_data, indent=2), encoding="utf-8")
    return lock_file


def update_manifest_lock(
    dest_dir: Union[Path, str],
    skill_name: str,
    uri: str,
    checksum: Optional[str] = None,
) -> Path:
    """Updates or adds a skill entry in the destination directory's .manifest.lock file."""
    d_dir = Path(dest_dir).resolve()
    try:
        lock_data = read_manifest_lock(d_dir)
    except Exception:
        lock_data = {"version": "1.0.0", "skills": {}}

    if not checksum:
        skill_dir = d_dir / skill_name
        checksum = calculate_skill_checksum(skill_dir)

    lock_data["skills"][skill_name] = {
        "skill_name": skill_name,
        "uri": uri,
        "sha256": checksum,
    }

    return write_manifest_lock(d_dir, lock_data)


def verify_manifest_lock(dest_dir: Union[Path, str]) -> Dict[str, Any]:
    """Validates that skills present in .manifest.lock match their original recorded checksums."""
    d_dir = Path(dest_dir).resolve()
    lock_data = read_manifest_lock(d_dir)

    skills_map: Dict[str, Dict[str, Any]] = lock_data.get("skills", {})
    results: List[Dict[str, Any]] = []
    verified_count = 0
    modified_count = 0
    missing_count = 0

    for name in sorted(skills_map.keys()):
        entry = skills_map[name]
        skill_dir = d_dir / name
        expected_sha = entry.get("sha256", "")
        uri = entry.get("uri", "")

        if not skill_dir.is_dir():
            missing_count += 1
            results.append({
                "skill_name": name,
                "uri": uri,
                "status": "missing",
                "expected_sha256": expected_sha,
                "error": "skill directory missing",
            })
            continue

        try:
            current_sha = calculate_skill_checksum(skill_dir)
            if current_sha == expected_sha:
                verified_count += 1
                results.append({
                    "skill_name": name,
                    "uri": uri,
                    "status": "verified",
                    "expected_sha256": expected_sha,
                    "actual_sha256": current_sha,
                })
            else:
                modified_count += 1
                results.append({
                    "skill_name": name,
                    "uri": uri,
                    "status": "modified",
                    "expected_sha256": expected_sha,
                    "actual_sha256": current_sha,
                    "error": "checksum mismatch (skill files modified)",
                })
        except Exception as e:
            modified_count += 1
            results.append({
                "skill_name": name,
                "uri": uri,
                "status": "modified",
                "expected_sha256": expected_sha,
                "error": f"failed to compute checksum: {e}",
            })

    return {
        "target_dir": str(d_dir),
        "total_skills": len(skills_map),
        "verified_count": verified_count,
        "modified_count": modified_count,
        "missing_count": missing_count,
        "results": results,
    }

