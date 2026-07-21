"""Interactive CLI interface for interacting with the ADK Programming Agent."""

import argparse
import asyncio
import os
import sys
import threading
from pathlib import Path
from typing import List, NoReturn, Optional

# Add package src to sys.path for direct script execution
current_file = Path(__file__).resolve()
pkg_root = current_file.parents[2]
if str(pkg_root) not in sys.path:
    sys.path.insert(0, str(pkg_root))

from skills_agent.agent import ADKProgrammingAgent, InMemorySessionService, InvocationContext, Session
from skills_agent.skills_loader import SkillRegistry, find_registry_root


class InteractiveAgentCLI:
    """Terminal REPL interface driving the ADK Programming Agent and FastAPI control plane."""

    def __init__(self, port: int = 8000) -> None:
        self.registry: SkillRegistry = SkillRegistry()
        self.agent: ADKProgrammingAgent = ADKProgrammingAgent(self.registry)
        self.session_service: InMemorySessionService = InMemorySessionService()
        self.session_id: str = "cli_session"
        self.port: int = port
        self.server_thread: Optional[threading.Thread] = None

    def print_banner(self) -> None:
        """Prints startup banner and loaded skills summary."""
        count = len(self.registry.skills)
        print("\n" + "=" * 88)
        print(" Enterprise ADK Programming Agent CLI (FastAPI Control Plane)")
        print("=" * 88)
        print(f"Registry Root: {self.registry.root}")
        print(f"Active Skills: {count} enterprise skills loaded")
        print(f"ADK Agent:     programming_agent (ready)")
        print(f"FastAPI Port:  {self.port}")
        print("-" * 88)
        print("Commands:")
        print("  /skills               - List all registered enterprise skills")
        print("  /info <skill_name>    - Display instructions and references for a skill")
        print("  /search <keyword>     - Search skills for a language, tool, or CWE rule")
        print("  /session              - Inspect active session state & token delegation")
        print("  /help                 - Display command help")
        print("  /quit (or exit)       - Exit the interactive CLI")
        print("-" * 88 + "\n")

    def handle_list_skills(self) -> None:
        """Displays formatted table of all loaded skills."""
        summaries = self.registry.list_skills()
        print(f"\nRegistered Enterprise Skills ({len(summaries)} total):")
        print("-" * 88)
        print(f"{'Skill Name':<24} | {'Refs':<4} | {'Exs':<4} | {'Description'}")
        print("-" * 88)
        for s in summaries:
            desc = (s.description[:48] + "...") if len(s.description) > 48 else s.description
            print(f"{s.name:<24} | {s.reference_count:<4} | {s.example_count:<4} | {desc}")
        print("-" * 88 + "\n")

    def handle_skill_info(self, skill_name: str) -> None:
        """Displays full details for a specific skill."""
        skill = self.registry.get(skill_name)
        if not skill:
            print(f"\nError: Skill '{skill_name}' not found in registry.\n")
            return

        print("\n" + "=" * 88)
        print(f" Skill: {skill.name}")
        print("=" * 88)
        print(f"Description: {skill.description}")
        print(f"Location:    {skill.path}")
        if skill.references:
            print(f"References:  {', '.join(skill.references.keys())}")
        if skill.examples:
            print(f"Examples:    {', '.join(skill.examples.keys())}")
        print("-" * 88)
        print("Instructions:")
        print(skill.instructions[:1500] + ("\n... [truncated]" if len(skill.instructions) > 1500 else ""))
        print("=" * 88 + "\n")

    def handle_search(self, query: str) -> None:
        """Searches and displays matching skills."""
        results = self.registry.search(query)
        print(f"\nSearch Results for '{query}' ({len(results)} matches):")
        print("-" * 88)
        for s in results:
            print(f"- [{s.name}]({s.path}): {s.description}")
        print("-" * 88 + "\n")

    async def execute_prompt_async(self, prompt: str) -> None:
        """Executes query against ADK agent and streams response to stdout."""
        session = await self.session_service.create_or_get_session(self.session_id)
        context = InvocationContext(
            agent=self.agent,
            session=session,
            session_service=self.session_service,
            invocation_id=f"inv_cli_{len(session.history)}",
            request=prompt,
        )

        print("\nAgent Response:\n" + "-" * 88)
        async for chunk in self.agent.run_async(context):
            sys.stdout.write(chunk)
            sys.stdout.flush()
        print("\n" + "-" * 88 + "\n")

    async def repl_loop_async(self) -> None:
        """Main interactive read-eval-print loop."""
        self.print_banner()

        while True:
            try:
                user_input = input("adk> ").strip()
            except (EOFError, KeyboardInterrupt):
                print("\nExiting ADK CLI.")
                break

            if not user_input:
                continue

            if user_input in ("/quit", "quit", "exit", "/exit"):
                print("Exiting ADK CLI session.")
                break
            elif user_input in ("/skills", "skills", "/list"):
                self.handle_list_skills()
            elif user_input.startswith(("/info", "info")):
                parts = user_input.split(maxsplit=1)
                if len(parts) > 1:
                    self.handle_skill_info(parts[1].strip())
                else:
                    print("Usage: /info <skill_name>")
            elif user_input.startswith(("/search", "search")):
                parts = user_input.split(maxsplit=1)
                if len(parts) > 1:
                    self.handle_search(parts[1].strip())
                else:
                    print("Usage: /search <keyword>")
            elif user_input in ("/session", "session"):
                session = await self.session_service.create_or_get_session(self.session_id)
                print(f"\nSession ID: {session.id}")
                print(f"History turns: {len(session.history)}")
                print(f"State: {session.state}\n")
            elif user_input in ("/help", "help", "/?"):
                self.print_banner()
            else:
                await self.execute_prompt_async(user_input)


def main() -> None:
    """CLI parser and main entrypoint."""
    parser = argparse.ArgumentParser(
        description="Interactive ADK Programming Agent CLI wrapped in FastAPI."
    )
    parser.add_argument(
        "-q", "--query", type=str, help="Single query to run against the ADK programming agent."
    )
    parser.add_argument(
        "--list-skills", action="store_true", help="List all registered skills and exit."
    )
    parser.add_argument(
        "--info", type=str, help="Display information for a specific skill and exit."
    )
    parser.add_argument(
        "--search", type=str, help="Search registered skills by keyword and exit."
    )
    parser.add_argument(
        "--port", type=int, default=8000, help="Port for the FastAPI control server."
    )

    args = parser.parse_args()
    cli = InteractiveAgentCLI(port=args.port)

    if args.list_skills:
        cli.handle_list_skills()
        return

    if args.info:
        cli.handle_skill_info(args.info)
        return

    if args.search:
        cli.handle_search(args.search)
        return

    if args.query:
        asyncio.run(cli.execute_prompt_async(args.query))
        return

    # Interactive REPL mode
    asyncio.run(cli.repl_loop_async())


if __name__ == "__main__":
    main()
