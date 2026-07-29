"""Native Google ADK Agent Example using qualified file:// and github:// skill roots via dotenv."""

import asyncio
import os
import sys
from pathlib import Path
from typing import AsyncGenerator, Dict, List, Optional

# Add project packages to sys.path for direct execution
project_root = Path(__file__).resolve().parents[2]
skills_loader_src = project_root / "packages/skills-loader/src"
skills_agent_src = project_root / "packages/skills-agent/src"
if str(skills_loader_src) not in sys.path:
    sys.path.insert(0, str(skills_loader_src))
if str(skills_agent_src) not in sys.path:
    sys.path.insert(0, str(skills_agent_src))

from skills_loader import SkillDefinition, SkillRegistry, parse_dotenv_file, parse_skill_root_uri


class ToolContext:
    """ADK ToolContext managing invocation state and Google OAuth2 IAM credentials."""

    def __init__(self, state: Optional[Dict[str, str]] = None) -> None:
        self.state: Dict[str, str] = state if state is not None else {}


class Session:
    """ADK Session tracking conversational memory and token delegation."""

    def __init__(self, id: str, app_name: str = "example-adk", user_id: str = "adk_user") -> None:
        self.id: str = id
        self.app_name: str = app_name
        self.user_id: str = user_id
        self.state: Dict[str, str] = {}
        self.history: List[Dict[str, str]] = []


class InMemorySessionService:
    """Asynchronous session store for ADK agent state management."""

    def __init__(self) -> None:
        self._sessions: Dict[str, Session] = {}

    async def get_session(self, session_id: str) -> Optional[Session]:
        """Retrieves a session by ID."""
        return self._sessions.get(session_id)

    async def create_or_get_session(self, session_id: str, user_token: Optional[str] = None) -> Session:
        """Retrieves or creates a session, delegating user tokens into session state."""
        if session_id not in self._sessions:
            self._sessions[session_id] = Session(id=session_id)
        session = self._sessions[session_id]
        if user_token:
            session.state["user_token"] = user_token
        return session


class ADKSkillToolset:
    """ADK Native Toolset powered by qualified file:// and github:// skill roots."""

    def __init__(self, registry: SkillRegistry) -> None:
        self.skills: Dict[str, SkillDefinition] = dict(registry.skills)

    def list_skills(self, tool_context: Optional[ToolContext] = None) -> List[Dict[str, object]]:
        """Lists all active skills across file:// and github:// roots."""
        return [s.to_dict() for s in self.skills.values()]

    def search_skills(self, query: str, tool_context: Optional[ToolContext] = None) -> List[Dict[str, object]]:
        """Searches skills by keyword across file:// and github:// definitions."""
        q = query.lower()
        matches: List[Dict[str, object]] = []
        for s in self.skills.values():
            if (
                q in s.name.lower()
                or q in s.description.lower()
                or q in s.instructions.lower()
                or any(q in ref.lower() for ref in s.references.values())
            ):
                matches.append(s.to_dict())
        return matches

    def generate_guidance(self, query: str, tool_context: Optional[ToolContext] = None) -> str:
        """Synthesizes guidelines from matching qualified skills into agent prompt context."""
        matches = self.search_skills(query, tool_context)
        if not matches:
            return f"No direct skill match found for '{query}'. Applying baseline ADK SDLC rules."

        lines: List[str] = [f"Found {len(matches)} matching skills for '{query}':\n"]
        for s in matches[:5]:
            lines.append(f"### Skill: {s['name']}")
            lines.append(f"**Description**: {s['description']}")
            lines.append(f"**Source Path**: {s['path']}\n")
        return "\n".join(lines)


class ADKNativeAgent:
    """Google ADK Native Programming Agent powered by qualified skill root URIs."""

    def __init__(self, toolset: ADKSkillToolset) -> None:
        self.name: str = "adk_native_agent"
        self.toolset: ADKSkillToolset = toolset

    async def execute_async(self, prompt: str, session: Session) -> AsyncGenerator[str, None]:
        """Executes agent prompt asynchronously, yielding response chunks."""
        tool_ctx = ToolContext(state=session.state)
        token = tool_ctx.state.get("user_token", "anonymous")

        guidance = self.toolset.generate_guidance(prompt, tool_ctx)

        loaded_skill_names = ", ".join([f"`{k}`" for k in self.toolset.skills.keys()])
        response_text = (
            f"--- ADK Native Agent Invocation ---\n"
            f"User Token: {token}\n"
            f"Active Qualified Skills: {loaded_skill_names}\n\n"
            f"{guidance}\n"
            f"**Execution Plan:**\n"
            f"1. Unified file:// and github:// qualified roots ({len(self.toolset.skills)} skills active).\n"
            f"2. Executed request: '{prompt}' under strict ADK invariants.\n"
        )

        session.history.append({"role": "user", "content": prompt})
        session.history.append({"role": "assistant", "content": response_text})

        chunk_size = 48
        for i in range(0, len(response_text), chunk_size):
            yield response_text[i : i + chunk_size]
            await asyncio.sleep(0.01)


async def main() -> None:
    print("=" * 88)
    print(" Native Google ADK Agent Example (Qualified file:// & github:// Roots)")
    print("=" * 88 + "\n")

    # 1. Read Dotenv Configuration containing qualified SKILLS_ROOTS
    dotenv_file = Path(__file__).parent / ".env"
    print(f"1. Reading environment configuration from '{dotenv_file.name}'...")
    env_config = parse_dotenv_file(dotenv_file)
    for k, v in env_config.items():
        masked_val = "******" if "TOKEN" in k or "SECRET" in k else v
        print(f"   [env] {k} = {masked_val}")
    print()

    # 2. Demonstrate URI scheme parsing
    print("2. Parsing qualified root URIs from SKILLS_ROOTS:")
    roots_str = env_config.get("SKILLS_ROOTS", "")
    for root_uri in [r.strip() for r in roots_str.split(",") if r.strip()]:
        scheme, target, ref, subpath = parse_skill_root_uri(root_uri)
        print(f"   -> URI: '{root_uri}'")
        print(f"      Scheme:  {scheme}")
        print(f"      Target:  {target}")
        print(f"      Ref:     {ref}")
        print(f"      Subpath: {subpath}")
    print()

    # 3. Instantiate single unified SkillRegistry parsing file:// and github:// URIs
    print("3. Instantiating single SkillRegistry loading qualified file:// and github:// roots...")
    registry = SkillRegistry(dotenv_path=dotenv_file)
    print(f"   -> Successfully loaded {len(registry.skills)} skill(s) across all qualified roots:")
    for name, skill in registry.skills.items():
        print(f"      - [{name}] ({skill.path})")
    print()

    # 4. Instantiate ADK Native Toolset and Agent
    print("4. Instantiating ADK Toolset and ADK Native Agent...")
    toolset = ADKSkillToolset(registry=registry)
    agent = ADKNativeAgent(toolset=toolset)
    session_service = InMemorySessionService()
    session = await session_service.create_or_get_session("adk_example_session", user_token="google_oauth2_token_xyz")

    print(f"   -> Total Combined Qualified Skills Active: {len(toolset.skills)}")
    print("-" * 88)

    # 5. Execute prompt against ADK Agent
    prompt = "How do I build resilient BigQuery CAPI services with Gemini API skills?"
    print(f"\nPrompt: '{prompt}'\n")

    print("Agent Response Stream:")
    print("-" * 88)
    async for chunk in agent.execute_async(prompt, session):
        sys.stdout.write(chunk)
        sys.stdout.flush()
    print("\n" + "-" * 88 + "\n")


def run() -> None:
    """Synchronous CLI entrypoint for example-adk."""
    os.chdir(Path(__file__).parent)
    asyncio.run(main())


def web() -> None:
    """Runs the native ADK example web server control plane."""
    os.chdir(Path(__file__).parent)
    from skills_agent.server import start_server
    start_server(host="0.0.0.0", port=8000)


if __name__ == "__main__":
    run()
