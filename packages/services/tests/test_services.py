"""Unit tests for retailcortex-skills-service package."""

import pytest
from sqlmodel import Session, SQLModel, create_engine
from sqlmodel.pool import StaticPool

from data import init_db
from model.app import AppRegisterRequest
from model.skill import SkillCreateRequest, SkillUpdateRequest
from services import (
    AppsService,
    EmbeddingService,
    SkillsService,
    get_tracer,
    setup_telemetry,
)


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


def test_embedding_service_methods() -> None:
    """Verifies EmbeddingService cosine similarity and initialization fallbacks."""
    svc = EmbeddingService(model_name="test-model", gemini_api_key="fake-key")
    assert svc.model_name == "test-model"

    sim = EmbeddingService.cosine_similarity([1.0, 0.0], [1.0, 0.0])
    assert sim == 1.0

    sim_zero = EmbeddingService.cosine_similarity([0.0, 0.0], [1.0, 1.0])
    assert sim_zero == 0.0

    emb = svc.generate_embedding("")
    assert emb is None


def test_telemetry_setup() -> None:
    """Verifies telemetry setup and tracer acquisition."""
    setup_telemetry(enable_telemetry=False)
    tracer = get_tracer("test")
    assert tracer is not None


def test_apps_and_skills_services(test_session: Session) -> None:
    """Verifies orchestration of AppsService and SkillsService."""
    req = AppRegisterRequest(app_name="AppSvc", email="svc@example.com")
    app_res = AppsService.register_app(test_session, req)
    AppsService.verify_app(test_session, app_res.verification_token)
    authed = AppsService.authenticate_api_key(test_session, app_res.api_key)
    assert authed.app_id == app_res.app_id

    skill_req = SkillCreateRequest(
        name="svc-skill",
        description="test service skill",
        instructions="# test",
        tags=["svc"],
    )
    created = SkillsService.create_skill(test_session, authed.app_id, skill_req)
    assert created.name == "svc-skill"

    fetched = SkillsService.get_skill(test_session, "svc-skill")
    assert fetched.id == created.id

    listed = SkillsService.list_skills(test_session)
    assert len(listed) == 1

    updated = SkillsService.update_skill(
        test_session,
        created.id,
        authed.app_id,
        SkillUpdateRequest(description="new desc"),
    )
    assert updated.description == "new desc"

    del_res = SkillsService.delete_skill(test_session, created.id, authed.app_id)
    assert del_res["status"] == "success"


@pytest.mark.asyncio
async def test_grpc_servicers() -> None:
    """Verifies AppServiceServicer and SkillServiceServicer gRPC endpoints."""
    from unittest.mock import AsyncMock, MagicMock
    from api.v1 import skill_pb2, skill_service_pb2
    from services.grpc_servicers import AppServiceServicer, SkillServiceServicer
    from data import get_engine, init_db, reset_engine

    import uuid

    reset_engine()
    engine = get_engine("sqlite:///:memory:")
    init_db(engine)
    app_srv = AppServiceServicer()
    skill_srv = SkillServiceServicer()
    ctx = MagicMock()
    ctx.abort = AsyncMock()

    unique_email = f"grpc-{uuid.uuid4()}@example.com"
    reg_req = skill_service_pb2.RegisterAppRequest(app_name="gRPCApp", email=unique_email)
    reg_res = await app_srv.RegisterApp(reg_req, ctx)
    assert reg_res.app_name == "gRPCApp"

    ver_req = skill_service_pb2.VerifyAppRequest(token=reg_res.verification_token)
    ver_res = await app_srv.VerifyApp(ver_req, ctx)
    assert ver_res.is_active is True

    skill_req = skill_service_pb2.RegisterSkillRequest(
        api_key=reg_res.api_key,
        name="grpc-skill",
        description="grpc desc",
        instructions="# inst",
        tags=["grpc"],
        trigger_phrases=["test grpc"],
    )
    skill_res = await skill_srv.RegisterSkill(skill_req, ctx)
    assert skill_res.name == "grpc-skill"

    get_res = await skill_srv.GetSkill(
        skill_service_pb2.GetSkillRequest(skill_id_or_name="grpc-skill"), ctx
    )
    assert get_res.name == "grpc-skill"

    list_res = await skill_srv.ListSkills(
        skill_service_pb2.ListSkillsRequest(s="grpc"), ctx
    )
    assert len(list_res.skills) >= 1

    upd_res = await skill_srv.UpdateSkill(
        skill_service_pb2.UpdateSkillRequest(
            api_key=reg_res.api_key,
            skill_id=skill_res.compiled_reference.skill_id,
            description="updated grpc desc",
        ),
        ctx,
    )
    assert upd_res.description == "updated grpc desc"

    del_res = await skill_srv.DeleteSkill(
        skill_service_pb2.DeleteSkillRequest(
            api_key=reg_res.api_key, skill_id=skill_res.compiled_reference.skill_id
        ),
        ctx,
    )
    assert del_res.status == "success"

