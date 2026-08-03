"""Agentic Polyglot Developer CLI using Google ADK & Retail Cortex Skills via skills-loader."""

import argparse
import os
import sys
from pathlib import Path
from typing import Dict, List, Optional

# Add project packages to sys.path for direct execution
project_root = Path(__file__).resolve().parents[2]
loader_src = project_root / "packages/loader/src"
bazel_src = project_root / "examples/skills-bazel/src"
go_src = project_root / "examples/skills-go/src"
java_src = project_root / "examples/skills-java/src"
proto_src = project_root / "examples/skills-protobuf/src"
py_src = project_root / "examples/skills-python/src"
frontend_src = project_root / "examples/skills-frontend/src"

for src_path in [loader_src, bazel_src, go_src, java_src, proto_src, py_src, frontend_src]:
    if str(src_path) not in sys.path:
        sys.path.insert(0, str(src_path))

from loader import (
    SkillDefinition,
    SkillRegistry,
    load_skills_from_roots,
    parse_dotenv_file,
    parse_skill_root_uri,
)


class PolyglotDeveloperAgent:
    """Agentic CLI developer agent that loads domain skills to scaffold polyglot Bazel projects."""

    def __init__(
        self,
        registry: Optional[SkillRegistry] = None,
        dotenv_path: Optional[Path] = None,
    ) -> None:
        env_file = dotenv_path or (Path(__file__).parent / ".env")
        dotenv_vars = parse_dotenv_file(env_file)

        if registry:
            self.registry: SkillRegistry = registry
        elif "SKILLS_ROOTS" in os.environ or "SKILLS_ROOTS" in dotenv_vars:
            self.registry = SkillRegistry.from_roots(dotenv_path=env_file)
        else:
            self.registry = SkillRegistry.from_roots(
                roots=[
                    "pkg://retailcortex_skills_bazel",
                    "pkg://retailcortex_skills_go",
                    "pkg://retailcortex_skills_java",
                    "pkg://retailcortex_skills_protobuf",
                    "pkg://retailcortex_skills_python",
                    "pkg://retailcortex_skills_frontend",
                ]
            )

        self.required_skills: List[str] = [
            "bazel-modules",
            "go-lang",
            "java-enterprise",
            "protobuf-grpc",
            "python-core",
            "python-adk-fastapi",
            "react-vite",
        ]

    def verify_skills_available(self) -> Dict[str, bool]:
        """Verifies all polyglot skill packages are correctly loaded in the registry."""
        status: Dict[str, bool] = {}
        for sname in self.required_skills:
            status[sname] = self.registry.get(sname) is not None
        return status

    def get_skill_guidance(self, skill_name: str) -> str:
        """Retrieves and formats guidelines from a loaded SkillDefinition."""
        skill = self.registry.get(skill_name)
        if not skill:
            return f"# Skill '{skill_name}' not loaded."
        return f"# Guidance from skill '{skill.name}' ({skill.path}):\n# {skill.description}\n"

    def scaffold_polyglot_project(self, target_dir: Path, project_name: str = "polyglot-app") -> List[Path]:
        """Generates a polyglot Bazel monorepo structure grounded in loaded skill definitions."""
        target_dir = target_dir.resolve()
        created_files: List[Path] = []

        # 1. Fetch skill definitions from registry via skills-loader
        bazel_skill = self.registry.get("bazel-modules")
        proto_skill = self.registry.get("protobuf-grpc")
        go_skill = self.registry.get("go-lang")
        java_skill = self.registry.get("java-enterprise")
        py_skill = self.registry.get("python-adk-fastapi")
        fe_skill = self.registry.get("react-vite")

        # 2. Root MODULE.bazel (Grounded in bazel-modules skill definition)
        bazel_header = self.get_skill_guidance("bazel-modules")
        module_content = f"""{bazel_header}module(
    name = "{project_name}",
    version = "1.0.0",
)

# Rule sets for polyglot builds derived from skills-bazel
bazel_dep(name = "rules_python", version = "0.34.0")
bazel_dep(name = "rules_go", version = "0.49.0")
bazel_dep(name = "rules_jvm_external", version = "6.2")
bazel_dep(name = "rules_proto", version = "6.0.2")
"""
        mod_file = target_dir / "MODULE.bazel"
        mod_file.parent.mkdir(parents=True, exist_ok=True)
        mod_file.write_text(module_content, encoding="utf-8")
        created_files.append(mod_file)

        # 3. Root BUILD.bazel
        root_build = f"""# Polyglot Monorepo Root Build Target
# Loaded from skill: {bazel_skill.name if bazel_skill else 'bazel-modules'}
package(default_visibility = ["//visibility:public"])

filegroup(
    name = "all_services",
    srcs = [
        "//proto:user_proto",
        "//services/go-service:go_service",
        "//services/java-service:java_service",
        "//services/python-service:python_service",
        "//apps/web-dashboard:web_dashboard",
    ],
)
"""
        root_build_file = target_dir / "BUILD.bazel"
        root_build_file.write_text(root_build, encoding="utf-8")
        created_files.append(root_build_file)

        # 4. Protobuf Contract (Grounded in protobuf-grpc skill definition)
        proto_header = self.get_skill_guidance("protobuf-grpc")
        proto_file = target_dir / "proto" / "user.proto"
        proto_file.parent.mkdir(parents=True, exist_ok=True)
        proto_file.write_text(f"""{proto_header}syntax = "proto3";

package polyglot.v1;

option go_package = "github.com/polyglot/proto/v1;protov1";
option java_package = "com.polyglot.proto.v1";

message UserProfile {{
  string user_id = 1;
  string display_name = 2;
  string email = 3;
}}

service UserService {{
  rpc GetUser (UserProfile) returns (UserProfile);
}}
""", encoding="utf-8")
        created_files.append(proto_file)

        # 5. Go Service (Grounded in go-lang skill definition)
        go_header = self.get_skill_guidance("go-lang")
        go_file = target_dir / "services" / "go-service" / "main.go"
        go_file.parent.mkdir(parents=True, exist_ok=True)
        go_file.write_text(f"""package main

// {go_header.strip()}

import (
	"fmt"
)

func main() {{
	fmt.Println("Polyglot Go Microservice initializing...")
}}
""", encoding="utf-8")
        created_files.append(go_file)

        # 6. Java Service (Grounded in java-enterprise skill definition)
        java_header = self.get_skill_guidance("java-enterprise")
        java_file = target_dir / "services" / "java-service" / "src" / "main" / "java" / "com" / "polyglot" / "Application.java"
        java_file.parent.mkdir(parents=True, exist_ok=True)
        java_file.write_text(f"""package com.polyglot;

// {java_header.strip()}

public class Application {{
    public static void main(String[] args) {{
        System.out.println("Polyglot Java Enterprise Service running...");
    }}
}}
""", encoding="utf-8")
        created_files.append(java_file)

        # 7. Python Service (Grounded in python-adk-fastapi skill definition)
        py_header = self.get_skill_guidance("python-adk-fastapi")
        py_file = target_dir / "services" / "python-service" / "main.py"
        py_file.parent.mkdir(parents=True, exist_ok=True)
        py_file.write_text(f"""\"\"\"Polyglot Python FastAPI / ADK Agent Service.

{py_header.strip()}
\"\"\"

from fastapi import FastAPI

app = FastAPI(title="Polyglot Agent Service")

@app.get("/health")
def health_check() -> dict[str, str]:
    return {{"status": "ok", "service": "python-agent"}}
""", encoding="utf-8")
        created_files.append(py_file)

        # 8. Frontend Web Dashboard (Grounded in react-vite skill definition)
        fe_header = self.get_skill_guidance("react-vite")
        fe_file = target_dir / "apps" / "web-dashboard" / "src" / "App.tsx"
        fe_file.parent.mkdir(parents=True, exist_ok=True)
        fe_file.write_text(f"""// {fe_header.strip()}
import React from 'react';

export const App: React.FC = () => {{
  return (
    <div className="p-4">
      <h1 className="text-2xl font-bold">Polyglot Bazel Web Dashboard</h1>
    </div>
  );
}};
""", encoding="utf-8")
        created_files.append(fe_file)

        return created_files


def main() -> None:
    """CLI entrypoint for Polyglot Developer Agent using skills-loader."""
    parser = argparse.ArgumentParser(description="Agentic CLI for scaffolding Bazel Polyglot Projects using Retail Cortex Skills via skills-loader.")
    parser.add_argument("--target-dir", type=str, default="./polyglot-output", help="Directory where the polyglot project will be generated.")
    parser.add_argument("--project-name", type=str, default="polyglot-workspace", help="Name of the generated Bazel module/workspace.")
    args = parser.parse_args()

    print("==================================================================")
    print(" Polyglot Developer Agent: Bazel Monorepo Generator (skills-loader)")
    print("==================================================================")

    agent = PolyglotDeveloperAgent()
    status = agent.verify_skills_available()

    print(f"\n1. Loaded {len(agent.registry.skills)} total skill(s) via SkillRegistry:")
    for summary in agent.registry.list_skills():
        print(f"   - [{summary.name}] ({summary.path})")

    print("\n2. Skill Package Verification Status:")
    for skill_name, is_loaded in status.items():
        state_str = "LOADED" if is_loaded else "MISSING"
        print(f"  - {skill_name:<25}: {state_str}")

    target_path = Path(args.target_dir)
    print(f"\n3. Scaffolding polyglot Bazel workspace in: {target_path}")
    created = agent.scaffold_polyglot_project(target_path, project_name=args.project_name)

    print(f"\nSuccessfully generated {len(created)} project files grounded in loaded SkillDefinition objects:")
    for p in created:
        rel_p = p.relative_to(target_path) if p.is_relative_to(target_path) else p
        print(f"  + {rel_p}")


if __name__ == "__main__":
    main()
