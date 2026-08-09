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

"""Type definitions and data structures for skills-agent."""

from dataclasses import dataclass, field
from typing import Dict, List, Optional

from loader import SkillDefinition, SkillSummary


@dataclass
class ChatMessage:
    """Single turn in a conversation."""

    role: str
    content: str
    timestamp: Optional[str] = None


@dataclass
class SessionState:
    """Tracks state and IAM delegation tokens across agent invocations."""

    session_id: str
    user_id: str = "default_user"
    user_email: Optional[str] = None
    user_token: Optional[str] = None
    messages: List[ChatMessage] = field(default_factory=list)
    state: Dict[str, str] = field(default_factory=dict)


@dataclass
class AgentPromptRequest:
    """Incoming request to invoke the ADK programming agent."""

    session_id: str
    prompt: str
    stream: bool = False
    skill_filter: Optional[List[str]] = None


@dataclass
class AgentPromptResponse:
    """Response returned from agent invocation."""

    session_id: str
    response: str
    skills_used: List[str] = field(default_factory=list)
    status: str = "completed"
