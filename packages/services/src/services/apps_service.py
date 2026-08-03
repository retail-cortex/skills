"""Service layer for application registration, email verification, and API key authentication."""

from typing import Optional
from sqlmodel import Session
from data.apps_repository import AppsRepository
from model.app import (
    RegisteredApp,
    AppRegisterRequest,
    AppRegisterResponse,
    AppVerifyResponse,
)


class AppsService:
    """Service layer orchestrating application registration and verification."""

    @classmethod
    def register_app(
        cls, session: Session, request: AppRegisterRequest, base_url: Optional[str] = None
    ) -> AppRegisterResponse:
        """Registers a new application and returns credentials."""
        return AppsRepository.register_app(session, request, base_url)

    @classmethod
    def verify_app(cls, session: Session, token: str) -> AppVerifyResponse:
        """Verifies an application email token."""
        return AppsRepository.verify_app(session, token)

    @classmethod
    def authenticate_api_key(cls, session: Session, api_key: str) -> RegisteredApp:
        """Validates API key and returns the active RegisteredApp."""
        return AppsRepository.authenticate_api_key(session, api_key)


AppsController = AppsService
