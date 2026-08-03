"""Compatibility wrapper exporting gRPC servicers and engine."""

from services.grpc_servicers import (
    AppServiceServicer,
    SkillServiceServicer,
    engine,
    handle_grpc_error,
)

__all__ = [
    "AppServiceServicer",
    "SkillServiceServicer",
    "engine",
    "handle_grpc_error",
]
