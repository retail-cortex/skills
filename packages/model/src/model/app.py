"""SQLModel schema for registered applications."""

from datetime import datetime, timezone
import uuid
from typing import Optional
from sqlmodel import Field, SQLModel
from pydantic import EmailStr


class RegisteredApp(SQLModel, table=True):
    """Represents a registered application authorized to register and manage skills."""

    __tablename__ = "registered_apps"

    app_id: str = Field(default_factory=lambda: str(uuid.uuid4()), primary_key=True)
    app_name: str = Field(index=True)
    email: str = Field(index=True)
    api_key_hash: str = Field(index=True)
    is_active: bool = Field(default=False, index=True)
    verification_token: str = Field(default_factory=lambda: str(uuid.uuid4()), index=True)
    created_at: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))
    verified_at: Optional[datetime] = Field(default=None)


class AppRegisterRequest(SQLModel):
    """Payload for registering a new application."""

    app_name: str
    email: EmailStr


class AppRegisterResponse(SQLModel):
    """Response returned upon application registration."""

    app_id: str
    app_name: str
    email: str
    api_key: str
    verification_token: str
    verification_url: str


class AppVerifyResponse(SQLModel):
    """Response returned upon email verification of an application."""

    app_id: str
    app_name: str
    email: str
    is_active: bool
    message: str
