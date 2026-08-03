"""Unit tests for retailcortex-skills-mcp package."""

import uuid
import pytest
from data import get_engine, init_db, reset_engine
from skills_mcp import (
    delete_skill,
    get_skill,
    mcp_server,
    register_app,
    register_skill,
    search_skills,
    update_skill,
    verify_app,
)


def test_mcp_tools_lifecycle() -> None:
    """Verifies piped MCP tools execute cleanly in memory without network calls."""
    reset_engine()
    engine = get_engine("sqlite:///:memory:")
    init_db(engine)

    assert mcp_server.name == "Skills Service MCP"

    email = f"mcp-{uuid.uuid4()}@example.com"
    app_res = register_app("MCPApp", email)
    assert app_res["app_name"] == "MCPApp"

    verify_res = verify_app(app_res["verification_token"])
    assert verify_res["is_active"] is True

    skill_res = register_skill(
        api_key=app_res["api_key"],
        name="mcp-skill",
        description="piped mcp desc",
        instructions="# mcp inst",
        category="test",
    )
    assert skill_res["name"] == "mcp-skill"

    fetched = get_skill("mcp-skill")
    assert fetched["id"] == skill_res["id"]

    searched = search_skills("mcp")
    assert len(searched) == 1

    upd_res = update_skill(
        api_key=app_res["api_key"],
        skill_id=skill_res["id"],
        description="updated mcp desc",
    )
    assert upd_res["description"] == "updated mcp desc"

    del_res = delete_skill(api_key=app_res["api_key"], skill_id=skill_res["id"])
    assert del_res["status"] == "success"
