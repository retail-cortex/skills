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

from typing import List, Optional
from uuid import UUID
import httpx
from fastapi import HTTPException
from fastmcp import FastMCP
from pydantic import BaseModel, Field

# Domain Model
class CustomerPayload(BaseModel):
    name: str = Field(description="Full legal name of customer")
    email: str = Field(description="Primary contact email")
    active: bool = Field(default=True)

# MCP Server Initialization
mcp = FastMCP(
    name="Customer Management MCP Server",
    instructions="Provides tools to list, query, and mutate customer records in the enterprise registry.",
)

API_BASE = "http://localhost:8000/api/v1"

@mcp.tool()
async def list_all_customers() -> List[dict]:
    """Retrieves all registered customers from the database."""
    async with httpx.AsyncClient(base_url=API_BASE) as client:
        resp = await client.get("/customers")
        resp.raise_for_status()
        return resp.json()

@mcp.tool()
async def get_customer_by_id(customer_id: UUID) -> dict:
    """Retrieves a single customer by their unique UUID identifier."""
    async with httpx.AsyncClient(base_url=API_BASE) as client:
        resp = await client.get(f"/customers/{customer_id}")
        if resp.status_code == 404:
            raise HTTPException(status_code=404, detail="Customer not found")
        resp.raise_for_status()
        return resp.json()

@mcp.tool()
async def create_new_customer(customer: CustomerPayload) -> dict:
    """Creates a new customer record."""
    async with httpx.AsyncClient(base_url=API_BASE) as client:
        resp = await client.post("/customers", json=customer.model_dump())
        resp.raise_for_status()
        return resp.json()

if __name__ == "__main__":
    mcp.run(transport="http", port=8001)
