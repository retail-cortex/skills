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

"""Unit tests for ADK programming agent and toolset."""

import asyncio
import unittest

from skills_agent.agent import (
    ADKProgrammingAgent,
    InMemorySessionService,
    InvocationContext,
    Session,
    SkillToolset,
    ToolContext,
    retry_with_jitter,
)
from skills_agent.skills_loader import SkillRegistry


class TestADKAgent(unittest.TestCase):
    """Tests evaluating ADK programming agent behaviors and tools."""

    def setUp(self) -> None:
        self.registry = SkillRegistry()
        self.agent = ADKProgrammingAgent(self.registry)
        self.session_service = InMemorySessionService()

    def test_tool_context_token_delegation(self) -> None:
        state = {"user_token": "google_oauth_token_xyz", "user_email": "engineer@company.com"}
        ctx = ToolContext(state=state)
        self.assertEqual(ctx.state["user_token"], "google_oauth_token_xyz")
        self.assertEqual(ctx.state["user_email"], "engineer@company.com")

    def test_in_memory_session_service(self) -> None:
        async def run_test() -> None:
            session = await self.session_service.create_or_get_session("sess_123", "token_abc")
            self.assertEqual(session.id, "sess_123")
            self.assertEqual(session.state.get("user_token"), "token_abc")

            retrieved = await self.session_service.get_session("sess_123")
            self.assertIsNotNone(retrieved)
            if retrieved:
                self.assertEqual(retrieved.id, "sess_123")

        asyncio.run(run_test())

    def test_skill_toolset_operations(self) -> None:
        toolset = SkillToolset(self.registry)
        all_skills = toolset.list_skills()
        self.assertGreaterEqual(len(all_skills), 23)

        details = toolset.get_skill_details("python-core")
        self.assertEqual(details.get("name"), "python-core")

        search_res = toolset.search_skills("terraform")
        self.assertTrue(len(search_res) > 0)

        # Test suggest_skills with default max of 3
        suggested = toolset.suggest_skills("python pytest database testing")
        self.assertTrue(0 < len(suggested) <= 3)

        guidance = toolset.generate_guidance("go-lang microservices")
        self.assertIn("go-lang", guidance.lower())

    def test_agent_execution(self) -> None:
        async def run_test() -> None:
            session = await self.session_service.create_or_get_session("test_exec_sess")
            resp = await self.agent.execute_query(
                "How should I structure a Go microservice?", session, self.session_service
            )
            self.assertTrue(len(resp) > 0)
            self.assertIn("Execution Strategy", resp)
            self.assertEqual(len(session.history), 2)

        asyncio.run(run_test())

    def test_retry_with_jitter(self) -> None:
        call_count = 0

        @retry_with_jitter(max_retries=2, base_delay=0.01)
        async def failing_operation() -> str:
            nonlocal call_count
            call_count += 1
            if call_count < 2:
                raise ValueError("Simulated transient 429 rate limit error")
            return "recovered"

        async def run_test() -> None:
            res = await failing_operation()
            self.assertEqual(res, "recovered")
            self.assertEqual(call_count, 2)

        asyncio.run(run_test())


if __name__ == "__main__":
    unittest.main()
