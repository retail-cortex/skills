"""Service layer for AI Agent skill persistence, versioning, and Gemini semantic vector search."""

from typing import Any, Dict, List, Optional
from sqlmodel import Session
from data.skills_repository import SkillsRepository
from model.skill import SkillCreateRequest, SkillResponse, SkillUpdateRequest
from services.embedding_service import embedding_service


class SkillsService:
    """Service layer orchestrating CRUD operations and Gemini vector search for skills."""

    @classmethod
    def list_skills(
        cls, session: Session, query: Optional[str] = None
    ) -> List[SkillResponse]:
        """Lists skills with optional Gemini semantic search."""
        query_vector = None
        if query and query.strip():
            query_vector = embedding_service.generate_embedding(query)
        return SkillsRepository.list_skills(
            session,
            query=query,
            query_vector=query_vector,
            cosine_similarity_fn=embedding_service.cosine_similarity,
        )

    @classmethod
    def get_skill(cls, session: Session, skill_id_or_name: str) -> SkillResponse:
        """Retrieves a skill by ID or unique name."""
        return SkillsRepository.get_skill(session, skill_id_or_name)

    @classmethod
    def create_skill(
        cls, session: Session, app_id: str, request: SkillCreateRequest
    ) -> SkillResponse:
        """Creates a skill and generates Gemini vector embeddings."""
        trigger_str = " ".join(request.trigger_phrases)
        text_for_embedding = f"{request.name} {request.description} {request.instructions} {trigger_str}"
        vector = embedding_service.generate_embedding(text_for_embedding)
        return SkillsRepository.create_skill(
            session,
            app_id=app_id,
            request=request,
            embedding_vector=vector,
            model_name=embedding_service.model_name,
        )

    @classmethod
    def update_skill(
        cls,
        session: Session,
        skill_id: str,
        app_id: str,
        request: SkillUpdateRequest,
        full_replace: bool = False,
    ) -> SkillResponse:
        """Updates a skill and regenerates vector embeddings."""
        existing = SkillsRepository.get_skill(session, skill_id)
        name = existing.name
        desc = request.description or existing.description
        inst = request.instructions or existing.instructions
        trig = request.trigger_phrases if request.trigger_phrases is not None else existing.trigger_phrases
        trigger_str = " ".join(trig)
        text_for_embedding = f"{name} {desc} {inst} {trigger_str}"
        vector = embedding_service.generate_embedding(text_for_embedding)

        return SkillsRepository.update_skill(
            session,
            skill_id=skill_id,
            app_id=app_id,
            request=request,
            full_replace=full_replace,
            embedding_vector=vector,
            model_name=embedding_service.model_name,
        )

    @classmethod
    def delete_skill(cls, session: Session, skill_id: str, app_id: str) -> Dict[str, Any]:
        """Deletes a skill and its sub-entities."""
        return SkillsRepository.delete_skill(session, skill_id=skill_id, app_id=app_id)


SkillsController = SkillsService
