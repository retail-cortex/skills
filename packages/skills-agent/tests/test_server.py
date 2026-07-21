"""Unit tests for FastAPI control plane server."""

import asyncio
import unittest

from skills_agent.agent import ADKProgrammingAgent, InMemorySessionService
from skills_agent.server import PromptRequestBody, create_app
from skills_agent.skills_loader import SkillRegistry


class TestServer(unittest.TestCase):
    """Tests evaluating the FastAPI web service wrapping the ADK agent."""

    def setUp(self) -> None:
        self.registry = SkillRegistry()
        self.agent = ADKProgrammingAgent(self.registry)
        self.session_service = InMemorySessionService()
        self.app = create_app(agent=self.agent, session_service=self.session_service)

    def test_app_initialization(self) -> None:
        self.assertIsNotNone(self.app)
        self.assertEqual(self.app.title, "ADK Skills Programming Agent Service")

    def test_server_routes_direct(self) -> None:
        # Test direct FastAPI route endpoints using TestClient or direct invocation
        try:
            from fastapi.testclient import TestClient

            client = TestClient(self.app)
            # Health endpoint
            resp = client.get("/health")
            self.assertEqual(resp.status_code, 200)
            data = resp.json()
            self.assertEqual(data["status"], "healthy")
            self.assertGreaterEqual(data["skills_loaded"], 23)

            # Skills list endpoint
            resp_skills = client.get("/api/v1/skills")
            self.assertEqual(resp_skills.status_code, 200)
            skills_data = resp_skills.json()
            self.assertGreaterEqual(len(skills_data), 23)

            # Single skill details endpoint
            resp_single = client.get("/api/v1/skills/python-core")
            self.assertEqual(resp_single.status_code, 200)
            self.assertEqual(resp_single.json()["name"], "python-core")

            # 404 skill
            resp_404 = client.get("/api/v1/skills/unknown-skill-xyz")
            self.assertEqual(resp_404.status_code, 404)

            # Agent chat endpoint (non-streaming)
            req_body = {"session_id": "test_http_sess", "prompt": "Python 3.13 guidelines", "stream": False}
            resp_chat = client.post("/api/v1/agent/chat", json=req_body, headers={"Authorization": "Bearer test_token"})
            self.assertEqual(resp_chat.status_code, 200)
            self.assertEqual(resp_chat.json()["status"], "completed")

        except ImportError:
            # Fallback assertion if TestClient (httpx) is not in minimal test runner
            self.assertTrue(len(self.agent.registry.skills) >= 23)


if __name__ == "__main__":
    unittest.main()
