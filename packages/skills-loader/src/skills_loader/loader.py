"""Dynamic skill scanner and loader for enterprise AI agent skills compatible with Google ADK."""

import os
import re
import subprocess
import tempfile
import urllib.request
import zipfile
from pathlib import Path
from typing import Dict, List, Optional, Tuple, Union

from skills_loader.types import SkillDefinition, SkillSummary


def find_registry_root() -> Path:
    """Discovers the root workspace directory containing the skills folder."""
    if "BUILD_WORKSPACE_DIRECTORY" in os.environ:
        workspace = Path(os.environ["BUILD_WORKSPACE_DIRECTORY"])
        if (workspace / "skills").is_dir():
            return workspace

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
            if (cand / "skills").is_dir():
                return cand

        for cand in test_srcdir.rglob("skills"):
            if cand.is_dir() and any(cand.glob("*/SKILL.md")):
                return cand.parent

    current: Path = Path(__file__).resolve().parent
    for parent in [current] + list(current.parents):
        if (parent / "skills").is_dir():
            return parent.resolve()

    runfiles = os.environ.get("PYTHON_RUNFILES", "")
    if runfiles:
        rf_path = Path(runfiles)
        if (rf_path / "skills").is_dir():
            return rf_path
        for cand in rf_path.rglob("skills"):
            if cand.is_dir() and any(cand.glob("*/SKILL.md")):
                return cand.parent

    return Path.cwd().resolve()


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

    Supported github URI formats:
    - github://owner/repo/subpath:ref (e.g. github://google/skills/skills/cloud/gemini-api:main)
    - github://owner/repo:ref/subpath (e.g. github://google/skills:main/skills/cloud/gemini-api)
    - github://owner/repo@ref/subpath (e.g. github://google/skills@v1.2.0/skills/cloud)
    - github://owner/repo/tree/ref/subpath (e.g. github://google/skills/tree/main/skills/cloud)
    """
    clean = uri.strip()
    if clean.startswith("file://"):
        target = clean[len("file://"):]
        return "file", target, None, None
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

        references: Dict[str, str] = {}
        ref_dir = skill_dir / "references"
        if ref_dir.is_dir():
            for ref_file in sorted(ref_dir.glob("*.md")):
                try:
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
    """Scans and loads skill directories in the repository, optionally filtering select skills by name."""
    root = skills_root or find_registry_root()
    skills_dir = root / "skills" if not root.name == "skills" else root

    loaded: Dict[str, SkillDefinition] = {}
    if not skills_dir.is_dir():
        return loaded

    ignored = {".git", ".bazel", "packages", "validator", "node_modules", "scratch", "build", "dist", ".venv"}
    filter_set = set(skill_filter) if skill_filter else None

    for entry in sorted(skills_dir.iterdir(), key=lambda p: p.name):
        if entry.is_dir() and entry.name not in ignored and not entry.name.startswith("."):
            if filter_set and entry.name not in filter_set:
                continue
            skill_def = load_skill_from_dir(entry)
            if skill_def:
                if filter_set and skill_def.name not in filter_set:
                    continue
                loaded[skill_def.name] = skill_def

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
                raise RuntimeError(f"Failed to load GitHub repo '{clean_repo}' at revision '{git_ref}': {e}")

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


def load_skills_from_roots(
    roots: Optional[List[str]] = None,
    skill_filter: Optional[List[str]] = None,
    github_token: Optional[str] = None,
    dotenv_path: Optional[Union[Path, str]] = None,
) -> Dict[str, SkillDefinition]:
    """Loads enterprise skills across multiple qualified `file://` and `github://` root URIs."""
    env_file = Path(dotenv_path) if dotenv_path else (Path.cwd() / ".env")
    dotenv_vars = parse_dotenv_file(env_file)

    if roots:
        root_uris = roots
    elif "SKILLS_ROOTS" in os.environ:
        root_uris = [r.strip() for r in os.environ["SKILLS_ROOTS"].split(",") if r.strip()]
    elif "SKILLS_ROOTS" in dotenv_vars:
        root_uris = [r.strip() for r in dotenv_vars["SKILLS_ROOTS"].split(",") if r.strip()]
    else:
        root_uris = ["file://skills"]

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

    return loaded


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
        return instance

    @classmethod
    def from_roots(
        cls,
        roots: List[str],
        skill_filter: Optional[List[str]] = None,
        github_token: Optional[str] = None,
        dotenv_path: Optional[Union[Path, str]] = None,
    ) -> "SkillRegistry":
        """Instantiates a SkillRegistry populated from qualified file:// and github:// root URIs."""
        instance = cls.__new__(cls)
        instance.root = Path.cwd().resolve()
        instance._skills = load_skills_from_roots(
            roots=roots,
            skill_filter=skill_filter,
            github_token=github_token,
            dotenv_path=dotenv_path,
        )
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
