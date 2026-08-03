"""FastMCP server providing piped in-memory MCP tools for skill management and discovery."""

import logging
from typing import Any, Dict, List, Optional
from fastmcp import FastMCP
from sqlmodel import Session

from data.database import engine
from services.apps_service import AppsService
from services.skills_service import SkillsService
from model.app import AppRegisterRequest
from model.skill import SkillCreateRequest, SkillUpdateRequest

logger = logging.getLogger("skills_mcp.server")

mcp_server = FastMCP(
    "Skills Service MCP",
    instructions="Enterprise MCP server for searching, registering, and retrieving AI Agent skills.",
)


@mcp_server.tool()
def search_skills(query: str = "") -> List[Dict[str, Any]]:
    """Searches registered skills using Gemini semantic vector matching."""
    with Session(engine) as session:
        results = SkillsService.list_skills(session, query=query)
        return [r.model_dump(mode="json") for r in results]


@mcp_server.tool()
def get_skill(skill_id_or_name: str) -> Dict[str, Any]:
    """Retrieves full details and compiled schema of a skill by ID or name."""
    with Session(engine) as session:
        skill = SkillsService.get_skill(session, skill_id_or_name)
        return skill.model_dump(mode="json")


@mcp_server.tool()
def register_app(app_name: str, email: str) -> Dict[str, Any]:
    """Registers a new application and returns API key and verification link."""
    with Session(engine) as session:
        req = AppRegisterRequest(app_name=app_name, email=email)
        res = AppsService.register_app(session, req)
        return res.model_dump(mode="json")


@mcp_server.tool()
def verify_app(token: str) -> Dict[str, Any]:
    """Verifies and activates a registered application."""
    with Session(engine) as session:
        res = AppsService.verify_app(session, token)
        return res.model_dump(mode="json")


@mcp_server.tool()
def register_skill(
    api_key: str,
    name: str,
    description: str,
    instructions: str,
    category: str = "general",
) -> Dict[str, Any]:
    """Registers a new skill using an active application API key."""
    with Session(engine) as session:
        app = AppsService.authenticate_api_key(session, api_key)
        req = SkillCreateRequest(
            name=name,
            description=description,
            instructions=instructions,
            category=category,
        )
        res = SkillsService.create_skill(session, app.app_id, req)
        return res.model_dump(mode="json")


@mcp_server.tool()
def update_skill(
    api_key: str,
    skill_id: str,
    description: Optional[str] = None,
    instructions: Optional[str] = None,
) -> Dict[str, Any]:
    """Updates an existing skill using an active application API key."""
    with Session(engine) as session:
        app = AppsService.authenticate_api_key(session, api_key)
        req = SkillUpdateRequest(description=description, instructions=instructions)
        res = SkillsService.update_skill(session, skill_id, app.app_id, req)
        return res.model_dump(mode="json")


@mcp_server.tool()
def delete_skill(api_key: str, skill_id: str) -> Dict[str, Any]:
    """Deletes an existing skill using an active application API key."""
    with Session(engine) as session:
        app = AppsService.authenticate_api_key(session, api_key)
        res = SkillsService.delete_skill(session, skill_id, app.app_id)
        return res
