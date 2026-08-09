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

import asyncio
import json
from typing import AsyncGenerator
from fastapi import FastAPI, Request, HTTPException
from fastapi.responses import StreamingResponse
from pydantic import BaseModel, Field
from google.adk.agents import Agent, InvocationContext
from google.adk.sessions import InMemorySessionService, Session
from google.adk.tools import ToolContext

# Request schema
class PromptRequest(BaseModel):
    session_id: str = Field(description="Unique session conversation ID")
    prompt: str = Field(description="User prompt text")

# Domain Tool accessing ToolContext state
async def get_account_summary(account_id: str, tool_context: ToolContext = None) -> dict:
    """Fetches high-level account summary and metrics."""
    # Read user auth or state injected into tool_context
    token = tool_context.state.get("user_token") if tool_context else "anonymous"
    return {
        "account_id": account_id,
        "status": "ACTIVE",
        "authorized_by": token,
        "balance": 450000.0,
    }

# ADK Agent Definition
adk_agent = Agent(
    name="financial_advising_agent",
    global_instruction="You are a financial advisor agent. Query account summaries to answer questions.",
    tools=[get_account_summary],
)

app = FastAPI(title="ADK Agent FastAPI Service")
session_service = InMemorySessionService()

async def generate_agent_stream(session_id: str, prompt: str, user_token: str) -> AsyncGenerator[str, None]:
    session = await session_service.get_session(session_id)
    if not session:
        session = Session(id=session_id, appName="enterprise_app", userId="user_123")
    
    # Store request credentials in session state
    session.state["user_token"] = user_token

    context = InvocationContext(
        agent=adk_agent,
        session=session,
        session_service=session_service,
        invocation_id=f"inv_{session_id}",
        request=prompt,
    )

    async for chunk in adk_agent.run_async(context):
        data = json.dumps({"content": str(chunk)})
        yield f"data: {data}\n\n"
    
    yield "data: [DONE]\n\n"

@app.post("/api/v1/agent/chat")
async def chat(req: PromptRequest, request: Request):
    user_token = request.headers.get("Authorization", "Bearer dev-token")
    return StreamingResponse(
        generate_agent_stream(req.session_id, req.prompt, user_token),
        media_type="text/event-stream",
    )
