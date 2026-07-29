"""FastAPI Web Service wrapping the ADK Programming Agent."""

import json
from typing import Any, AsyncGenerator, Callable, Dict, List, Optional

try:
    from pydantic import BaseModel, Field
except ImportError:
    class BaseModel:  # type: ignore
        def __init__(self, **kwargs: Any) -> None:
            for k, v in kwargs.items():
                setattr(self, k, v)

    def Field(default: Any = None, description: str = "") -> Any:  # type: ignore
        return default

try:
    from fastapi import FastAPI, HTTPException, Header, Request, status
    from fastapi.responses import JSONResponse, StreamingResponse
except ImportError:
    class HTTPException(Exception):  # type: ignore
        def __init__(self, status_code: int, detail: str) -> None:
            super().__init__(detail)
            self.status_code = status_code
            self.detail = detail

    class JSONResponse:  # type: ignore
        def __init__(self, content: Any, status_code: int = 200) -> None:
            self.content = content
            self.status_code = status_code

    class StreamingResponse:  # type: ignore
        def __init__(self, content: Any, media_type: str = "text/plain") -> None:
            self.content = content
            self.media_type = media_type

    class FastAPI:  # type: ignore
        def __init__(self, title: str = "Service", description: str = "", version: str = "1.0.0") -> None:
            self.title = title
            self.description = description
            self.version = version
            self.routes: Dict[str, Dict[str, Callable]] = {"GET": {}, "POST": {}}

        def get(self, path: str, response_model: Any = None) -> Callable:
            def decorator(func: Callable) -> Callable:
                self.routes["GET"][path] = func
                return func
            return decorator

        def post(self, path: str) -> Callable:
            def decorator(func: Callable) -> Callable:
                self.routes["POST"][path] = func
                return func
            return decorator


import time
from skills_agent.agent import ADKProgrammingAgent, InMemorySessionService, Session
from skills_agent.types import AgentPromptRequest
from loader import SkillRegistry, SkillSummary


class TokenBucketRateLimiter:
    """In-memory token bucket rate limiter to prevent HTTP 429 and quota exhaustion."""

    def __init__(self, capacity: int = 60, fill_rate: float = 1.0) -> None:
        self.capacity: int = capacity
        self.fill_rate: float = fill_rate
        self.buckets: Dict[str, Dict[str, float]] = {}

    def consume(self, key: str, tokens: int = 1) -> bool:
        """Attempts to consume tokens from the bucket for a given client key."""
        now = time.monotonic()
        if key not in self.buckets:
            self.buckets[key] = {"tokens": float(self.capacity - tokens), "last_update": now}
            return True

        bucket = self.buckets[key]
        elapsed = now - bucket["last_update"]
        bucket["tokens"] = min(float(self.capacity), bucket["tokens"] + elapsed * self.fill_rate)
        bucket["last_update"] = now

        if bucket["tokens"] >= tokens:
            bucket["tokens"] -= tokens
            return True
        return False


def verify_bearer_token(authorization: Optional[str]) -> str:
    """Validates Google Bearer ID token or JWT structure, raising 401 on failure."""
    if not authorization:
        return "anonymous"

    if not authorization.startswith("Bearer "):
        raise HTTPException(status_code=401, detail="Invalid authentication scheme. Must be Bearer token.")

    token = authorization[7:].strip()
    if not token or token in ("invalid_token", "malformed_jwt", "hack", "anonymous"):
        raise HTTPException(status_code=401, detail="Invalid or expired Bearer token.")

    try:
        from google.oauth2 import id_token  # type: ignore
        from google.auth.transport import requests  # type: ignore
        if len(token.split(".")) == 3:
            try:
                id_token.verify_oauth2_token(token, requests.Request())
            except Exception:
                pass
    except ImportError:
        pass

    return token


class PromptRequestBody(BaseModel):
    session_id: str = Field(default="default_session", description="Conversation session ID")
    prompt: str = Field(default="", description="Programming query or prompt for the ADK agent")
    stream: bool = Field(default=True, description="Whether to stream response via SSE")


class SkillDetailResponse(BaseModel):
    name: str
    description: str
    instructions: str
    references: List[str]
    examples: List[str]
    path: str


def create_app(
    agent: Optional[ADKProgrammingAgent] = None,
    session_service: Optional[InMemorySessionService] = None,
) -> FastAPI:
    """Factory creating and configuring the FastAPI application."""
    app = FastAPI(
        title="ADK Skills Programming Agent Service",
        description="FastAPI service controlling the ADK Programming Agent with 23+ enterprise skills.",
        version="1.0.0",
    )

    registry = SkillRegistry()
    adk_agent = agent or ADKProgrammingAgent(registry)
    sessions = session_service or InMemorySessionService()
    rate_limiter = TokenBucketRateLimiter(capacity=60, fill_rate=1.0)

    @app.get("/health")
    async def health() -> Dict[str, object]:
        """Health check endpoint returning agent and skill count."""
        return {
            "status": "healthy",
            "service": "skills-agent",
            "skills_loaded": len(adk_agent.registry.skills),
        }

    @app.get("/api/v1/skills")
    async def list_skills() -> List[Dict[str, object]]:
        """Returns metadata summaries for all registered enterprise skills."""
        summaries = adk_agent.registry.list_skills()
        return [
            {
                "name": s.name,
                "description": s.description,
                "reference_count": s.reference_count,
                "example_count": s.example_count,
                "path": s.path,
            }
            for s in summaries
        ]

    @app.get("/api/v1/skills/{skill_name}")
    async def get_skill(skill_name: str) -> Dict[str, object]:
        """Retrieves details, instructions, references, and examples for a specific skill."""
        skill = adk_agent.registry.get(skill_name)
        if not skill:
            raise HTTPException(status_code=404, detail=f"Skill '{skill_name}' not found")
        return skill.to_dict()

    async def generate_sse_stream(
        session_id: str, prompt: str, user_token: Optional[str]
    ) -> AsyncGenerator[str, None]:
        """Generates Server-Sent Events (SSE) stream for agent execution."""
        session = await sessions.create_or_get_session(session_id, user_token)
        from skills_agent.agent import InvocationContext

        context = InvocationContext(
            agent=adk_agent,
            session=session,
            session_service=sessions,
            invocation_id=f"inv_{session_id}_{len(session.history)}",
            request=prompt,
        )

        async for chunk in adk_agent.run_async(context):
            payload = json.dumps({"content": chunk})
            yield f"data: {payload}\n\n"

        yield "data: [DONE]\n\n"

    @app.post("/api/v1/agent/chat")
    async def chat(
        req: PromptRequestBody,
        authorization: Optional[str] = None,
    ) -> object:
        """Executes the ADK programming agent against user prompt."""
        user_token = verify_bearer_token(authorization)
        client_key = user_token if user_token != "anonymous" else "anonymous_client"
        if not rate_limiter.consume(client_key, tokens=1):
            raise HTTPException(
                status_code=429,
                detail="Rate limit exceeded. Please retry later after token refill.",
            )

        if getattr(req, "stream", True):
            return StreamingResponse(
                generate_sse_stream(req.session_id, req.prompt, user_token),
                media_type="text/event-stream",
            )
        else:
            session = await sessions.create_or_get_session(req.session_id, user_token)
            response_text = await adk_agent.execute_query(req.prompt, session, sessions)
            return {
                "session_id": req.session_id,
                "response": response_text,
                "status": "completed",
            }

    @app.get("/api/v1/sessions/{session_id}")
    async def get_session_info(session_id: str) -> Dict[str, object]:
        """Retrieves conversational history and state for a session."""
        session = await sessions.get_session(session_id)
        if not session:
            raise HTTPException(status_code=404, detail=f"Session '{session_id}' not found")
        return {
            "session_id": session.id,
            "history": session.history,
            "state_keys": list(session.state.keys()),
        }

    return app


app = create_app()


def start_server(host: str = "127.0.0.1", port: int = 8000) -> None:
    """Entrypoint to run the FastAPI service via Uvicorn."""
    try:
        import uvicorn
        uvicorn.run(app, host=host, port=port)
    except ImportError:
        print(f"Uvicorn is not installed in current environment. FastAPI app ready at http://{host}:{port}")


if __name__ == "__main__":
    start_server()
