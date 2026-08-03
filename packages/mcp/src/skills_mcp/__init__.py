"""Retail Cortex Skills MCP library exporting piped in-memory FastMCP server and tools."""

from skills_mcp.server import (
    mcp_server,
    search_skills,
    get_skill,
    register_app,
    verify_app,
    register_skill,
    update_skill,
    delete_skill,
)

__all__ = [
    "mcp_server",
    "search_skills",
    "get_skill",
    "register_app",
    "verify_app",
    "register_skill",
    "update_skill",
    "delete_skill",
]
