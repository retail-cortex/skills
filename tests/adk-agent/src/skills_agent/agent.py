# Copyright 2026 Ryan McGuinness
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""ADK Programming Agent implementation with skill toolset and 429 rate limit resilience."""

import asyncio
import json
import random
from typing import AsyncGenerator, Callable, Dict, List, Optional

from loader import SkillDefinition, SkillRegistry


class ToolContext:
    """ADK ToolContext holding invocation state and Google OAuth2 credentials."""

    def __init__(self, state: Optional[Dict[str, str]] = None) -> None:
        self.state: Dict[str, str] = state if state is not None else {}


class Session:
    """ADK Session tracking conversational memory and IAM token delegation."""

    def __init__(self, id: str, app_name: str = "skills-agent", user_id: str = "default_user") -> None:
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
        """Retrieves or creates a session, updating credentials in session state."""
        if session_id not in self._sessions:
            self._sessions[session_id] = Session(id=session_id)
        session = self._sessions[session_id]
        if user_token:
            session.state["user_token"] = user_token
        return session


class InvocationContext:
    """Encapsulates context for an ADK agent execution request."""

    def __init__(
        self,
        agent: "ADKProgrammingAgent",
        session: Session,
        session_service: InMemorySessionService,
        invocation_id: str,
        request: str,
    ) -> None:
        self.agent: "ADKProgrammingAgent" = agent
        self.session: Session = session
        self.session_service: InMemorySessionService = session_service
        self.invocation_id: str = invocation_id
        self.request: str = request


def retry_with_jitter(max_retries: int = 3, base_delay: float = 0.1) -> Callable:
    """Decorator applying exponential backoff with randomized jitter for 429 quota resilience."""

    def decorator(func: Callable) -> Callable:
        async def wrapper(*args: object, **kwargs: object) -> object:
            retries = 0
            while True:
                try:
                    return await func(*args, **kwargs)
                except Exception as e:
                    retries += 1
                    if retries > max_retries:
                        raise e
                    sleep_time = base_delay * (2 ** (retries - 1)) + random.uniform(0, 0.05)
                    await asyncio.sleep(sleep_time)

        return wrapper

    return decorator


class SkillToolset:
    """Domain toolset exposing the 23+ repository skills to the ADK agent."""

    def __init__(self, registry: SkillRegistry) -> None:
        self.registry: SkillRegistry = registry

    def list_skills(self, tool_context: Optional[ToolContext] = None) -> List[Dict[str, object]]:
        """Lists all registered enterprise skills."""
        return [s.to_dict() for s in self.registry.skills.values()]

    def get_skill_details(self, skill_name: str, tool_context: Optional[ToolContext] = None) -> Dict[str, object]:
        """Retrieves detailed instructions, references, and examples for a skill."""
        skill = self.registry.get(skill_name)
        if not skill:
            return {"error": f"Skill '{skill_name}' not found."}
        return skill.to_dict()

    def search_skills(self, query: str, tool_context: Optional[ToolContext] = None) -> List[Dict[str, object]]:
        """Searches skills by keyword or domain pattern."""
        matches = self.registry.search(query)
        return [s.to_dict() for s in matches]

    def suggest_skills(self, query: str, max_skills: int = 3, tool_context: Optional[ToolContext] = None) -> List[Dict[str, object]]:
        """Suggests the top-k most relevant skills for an agent prompt based on vector search ranking."""
        matches = self.registry.suggest_skills(query, max_skills=max_skills)
        return [s.to_dict() for s in matches]

    def generate_guidance(self, query: str, tool_context: Optional[ToolContext] = None) -> str:
        """Generates architectural guidance by synthesizing matching enterprise skills."""
        matches = self.registry.suggest_skills(query, max_skills=3)
        if not matches:
            return f"No direct skill match found for '{query}'. Using standard enterprise SDLC guidelines."

        lines: List[str] = [f"Found {len(matches)} matching enterprise skills for '{query}':\n"]
        for s in matches:
            lines.append(f"### Skill: {s.name}")
            lines.append(f"**Description**: {s.description}")
            if s.instructions:
                # Extract first 500 chars of instructions
                snippet = s.instructions[:400].replace("\n", " ")
                lines.append(f"**Guidelines**: {snippet}...\n")
        return "\n".join(lines)


class ADKProgrammingAgent:
    """Primary enterprise programming agent in ADK utilizing all repository skills."""

    def __init__(self, registry: Optional[SkillRegistry] = None) -> None:
        self.registry: SkillRegistry = registry or SkillRegistry()
        self.toolset: SkillToolset = SkillToolset(self.registry)
        self.name: str = "programming_agent"
        self.global_instruction: str = (
            "You are the primary enterprise programming agent. "
            "You have access to all skills in the repository registry. "
            "Enforce strict Python/Go/Java/Bazel/Terraform SDLC standards, 90% TDD coverage, "
            "HTTP 429 rate limit backoff resilience, and CWE security invariants."
        )

    @retry_with_jitter(max_retries=3, base_delay=0.05)
    async def _simulate_llm_reasoning(self, prompt: str, relevant_skills: List[SkillDefinition]) -> str:
        """Simulates AI model inference synthesized with relevant skill instructions."""
        await asyncio.sleep(0.01)  # Minimal async yield
        response_parts: List[str] = []

        if relevant_skills:
            skill_names = ", ".join([f"`{s.name}`" for s in relevant_skills])
            response_parts.append(f"Consulted Enterprise Skills: {skill_names}.\n")

            for s in relevant_skills[:3]:
                response_parts.append(f"#### [{s.name}]({s.path})\n{s.description}\n")
                # Add key actionable rule from instructions
                first_lines = [line for line in s.instructions.splitlines() if line.startswith(("-", "*", "1.", "2.", "3.", "##"))][:4]
                if first_lines:
                    response_parts.append("**Core Invariants:**\n" + "\n".join(first_lines) + "\n")
        else:
            response_parts.append("General Enterprise Programming Guidance:\n")
            response_parts.append("Enforce strict typing, paired positive/negative TDD, 429 backoff resilience, and CWE security checks.\n")

        response_parts.append(f"\n**Execution Strategy for Request:**\n> {prompt}\n")
        response_parts.append("1. Validate architecture against the registered L3 skill references.\n")
        response_parts.append("2. Implement paired TDD test suites before writing production logic.\n")
        response_parts.append("3. Verify build and static analysis via Bazel (`bazel test //...`).\n")

        return "\n".join(response_parts)

    async def run_async(self, context: InvocationContext) -> AsyncGenerator[str, None]:
        """Runs the ADK agent execution loop asynchronously, streaming response chunks."""
        prompt = context.request
        # Check OAuth token delegation in ToolContext / Session
        tool_ctx = ToolContext(state=context.session.state)
        token = tool_ctx.state.get("user_token", "unauthenticated")

        # Pre-call optimization: dynamically pull top relevant skills based on search optimization (max 3)
        relevant_skills = self.registry.suggest_skills(prompt, max_skills=3)

        # Generate response
        content = await self._simulate_llm_reasoning(prompt, relevant_skills)

        # Record in session history
        context.session.history.append({"role": "user", "content": prompt})
        context.session.history.append({"role": "assistant", "content": content})

        # Stream in chunks
        chunk_size = 64
        for i in range(0, len(content), chunk_size):
            chunk = content[i : i + chunk_size]
            yield chunk
            await asyncio.sleep(0.005)

    async def execute_query(
        self, prompt: str, session: Session, session_service: InMemorySessionService
    ) -> str:
        """Helper to synchronously execute a single query and return full response text."""
        context = InvocationContext(
            agent=self,
            session=session,
            session_service=session_service,
            invocation_id=f"inv_{session.id}_{len(session.history)}",
            request=prompt,
        )
        chunks: List[str] = []
        async for chunk in self.run_async(context):
            chunks.append(chunk)
        return "".join(chunks)
