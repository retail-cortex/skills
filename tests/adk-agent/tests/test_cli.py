"""Unit tests for interactive CLI interface."""

import asyncio
import io
import sys
import unittest
from unittest.mock import patch

from skills_agent.cli import InteractiveAgentCLI, main


class TestCLI(unittest.TestCase):
    """Tests evaluating CLI commands and REPL handlers."""

    def setUp(self) -> None:
        self.cli = InteractiveAgentCLI(port=8000)

    def test_cli_initialization(self) -> None:
        self.assertIsNotNone(self.cli.registry)
        self.assertIsNotNone(self.cli.agent)
        self.assertGreaterEqual(len(self.cli.registry.skills), 23)

    def test_handle_list_skills_output(self) -> None:
        captured = io.StringIO()
        with patch("sys.stdout", captured):
            self.cli.handle_list_skills()
        output = captured.getvalue()
        self.assertIn("Registered Enterprise Skills", output)
        self.assertIn("python-core", output)

    def test_handle_skill_info_output(self) -> None:
        captured = io.StringIO()
        with patch("sys.stdout", captured):
            self.cli.handle_skill_info("python-core")
        output = captured.getvalue()
        self.assertIn("Skill: python-core", output)

    def test_handle_search_output(self) -> None:
        captured = io.StringIO()
        with patch("sys.stdout", captured):
            self.cli.handle_search("go-lang")
        output = captured.getvalue()
        self.assertIn("Search Results for 'go-lang'", output)

    def test_execute_prompt_async(self) -> None:
        captured = io.StringIO()
        with patch("sys.stdout", captured):
            asyncio.run(self.cli.execute_prompt_async("Help me configure Go microservice"))
        output = captured.getvalue()
        self.assertIn("Agent Response:", output)

    def test_main_cli_arguments(self) -> None:
        with patch.object(sys, "argv", ["skills-agent", "--list-skills"]):
            captured = io.StringIO()
            with patch("sys.stdout", captured):
                main()
            self.assertIn("Registered Enterprise Skills", captured.getvalue())


if __name__ == "__main__":
    unittest.main()
