"""FastAPI router for application registration and verification."""

from fastapi import APIRouter, Depends, Query, Request
from sqlmodel import Session

from data import get_session
from model.app import (
    AppRegisterRequest,
    AppRegisterResponse,
    AppVerifyResponse,
)
from services.apps_service import AppsService as AppsController

router = APIRouter(prefix="/apps", tags=["Applications"])


@router.post("/register", response_model=AppRegisterResponse, status_code=201)
def register_application(
    payload: AppRegisterRequest,
    request: Request,
    session: Session = Depends(get_session),
) -> AppRegisterResponse:
    """Registers a new application and issues an API key and verification link."""
    base_url = str(request.base_url).rstrip("/")
    return AppsController.register_app(session, payload, base_url=base_url)


@router.get("/verify", response_model=AppVerifyResponse)
def verify_application(
    token: str = Query(..., description="Verification token sent via email"),
    session: Session = Depends(get_session),
) -> AppVerifyResponse:
    """Verifies and activates a registered application."""
    return AppsController.verify_app(session, token)
