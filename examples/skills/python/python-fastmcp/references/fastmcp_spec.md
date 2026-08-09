# FastMCP Protocol & Tool Specification

## FastMCP Architecture

FastMCP simplifies the Model Context Protocol (MCP) by translating Python function signatures and type hints into valid JSON schema tool definitions for LLM agents.

## Error Handling & Status Codes

Raise `fastapi.HTTPException` or MCP-specific error types to signal structured failures back to the calling agent:

```python
from fastapi import HTTPException

@mcp.tool()
async def fetch_customer(customer_id: str) -> dict:
    """Fetch customer details from the database."""
    async with httpx.AsyncClient(base_url="http://localhost:8000") as client:
        resp = await client.get(f"/customers/{customer_id}")
        if resp.status_code == 404:
            raise HTTPException(status_code=404, detail=f"Customer {customer_id} not found.")
        resp.raise_for_status()
        return resp.json()
```

## Transport Modes

FastMCP supports multiple connection transports:
- `transport="stdio"`: Standard input/output for local desktop or CLI agents (e.g. Gemini CLI, Claude Desktop).
- `transport="http"` / `transport="sse"`: Server-Sent Events over HTTP for networked microservices.
