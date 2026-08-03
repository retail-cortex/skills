"""Unit tests for retailcortex-skills-data package."""

import pytest
from fastapi import HTTPException
from sqlmodel import Session, SQLModel, create_engine
from sqlmodel.pool import StaticPool

from data import (
    AppsRepository,
    SkillsRepository,
    get_engine,
    get_session,
    init_db,
    reset_engine,
)
from model.app import AppRegisterRequest
from model.skill import SkillCreateRequest, SkillUpdateRequest


@pytest.fixture
def test_session():
    """Provides an in-memory SQLite database session for unit tests."""
    engine = create_engine(
        "sqlite:///:memory:",
        connect_args={"check_same_thread": False},
        poolclass=StaticPool,
    )
    SQLModel.metadata.create_all(engine)
    with Session(engine) as session:
        yield session


def test_database_helpers() -> None:
    """Verifies get_engine, reset_engine, init_db, and get_session."""
    eng = get_engine("sqlite:///:memory:")
    assert eng is not None
    init_db(eng)
    sessions = list(get_session(eng))
    assert len(sessions) == 1
    reset_engine()


def test_apps_repository_lifecycle(test_session: Session) -> None:
    """Verifies AppsRepository register, verify, and authenticate flows."""
    req = AppRegisterRequest(app_name="TestApp", email="dev@example.com")
    reg_res = AppsRepository.register_app(test_session, req, base_url="http://test")
    assert reg_res.app_name == "TestApp"
    assert reg_res.api_key.startswith("sk_live_")

    with pytest.raises(HTTPException) as exc:
        AppsRepository.register_app(test_session, req)
    assert exc.value.status_code == 400

    with pytest.raises(HTTPException) as exc:
        AppsRepository.authenticate_api_key(test_session, reg_res.api_key)
    assert exc.value.status_code == 403

    verify_res = AppsRepository.verify_app(test_session, reg_res.verification_token)
    assert verify_res.is_active is True

    authed_app = AppsRepository.authenticate_api_key(test_session, reg_res.api_key)
    assert authed_app.app_name == "TestApp"

    with pytest.raises(HTTPException) as exc:
        AppsRepository.verify_app(test_session, "invalid-token")
    assert exc.value.status_code == 404


def test_skills_repository_crud(test_session: Session) -> None:
    """Verifies SkillsRepository create, get, list, update, search, and delete."""
    app_req = AppRegisterRequest(app_name="App", email="owner@example.com")
    app_res = AppsRepository.register_app(test_session, app_req)
    AppsRepository.verify_app(test_session, app_res.verification_token)

    skill_req = SkillCreateRequest(
        name="test-skill",
        description="A great test skill",
        instructions="# Instructions\nAlways be direct.",
        tags=["ai", "test"],
        trigger_phrases=["run test"],
    )
    created = SkillsRepository.create_skill(
        test_session,
        app_id=app_res.app_id,
        request=skill_req,
        embedding_vector=[0.1, 0.2, 0.3],
    )
    assert created.name == "test-skill"
    assert len(created.tags) == 2

    fetched = SkillsRepository.get_skill(test_session, "test-skill")
    assert fetched.id == created.id

    listed = SkillsRepository.list_skills(test_session)
    assert len(listed) == 1

    keyword = SkillsRepository.list_skills(test_session, query="great")
    assert len(keyword) == 1

    vec = SkillsRepository.list_skills(test_session, query="semantic", query_vector=[0.1, 0.2, 0.3])
    assert len(vec) == 1
    assert vec[0].similarity_score is not None

    update_req = SkillUpdateRequest(description="Updated description", version="1.1.0")
    updated = SkillsRepository.update_skill(
        test_session,
        skill_id=created.id,
        app_id=app_res.app_id,
        request=update_req,
    )
    assert updated.description == "Updated description"
    assert updated.version == "1.1.0"

    res = SkillsRepository.delete_skill(test_session, created.id, app_res.app_id)
    assert res["status"] == "success"

    with pytest.raises(HTTPException) as exc:
        SkillsRepository.get_skill(test_session, "nonexistent")
    assert exc.value.status_code == 404
