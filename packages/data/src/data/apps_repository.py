"""Repository managing application registration, verification, and API key authentication."""

import hashlib
import os
import secrets
from datetime import datetime, timezone
from typing import Optional
from sqlmodel import Session, select
from fastapi import HTTPException, status

from model.app import (
    RegisteredApp,
    AppRegisterRequest,
    AppRegisterResponse,
    AppVerifyResponse,
)


class AppsRepository:
    """Repository for application lifecycle and credential verification."""

    @staticmethod
    def _hash_api_key(api_key: str) -> str:
        """Computes SHA-256 hash of API key for secure DB storage."""
        return hashlib.sha256(api_key.encode("utf-8")).hexdigest()

    @classmethod
    def register_app(
        cls, session: Session, request: AppRegisterRequest, base_url: Optional[str] = None
    ) -> AppRegisterResponse:
        """Registers a new application, issuing API key and email verification token."""
        statement = select(RegisteredApp).where(RegisteredApp.email == request.email)
        existing_app = session.exec(statement).first()
        if existing_app:
            raise HTTPException(
                status_code=status.HTTP_400_BAD_REQUEST,
                detail=f"Application with email '{request.email}' is already registered.",
            )

        raw_api_key = f"sk_live_{secrets.token_hex(24)}"
        api_key_hash = cls._hash_api_key(raw_api_key)

        app = RegisteredApp(
            app_name=request.app_name,
            email=request.email,
            api_key_hash=api_key_hash,
            is_active=False,
        )
        session.add(app)
        session.commit()
        session.refresh(app)

        default_base_url = os.getenv("BASE_URL", "http://localhost:8000")
        host_url = (base_url or default_base_url).rstrip("/")
        verification_url = f"{host_url}/api/v1/apps/verify?token={app.verification_token}"

        return AppRegisterResponse(
            app_id=app.app_id,
            app_name=app.app_name,
            email=app.email,
            api_key=raw_api_key,
            verification_token=app.verification_token,
            verification_url=verification_url,
        )

    @classmethod
    def verify_app(cls, session: Session, token: str) -> AppVerifyResponse:
        """Activates a registered application using its verification token."""
        statement = select(RegisteredApp).where(RegisteredApp.verification_token == token)
        app = session.exec(statement).first()

        if not app:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail="Invalid or expired verification token.",
            )

        app.is_active = True
        app.verified_at = datetime.now(timezone.utc)
        session.add(app)
        session.commit()
        session.refresh(app)

        return AppVerifyResponse(
            app_id=app.app_id,
            app_name=app.app_name,
            email=app.email,
            is_active=app.is_active,
            message="Application email verified successfully. Account is now active.",
        )

    @classmethod
    def authenticate_api_key(cls, session: Session, api_key: str) -> RegisteredApp:
        """Validates API key and ensures the associated application is active."""
        if not api_key:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Missing API Key in header 'X-API-Key'.",
            )

        key_hash = cls._hash_api_key(api_key)
        statement = select(RegisteredApp).where(RegisteredApp.api_key_hash == key_hash)
        app = session.exec(statement).first()

        if not app:
            raise HTTPException(
                status_code=status.HTTP_401_UNAUTHORIZED,
                detail="Invalid API Key provided.",
            )

        if not app.is_active:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Application is pending email verification. Please verify your app before registering skills.",
            )

        return app


# Backwards-compatibility alias
AppsController = AppsRepository
