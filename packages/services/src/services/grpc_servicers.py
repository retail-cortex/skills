"""Async gRPC Servicer implementations for SkillService and AppService using Protobuf."""

import logging
from google.protobuf.json_format import ParseDict, MessageToDict
import grpc
from sqlmodel import Session

from api.v1 import skill_pb2, skill_service_pb2, skill_service_pb2_grpc
from data.database import engine
from model.app import AppRegisterRequest
from model.skill import SkillCreateRequest, SkillUpdateRequest
from services.apps_service import AppsService as AppsController
from services.skills_service import SkillsService as SkillsController

logger = logging.getLogger("skills_service.grpc_servicers")


async def handle_grpc_error(exc: Exception, context: grpc.aio.ServicerContext) -> None:
    """Helper mapping exceptions to appropriate gRPC status codes and aborting context."""
    logger.error("gRPC execution error: %s", exc)
    from fastapi import HTTPException

    if isinstance(exc, HTTPException):
        code_map = {
            400: grpc.StatusCode.INVALID_ARGUMENT,
            401: grpc.StatusCode.UNAUTHENTICATED,
            403: grpc.StatusCode.PERMISSION_DENIED,
            404: grpc.StatusCode.NOT_FOUND,
        }
        status_code = code_map.get(exc.status_code, grpc.StatusCode.INTERNAL)
        await context.abort(status_code, str(exc.detail))
    elif isinstance(exc, (ValueError, TypeError)):
        await context.abort(grpc.StatusCode.INVALID_ARGUMENT, str(exc))
    else:
        await context.abort(grpc.StatusCode.INTERNAL, str(exc))


class AppServiceServicer(skill_service_pb2_grpc.AppServiceServicer):
    """gRPC Servicer handling application registration and email token verification."""

    async def RegisterApp(
        self, request: skill_service_pb2.RegisterAppRequest, context: grpc.aio.ServicerContext
    ) -> skill_service_pb2.RegisterAppResponse:
        """gRPC handler for registering an application."""
        try:
            with Session(engine) as session:
                req = AppRegisterRequest(app_name=request.app_name, email=request.email)
                res = AppsController.register_app(session, req)
                return skill_service_pb2.RegisterAppResponse(
                    app_id=res.app_id,
                    app_name=res.app_name,
                    email=res.email,
                    api_key=res.api_key,
                    verification_token=res.verification_token,
                    verification_url=res.verification_url,
                )
        except Exception as exc:
            await handle_grpc_error(exc, context)

    async def VerifyApp(
        self, request: skill_service_pb2.VerifyAppRequest, context: grpc.aio.ServicerContext
    ) -> skill_service_pb2.VerifyAppResponse:
        """gRPC handler for verifying application email token."""
        try:
            with Session(engine) as session:
                res = AppsController.verify_app(session, request.token)
                return skill_service_pb2.VerifyAppResponse(
                    app_id=res.app_id,
                    app_name=res.app_name,
                    email=res.email,
                    is_active=res.is_active,
                    message=res.message,
                )
        except Exception as exc:
            await handle_grpc_error(exc, context)


class SkillServiceServicer(skill_service_pb2_grpc.SkillServiceServicer):
    """gRPC Servicer handling skill CRUD operations and Gemini vector search."""

    @staticmethod
    def _to_proto_skill(skill_res) -> skill_pb2.SkillDefinition:
        """Helper mapping SkillResponse object to SkillDefinition Protobuf message."""
        data_dict = skill_res.model_dump(mode="json")

        proto_skill = skill_pb2.SkillDefinition()

        json_schema = data_dict.pop("json_schema", {})
        hitl_tier = data_dict.pop("hitl_tier", "TIER_1_AUTO_READ")
        sha256_hash = data_dict.get("sha256_hash", "")
        skill_id = data_dict.pop("id", "")

        proto_skill.name = data_dict.get("name", "")
        proto_skill.description = data_dict.get("description", "")
        proto_skill.instructions = data_dict.get("instructions", "")
        proto_skill.license = data_dict.get("license") or ""
        proto_skill.author = data_dict.get("author") or ""
        proto_skill.version = data_dict.get("version") or "1.0.0"
        proto_skill.category = data_dict.get("category") or ""

        for k, v in data_dict.get("metadata", {}).items():
            proto_skill.metadata[k] = str(v)
        for k, v in data_dict.get("references", {}).items():
            proto_skill.references[k] = str(v)
        for k, v in data_dict.get("examples", {}).items():
            proto_skill.examples[k] = str(v)

        for tag in data_dict.get("tags", []):
            proto_skill.tags.append(str(tag))
        for tp in data_dict.get("trigger_phrases", []):
            proto_skill.trigger_phrases.append(str(tp))

        if sha256_hash:
            proto_skill.compiled_reference.skill_id = skill_id
            proto_skill.compiled_reference.name = proto_skill.name
            proto_skill.compiled_reference.description = proto_skill.description
            proto_skill.compiled_reference.sha256_hash = sha256_hash
            proto_skill.compiled_reference.json_schema = str(json_schema)

        return proto_skill

    async def ListSkills(
        self, request: skill_service_pb2.ListSkillsRequest, context: grpc.aio.ServicerContext
    ) -> skill_service_pb2.ListSkillsResponse:
        """gRPC handler for listing and searching skills."""
        try:
            with Session(engine) as session:
                skills = SkillsController.list_skills(session, query=request.s)
                proto_list = [self._to_proto_skill(s) for s in skills]
                return skill_service_pb2.ListSkillsResponse(skills=proto_list)
        except Exception as exc:
            await handle_grpc_error(exc, context)

    async def GetSkill(
        self, request: skill_service_pb2.GetSkillRequest, context: grpc.aio.ServicerContext
    ) -> skill_pb2.SkillDefinition:
        """gRPC handler for fetching a specific skill."""
        try:
            with Session(engine) as session:
                skill = SkillsController.get_skill(session, request.skill_id_or_name)
                return self._to_proto_skill(skill)
        except Exception as exc:
            await handle_grpc_error(exc, context)

    async def RegisterSkill(
        self, request: skill_service_pb2.RegisterSkillRequest, context: grpc.aio.ServicerContext
    ) -> skill_pb2.SkillDefinition:
        """gRPC handler for registering a new skill."""
        try:
            with Session(engine) as session:
                app = AppsController.authenticate_api_key(session, request.api_key)
                req = SkillCreateRequest(
                    name=request.name,
                    description=request.description,
                    instructions=request.instructions,
                    license=request.license or None,
                    author=request.author or None,
                    version=request.version or "1.0.0",
                    category=request.category or None,
                    tags=list(request.tags),
                    trigger_phrases=list(request.trigger_phrases),
                    metadata=dict(request.metadata),
                    references=dict(request.references),
                    examples=dict(request.examples),
                )
                skill = SkillsController.create_skill(session, app.app_id, req)
                return self._to_proto_skill(skill)
        except Exception as exc:
            await handle_grpc_error(exc, context)

    async def UpdateSkill(
        self, request: skill_service_pb2.UpdateSkillRequest, context: grpc.aio.ServicerContext
    ) -> skill_pb2.SkillDefinition:
        """gRPC handler for updating an existing skill."""
        try:
            with Session(engine) as session:
                app = AppsController.authenticate_api_key(session, request.api_key)
                req = SkillUpdateRequest(
                    description=request.description or None,
                    instructions=request.instructions or None,
                    license=request.license or None,
                    category=request.category or None,
                    version=request.version or None,
                    metadata=dict(request.metadata) if request.metadata else None,
                    references=dict(request.references) if request.references else None,
                    examples=dict(request.examples) if request.examples else None,
                )
                skill = SkillsController.update_skill(session, request.skill_id, app.app_id, req)
                return self._to_proto_skill(skill)
        except Exception as exc:
            await handle_grpc_error(exc, context)

    async def DeleteSkill(
        self, request: skill_service_pb2.DeleteSkillRequest, context: grpc.aio.ServicerContext
    ) -> skill_service_pb2.DeleteSkillResponse:
        """gRPC handler for deleting a skill."""
        try:
            with Session(engine) as session:
                app = AppsController.authenticate_api_key(session, request.api_key)
                res = SkillsController.delete_skill(session, request.skill_id, app.app_id)
                return skill_service_pb2.DeleteSkillResponse(
                    status=res["status"], message=res["message"]
                )
        except Exception as exc:
            await handle_grpc_error(exc, context)
