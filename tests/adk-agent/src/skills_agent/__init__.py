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

"""Skills Agent Package for ADK Programming Agent with FastAPI control plane."""

from skills_agent.agent import ADKProgrammingAgent, InMemorySessionService, InvocationContext, Session, ToolContext
from skills_agent.cli import InteractiveAgentCLI, main
from skills_agent.server import app, create_app
from skills_agent.skills_loader import SkillRegistry, find_registry_root, load_all_skills
from skills_agent.types import AgentPromptRequest, AgentPromptResponse, SessionState, SkillDefinition, SkillSummary

__all__ = [
    "ADKProgrammingAgent",
    "InMemorySessionService",
    "InvocationContext",
    "Session",
    "ToolContext",
    "InteractiveAgentCLI",
    "main",
    "app",
    "create_app",
    "SkillRegistry",
    "find_registry_root",
    "load_all_skills",
    "AgentPromptRequest",
    "AgentPromptResponse",
    "SessionState",
    "SkillDefinition",
    "SkillSummary",
]
