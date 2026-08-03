"""Unit tests for retailcortex-skills-model package."""

from datetime import datetime, timezone
from model import (
    RegisteredApp,
    AppRegisterRequest,
    AppRegisterResponse,
    AppVerifyResponse,
    Skill,
    SkillVersion,
    SkillMetadata,
    SkillResource,
    SkillExample,
    SkillEmbedding,
    SkillCreateRequest,
    SkillUpdateRequest,
    SkillResponse,
)


def test_registered_app_model() -> None:
    """Verifies RegisteredApp creation and defaults."""
    app = RegisteredApp(
        app_name="TestApp",
        email="test@example.com",
        api_key_hash="hash123",
    )
    assert app.app_name == "TestApp"
    assert app.is_active is False
    assert app.app_id is not None
    assert app.verification_token is not None


def test_app_dto_models() -> None:
    """Verifies AppRegisterRequest, AppRegisterResponse, AppVerifyResponse."""
    req = AppRegisterRequest(app_name="App", email="a@example.com")
    assert req.app_name == "App"

    res = AppRegisterResponse(
        app_id="1",
        app_name="App",
        email="a@example.com",
        api_key="key",
        verification_token="token",
        verification_url="http://url",
    )
    assert res.app_id == "1"

    verify_res = AppVerifyResponse(
        app_id="1",
        app_name="App",
        email="a@example.com",
        is_active=True,
        message="Verified",
    )
    assert verify_res.is_active is True


def test_skill_models() -> None:
    """Verifies Skill and sub-entity models."""
    skill = Skill(
        app_id="app-1",
        name="test-skill",
        description="description",
        instructions="# test",
        sha256_hash="sha123",
    )
    assert skill.name == "test-skill"
    assert skill.hitl_tier == "TIER_1_AUTO_READ"

    version = SkillVersion(
        skill_id=skill.id,
        version="1.0.0",
        json_schema_json="{}",
        sha256_hash="sha123",
    )
    assert version.version == "1.0.0"

    meta = SkillMetadata(skill_id=skill.id, key="domain", value="test")
    assert meta.key == "domain"

    res = SkillResource(skill_id=skill.id, name="res1", content="content")
    assert res.name == "res1"

    ex = SkillExample(skill_id=skill.id, name="ex1", content="example")
    assert ex.name == "ex1"

    emb = SkillEmbedding(skill_id=skill.id, embedding_json="[0.1, 0.2]")
    assert emb.model_name == "text-embedding-004"


def test_skill_dtos() -> None:
    """Verifies SkillCreateRequest, SkillUpdateRequest, and SkillResponse."""
    create_req = SkillCreateRequest(
        name="skill1",
        description="desc",
        instructions="# inst",
    )
    assert create_req.version == "1.0.0"

    update_req = SkillUpdateRequest(description="new desc")
    assert update_req.description == "new desc"

    response = SkillResponse(
        id="1",
        app_id="app-1",
        name="skill1",
        description="desc",
        instructions="# inst",
        license=None,
        author=None,
        category=None,
        tags=[],
        trigger_phrases=[],
        version="1.0.0",
        sha256_hash="hash",
        hitl_tier="TIER_1",
        json_schema={},
        metadata={},
        references={},
        examples={},
        created_at=datetime.now(timezone.utc),
        updated_at=datetime.now(timezone.utc),
    )
    assert response.name == "skill1"
