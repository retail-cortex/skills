import json
import os
import sys
from pathlib import Path
from typing import NoReturn

# Add src to sys.path
sys.path.insert(0, str(Path(__file__).parents[1]))

from validator.audit import audit_all_skills

def find_registry_root() -> Path:
    if "BUILD_WORKSPACE_DIRECTORY" in os.environ:
        return Path(os.environ["BUILD_WORKSPACE_DIRECTORY"])

    current: Path = Path(__file__).resolve().parent
    for parent in [current] + list(current.parents):
        if (parent / "skills").is_dir() or (parent / "examples" / "skills").is_dir():
            return parent
    return Path(__file__).parents[4]

def main() -> None:
    registry_root: Path = find_registry_root()

    print("\n" + "=" * 88)
    print(" AI Agent Skill Builder: Automated 5-Point SDLC & Security Audit")
    print("=" * 88)
    print(f"Scanning registry at: {registry_root}\n")

    summary = audit_all_skills(registry_root)

    header = f"{'Skill Name':<24} | {'Frontmatter':<11} | {'L3 Tree':<8} | {'Security':<8} | {'429 Rate':<8} | {'Links':<6} | {'Status':<6}"
    print(header)
    print("-" * len(header))

    for r in summary.results:
        fm_badge = "PASS" if r.frontmatter_valid else "FAIL"
        l3_badge = "PASS" if r.l3_tree_valid else "FAIL"
        cwe_badge = "PASS" if r.cwe_security_valid else "FAIL"
        r429_badge = "PASS" if r.rate_limit_429_valid else "FAIL"
        link_badge = "PASS" if r.clickable_links_valid else "FAIL"
        status_badge = "PASSED" if r.passed else "FAILED"

        row = f"{r.skill_name:<24} | {fm_badge:<11} | {l3_badge:<8} | {cwe_badge:<8} | {r429_badge:<8} | {link_badge:<6} | {status_badge:<6}"
        print(row)

    print("-" * len(header))
    print(f"\nTotal Skills: {summary.total_skills} | Passed: {summary.passed_skills} | Failed: {summary.failed_skills}\n")

    # Export report JSON
    try:
        report_file: Path = registry_root / "validator_report.json"
        with open(report_file, "w", encoding="utf-8") as f:
            json.dump(summary.to_dict(), f, indent=2)
        print(f"Saved audit report artifact to: {report_file}\n")
    except Exception:
        pass

    if summary.failed_skills > 0:
        print("ERROR: One or more skills failed the 5-point SDLC audit. Exiting with code 1.\n")
        sys.exit(1)
    else:
        print(f"SUCCESS: All {summary.passed_skills} enterprise skills passed the 5-point SDLC and security audit!\n")
        sys.exit(0)

if __name__ == "__main__":
    main()
