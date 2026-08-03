"""FastAPI REST router for skills CRUD and semantic search."""

from typing import List, Optional
from fastapi import APIRouter, Header, Depends, Query, status
from sqlmodel import Session

from data import get_session
from model.app import RegisteredApp
from model.skill import (
    SkillCreateRequest,
    SkillUpdateRequest,
    SkillResponse,
)
from services.apps_service import AppsService as AppsController
from services.skills_service import SkillsService as SkillsController

router = APIRouter(prefix="/skills", tags=["Skills"])


def get_current_app(
    x_api_key: Optional[str] = Header(None, alias="X-API-Key"),
    session: Session = Depends(get_session),
) -> RegisteredApp:
    """Dependency verifying the X-API-Key header and returning the active RegisteredApp."""
    return AppsController.authenticate_api_key(session, x_api_key or "")


@router.get("", response_model=List[SkillResponse])
def list_skills(
    s: Optional[str] = Query(None, description="Optional search query string for Gemini semantic vector search"),
    session: Session = Depends(get_session),
) -> List[SkillResponse]:
    """Lists registered skills or executes Gemini semantic vector search if 's' query parameter is provided."""
    return SkillsController.list_skills(session, query=s)


@router.get("/{skill_id_or_name}", response_model=SkillResponse)
def get_skill(
    skill_id_or_name: str,
    session: Session = Depends(get_session),
) -> SkillResponse:
    """Fetches a specific skill by its ID or unique skill name."""
    return SkillsController.get_skill(session, skill_id_or_name)


@router.post("", response_model=SkillResponse, status_code=status.HTTP_201_CREATED)
def register_skill(
    payload: SkillCreateRequest,
    current_app: RegisteredApp = Depends(get_current_app),
    session: Session = Depends(get_session),
) -> SkillResponse:
    """Compiles and registers a new skill associated with the authenticated application."""
    return SkillsController.create_skill(session, current_app.app_id, payload)


@router.put("/{skill_id}", response_model=SkillResponse)
def replace_skill(
    skill_id: str,
    payload: SkillUpdateRequest,
    current_app: RegisteredApp = Depends(get_current_app),
    session: Session = Depends(get_session),
) -> SkillResponse:
    """Replaces an existing skill and creates a new version entry."""
    return SkillsController.update_skill(
        session, skill_id, current_app.app_id, payload, full_replace=True
    )


@router.patch("/{skill_id}", response_model=SkillResponse)
def update_skill(
    skill_id: str,
    payload: SkillUpdateRequest,
    current_app: RegisteredApp = Depends(get_current_app),
    session: Session = Depends(get_session),
) -> SkillResponse:
    """Partially updates an existing skill."""
    return SkillsController.update_skill(
        session, skill_id, current_app.app_id, payload, full_replace=False
    )


@router.delete("/{skill_id}")
def delete_skill(
    skill_id: str,
    current_app: RegisteredApp = Depends(get_current_app),
    session: Session = Depends(get_session),
) -> dict:
    """Deletes a skill and its associated version, metadata, and embedding records."""
    return SkillsController.delete_skill(session, skill_id, current_app.app_id)
