"""Pytest suite for Skills Service REST API, FastMCP server, and async gRPC servicers."""

import json
import pytest
from fastapi.testclient import TestClient
import grpc
from sqlmodel import SQLModel, Session, create_engine, select
from sqlmodel.pool import StaticPool

from api.v1 import skill_pb2, skill_service_pb2
from skills_app.main import app
from data import database
from skills_mcp import (
    mcp_server,
    register_skill as mcp_register_skill,
    get_skill as mcp_get_skill,
    search_skills as mcp_search_skills,
)
import skills_mcp.server as mcp_server_module
from services.grpc_servicers import AppServiceServicer, SkillServiceServicer
from services import grpc as grpc_pkg
from model.skill import (
    Skill,
    SkillMetadata,
    SkillResource,
    SkillExample,
)
from skills_app.config import settings


class MockServicerContext:
    """Mock gRPC ServicerContext for testing async gRPC servicers without a network server."""

    def __init__(self) -> None:
        self.aborted_code: grpc.StatusCode | None = None
        self.aborted_details: str | None = None

    async def abort(self, code: grpc.StatusCode, details: str) -> None:
        """Simulate context.abort by storing code/details and raising grpc.RpcError."""
        self.aborted_code = code
        self.aborted_details = details
        raise grpc.RpcError(code, details)


@pytest.fixture
def test_engine():
    """Create an in-memory SQLite database engine for isolated testing."""
    engine = create_engine(
        "sqlite://",
        connect_args={"check_same_thread": False},
        poolclass=StaticPool,
    )
    SQLModel.metadata.create_all(engine)

    # Patch engine across modules so REST endpoints, FastMCP tools, and gRPC servicers share test DB
    original_db_engine = database.engine
    original_mcp_engine = mcp_server_module.engine
    original_grpc_engine = grpc_pkg.servicers.engine

    database.engine = engine
    mcp_server_module.engine = engine
    grpc_pkg.servicers.engine = engine
    try:
        yield engine
    finally:
        database.engine = original_db_engine
        mcp_server_module.engine = original_mcp_engine
        grpc_pkg.servicers.engine = original_grpc_engine


@pytest.fixture
def client(test_engine):
    """Return a FastAPI TestClient configured to use the test database engine."""
    def get_session_override():
        with Session(test_engine) as session:
            yield session

    app.dependency_overrides[database.get_session] = get_session_override
    with TestClient(app) as c:
        yield c
    app.dependency_overrides.clear()


@pytest.fixture
def verified_app(client):
    """Register and verify a test application, returning its credentials dictionary."""
    reg_payload = {"app_name": "Test App", "email": "tester@example.com"}
    reg_resp = client.post("/api/v1/apps/register", json=reg_payload)
    assert reg_resp.status_code == 201
    reg_data = reg_resp.json()

    verify_resp = client.get(
        f"/api/v1/apps/verify?token={reg_data['verification_token']}"
    )
    assert verify_resp.status_code == 200
    assert verify_resp.json()["is_active"] is True
    return reg_data


def test_health_check(client: TestClient) -> None:
    """Test health check endpoint."""
    response = client.get("/health")
    assert response.status_code == 200
    json_data = response.json()
    assert json_data["status"] == "ok"
    assert json_data["ports"]["rest"] == settings.port
    assert json_data["ports"]["grpc"] == settings.grpc_port


def test_app_registration_and_verification_flow(client: TestClient) -> None:
    """Test registering an application, verifying email token, and using API key."""
    reg_payload = {"app_name": "Test Suite App", "email": "dev@example.com"}
    reg_resp = client.post("/api/v1/apps/register", json=reg_payload)
    assert reg_resp.status_code == 201
    reg_data = reg_resp.json()

    assert reg_data["app_name"] == "Test Suite App"
    assert reg_data["email"] == "dev@example.com"
    api_key = reg_data["api_key"]
    token = reg_data["verification_token"]
    assert api_key.startswith("sk_live_")

    skill_payload = {
        "name": "python-test-skill",
        "description": "Skill for testing python execution",
        "instructions": "Execute python scripts safely.",
        "category": "python",
    }
    unverified_resp = client.post(
        "/api/v1/skills",
        json=skill_payload,
        headers={"X-API-Key": api_key},
    )
    assert unverified_resp.status_code == 403

    verify_resp = client.get(f"/api/v1/apps/verify?token={token}")
    assert verify_resp.status_code == 200
    assert verify_resp.json()["is_active"] is True

    created_resp = client.post(
        "/api/v1/skills",
        json=skill_payload,
        headers={"X-API-Key": api_key},
    )
    assert created_resp.status_code == 201
    created_data = created_resp.json()
    assert created_data["name"] == "python-test-skill"
    assert created_data["app_id"] == reg_data["app_id"]
    skill_id = created_data["id"]

    get_resp = client.get(f"/api/v1/skills/{skill_id}")
    assert get_resp.status_code == 200
    assert get_resp.json()["name"] == "python-test-skill"

    list_resp = client.get("/api/v1/skills?s=python")
    assert list_resp.status_code == 200
    items = list_resp.json()
    assert len(items) >= 1
    assert items[0]["name"] == "python-test-skill"

    patch_resp = client.patch(
        f"/api/v1/skills/{skill_id}",
        json={"description": "Updated python description"},
        headers={"X-API-Key": api_key},
    )
    assert patch_resp.status_code == 200
    assert patch_resp.json()["description"] == "Updated python description"

    del_resp = client.delete(f"/api/v1/skills/{skill_id}", headers={"X-API-Key": api_key})
    assert del_resp.status_code == 200

    get_after_del = client.get(f"/api/v1/skills/{skill_id}")
    assert get_after_del.status_code == 404


def test_skill_tags_trigger_phrases_and_child_table_update_cleanup(
    client: TestClient, test_engine, verified_app: dict
) -> None:
    """Verify tags/trigger_phrases persistence and ensure old child rows are deleted on update."""
    api_key = verified_app["api_key"]
    skill_payload = {
        "name": "nlp-tagging-skill",
        "description": "Skill with tags and trigger phrases",
        "instructions": "Process text tags safely.",
        "category": "nlp",
        "tags": ["nlp", "gemini", "python"],
        "trigger_phrases": ["analyze tags", "summarize text"],
        "metadata": {"key1": "val1", "key2": "val2"},
        "references": {"doc.md": "# Documentation"},
        "examples": {"example.py": "print('hello')"},
    }
    created_resp = client.post(
        "/api/v1/skills",
        json=skill_payload,
        headers={"X-API-Key": api_key},
    )
    assert created_resp.status_code == 201
    data = created_resp.json()
    skill_id = data["id"]

    assert data["tags"] == ["nlp", "gemini", "python"]
    assert data["trigger_phrases"] == ["analyze tags", "summarize text"]
    assert data["metadata"] == {"key1": "val1", "key2": "val2"}

    # Check database directly to ensure child records exist
    with Session(test_engine) as session:
        meta_count = len(session.exec(select(SkillMetadata).where(SkillMetadata.skill_id == skill_id)).all())
        res_count = len(session.exec(select(SkillResource).where(SkillResource.skill_id == skill_id)).all())
        ex_count = len(session.exec(select(SkillExample).where(SkillExample.skill_id == skill_id)).all())
        assert meta_count == 2
        assert res_count == 1
        assert ex_count == 1

    # Perform an update replacing metadata, references, examples, and tags
    update_payload = {
        "metadata": {"new_key": "new_val"},
        "references": {"new_doc.md": "# New Doc"},
        "examples": {"new_example.py": "print('updated')"},
        "tags": ["updated_nlp"],
    }
    update_resp = client.patch(
        f"/api/v1/skills/{skill_id}",
        json=update_payload,
        headers={"X-API-Key": api_key},
    )
    assert update_resp.status_code == 200
    updated_data = update_resp.json()
    assert updated_data["tags"] == ["updated_nlp"]
    assert updated_data["metadata"] == {"new_key": "new_val"}

    # Check DB again: ensure old rows were deleted and no duplicates remain
    with Session(test_engine) as session:
        meta_records = session.exec(select(SkillMetadata).where(SkillMetadata.skill_id == skill_id)).all()
        res_records = session.exec(select(SkillResource).where(SkillResource.skill_id == skill_id)).all()
        ex_records = session.exec(select(SkillExample).where(SkillExample.skill_id == skill_id)).all()
        assert len(meta_records) == 1
        assert meta_records[0].key == "new_key"
        assert len(res_records) == 1
        assert res_records[0].name == "new_doc.md"
        assert len(ex_records) == 1
        assert ex_records[0].name == "new_example.py"


def test_fastmcp_tools(client: TestClient, verified_app: dict) -> None:
    """Test FastMCP tool endpoints for skill registration, lookup, and search."""
    api_key = verified_app["api_key"]

    # Register skill using FastMCP tool
    res = mcp_register_skill(
        api_key=api_key,
        name="mcp-tool-skill",
        description="Registered via MCP",
        instructions="Execute MCP commands.",
        category="mcp",
    )
    assert res["name"] == "mcp-tool-skill"
    skill_id = res["id"]

    # Get skill using FastMCP tool
    fetched = mcp_get_skill(skill_id_or_name=skill_id)
    assert fetched["id"] == skill_id
    assert fetched["name"] == "mcp-tool-skill"

    # Search skills using FastMCP tool
    search_results = mcp_search_skills(query="mcp")
    assert len(search_results) >= 1
    assert any(s["name"] == "mcp-tool-skill" for s in search_results)


@pytest.mark.asyncio
async def test_grpc_servicers_happy_path(test_engine) -> None:
    """Test AppServiceServicer and SkillServiceServicer async gRPC methods."""
    app_servicer = AppServiceServicer()
    skill_servicer = SkillServiceServicer()
    context = MockServicerContext()

    # Register application via gRPC
    reg_req = skill_service_pb2.RegisterAppRequest(
        app_name="gRPC App",
        email="grpc@example.com",
    )
    reg_resp = await app_servicer.RegisterApp(reg_req, context)
    assert reg_resp.app_id
    assert reg_resp.api_key.startswith("sk_live_")
    assert reg_resp.verification_token

    # Verify application via gRPC
    verify_req = skill_service_pb2.VerifyAppRequest(token=reg_resp.verification_token)
    verify_resp = await app_servicer.VerifyApp(verify_req, context)
    assert verify_resp.is_active is True

    # Register skill via gRPC
    skill_req = skill_service_pb2.RegisterSkillRequest(
        api_key=reg_resp.api_key,
        name="grpc-test-skill",
        description="Skill registered via gRPC",
        instructions="Run gRPC methods.",
        category="grpc",
        tags=["proto", "async"],
        trigger_phrases=["run grpc"],
    )
    skill_proto = await skill_servicer.RegisterSkill(skill_req, context)
    assert skill_proto.name == "grpc-test-skill"
    assert list(skill_proto.tags) == ["proto", "async"]
    assert list(skill_proto.trigger_phrases) == ["run grpc"]
    skill_id = skill_proto.compiled_reference.skill_id
    assert skill_id

    # Get skill via gRPC
    get_req = skill_service_pb2.GetSkillRequest(skill_id_or_name=skill_id)
    fetched_proto = await skill_servicer.GetSkill(get_req, context)
    assert fetched_proto.name == "grpc-test-skill"
    assert list(fetched_proto.tags) == ["proto", "async"]

    # List skills via gRPC
    list_req = skill_service_pb2.ListSkillsRequest(s="grpc")
    list_resp = await skill_servicer.ListSkills(list_req, context)
    assert len(list_resp.skills) >= 1
    assert any(s.name == "grpc-test-skill" for s in list_resp.skills)

    # Update skill via gRPC
    update_req = skill_service_pb2.UpdateSkillRequest(
        api_key=reg_resp.api_key,
        skill_id=skill_id,
        description="Updated gRPC description",
    )
    updated_proto = await skill_servicer.UpdateSkill(update_req, context)
    assert updated_proto.description == "Updated gRPC description"

    # Delete skill via gRPC
    del_req = skill_service_pb2.DeleteSkillRequest(
        api_key=reg_resp.api_key,
        skill_id=skill_id,
    )
    del_resp = await skill_servicer.DeleteSkill(del_req, context)
    assert del_resp.status == "success"


@pytest.mark.parametrize(
    "method,endpoint,payload,headers,expected_status",
    [
        ("POST", "/api/v1/skills", {"name": "invalid", "description": "d", "instructions": "i"}, {}, 401),
        ("POST", "/api/v1/skills", {"name": "invalid", "description": "d", "instructions": "i"}, {"X-API-Key": "sk_live_fake"}, 401),
        ("GET", "/api/v1/skills/non-existent-id", None, {}, 404),
    ],
)
def test_rest_error_cases(
    client: TestClient,
    method: str,
    endpoint: str,
    payload: dict | None,
    headers: dict,
    expected_status: int,
) -> None:
    """Parameterized validation of REST API error codes."""
    if method == "POST":
        response = client.post(endpoint, json=payload, headers=headers)
    elif method == "GET":
        response = client.get(endpoint, headers=headers)
    else:
        raise ValueError(f"Unsupported method: {method}")
    assert response.status_code == expected_status


@pytest.mark.asyncio
@pytest.mark.parametrize(
    "api_key,skill_id,expected_code",
    [
        ("sk_live_invalid", "some-id", grpc.StatusCode.UNAUTHENTICATED),
    ],
)
async def test_grpc_error_handling(
    test_engine, api_key: str, skill_id: str, expected_code: grpc.StatusCode
) -> None:
    """Parameterized test verifying proper gRPC status code mapping for authentication failures."""
    servicer = SkillServiceServicer()
    context = MockServicerContext()

    req = skill_service_pb2.DeleteSkillRequest(api_key=api_key, skill_id=skill_id)
    with pytest.raises(grpc.RpcError):
        await servicer.DeleteSkill(req, context)

    assert context.aborted_code == expected_code


if __name__ == "__main__":
    import sys
    sys.exit(pytest.main([__file__] + sys.argv[1:]))
