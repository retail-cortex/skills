"""Main FastAPI application entry point for Skills Service with async gRPC server."""

import logging
from contextlib import asynccontextmanager
from typing import AsyncGenerator
from fastapi import FastAPI
import grpc
import uvicorn

from api.v1 import skill_service_pb2_grpc
from skills_app.config import settings
from data import init_db
from services import setup_telemetry, SkillServiceServicer, AppServiceServicer
from skills_app.api.v1.apps import router as apps_router
from skills_app.api.v1.skills import router as skills_router
from skills_mcp import mcp_server

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("skills_service.main")


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    """App lifespan context manager handling DB init, OTel setup, and async gRPC server execution."""
    logger.info("Initializing database tables...")
    init_db()

    logger.info("Initializing OpenTelemetry telemetry...")
    setup_telemetry(
        enable_telemetry=settings.enable_opentelemetry,
        service_name=settings.otel_service_name,
        gcp_project_id=settings.gcp_project_id,
    )

    # Initialize and start Async gRPC Server on FastAPI event loop
    grpc_address = f"{settings.host}:{settings.grpc_port}"
    logger.info("Starting Async gRPC Server on %s...", grpc_address)
    grpc_server = grpc.aio.server()
    skill_service_pb2_grpc.add_SkillServiceServicer_to_server(SkillServiceServicer(), grpc_server)
    skill_service_pb2_grpc.add_AppServiceServicer_to_server(AppServiceServicer(), grpc_server)
    grpc_server.add_insecure_port(grpc_address)
    await grpc_server.start()

    yield

    logger.info("Shutting down Async gRPC Server...")
    await grpc_server.stop(grace=5)
    logger.info("Shutdown complete.")


app = FastAPI(
    title="Skills Service API",
    description="Enterprise FastAPI + SQLModel + FastMCP + gRPC server for AI Agent Skills.",
    version="0.1.0",
    lifespan=lifespan,
)

# Include v1 REST Routers
app.include_router(apps_router, prefix="/api/v1")
app.include_router(skills_router, prefix="/api/v1")


@app.get("/health", tags=["Health"])
def health_check() -> dict:
    """Basic health check endpoint."""
    return {
        "status": "ok",
        "service": "skills-service",
        "version": "0.1.0",
        "ports": {
            "rest": settings.port,
            "grpc": settings.grpc_port,
        },
    }


# Mount FastMCP HTTP/SSE application at /mcp
try:
    mcp_app = mcp_server.http_app()
    app.mount("/mcp", mcp_app)
    logger.info("FastMCP mounted successfully at /mcp")
except Exception as mcp_err:
    logger.warning("Could not mount FastMCP HTTP app directly: %s", mcp_err)


def main() -> None:
    """Main execution function to run Uvicorn web server."""
    logger.info("Starting Uvicorn server on %s:%d (REST) & gRPC on port %d...", settings.host, settings.port, settings.grpc_port)
    uvicorn.run(
        "skills_app.main:app",
        host=settings.host,
        port=settings.port,
        reload=False,
    )


if __name__ == "__main__":
    main()
