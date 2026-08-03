"""Repository handling CRUD operations and database persistence for skills."""

import json
import math
from datetime import datetime, timezone
from typing import Any, Callable, Dict, List, Optional, Tuple
from sqlmodel import Session, select
from fastapi import HTTPException, status

from loader.compiler import SkillCompiler
from loader.types import SkillDefinition

from model.skill import (
    Skill,
    SkillVersion,
    SkillMetadata,
    SkillResource,
    SkillExample,
    SkillEmbedding,
    SkillCreateRequest,
    SkillUpdateRequest,
    SkillResponse,
)


class SkillsRepository:
    """Repository handling CRUD persistence, versioning, sub-entities, and keyword/vector search for skills."""

    compiler = SkillCompiler(strict_schemas=True)

    @staticmethod
    def cosine_similarity(vec1: List[float], vec2: List[float]) -> float:
        """Computes cosine similarity between two numeric vectors."""
        if not vec1 or not vec2 or len(vec1) != len(vec2):
            return 0.0
        dot_product = sum(a * b for a, b in zip(vec1, vec2))
        norm1 = math.sqrt(sum(a * a for a in vec1))
        norm2 = math.sqrt(sum(b * b for b in vec2))
        if norm1 == 0.0 or norm2 == 0.0:
            return 0.0
        return dot_product / (norm1 * norm2)

    @classmethod
    def _build_skill_response(
        cls, session: Session, skill: Skill, similarity_score: Optional[float] = None
    ) -> SkillResponse:
        """Helper to construct a complete SkillResponse object with linked entities."""
        ver_stmt = (
            select(SkillVersion)
            .where(SkillVersion.skill_id == skill.id)
            .order_by(SkillVersion.created_at.desc())
        )
        latest_ver = session.exec(ver_stmt).first()
        version_str = latest_ver.version if latest_ver else "1.0.0"
        json_schema = json.loads(latest_ver.json_schema_json) if latest_ver else {}

        meta_stmt = select(SkillMetadata).where(SkillMetadata.skill_id == skill.id)
        meta_records = session.exec(meta_stmt).all()
        metadata_dict = {m.key: m.value for m in meta_records}

        res_stmt = select(SkillResource).where(SkillResource.skill_id == skill.id)
        res_records = session.exec(res_stmt).all()
        resources_dict = {r.name: r.content for r in res_records}

        ex_stmt = select(SkillExample).where(SkillExample.skill_id == skill.id)
        ex_records = session.exec(ex_stmt).all()
        examples_dict = {e.name: e.content for e in ex_records}

        tags_list = json.loads(skill.tags_json) if hasattr(skill, "tags_json") and skill.tags_json else []
        trigger_phrases_list = json.loads(skill.trigger_phrases_json) if hasattr(skill, "trigger_phrases_json") and skill.trigger_phrases_json else []

        return SkillResponse(
            id=skill.id,
            app_id=skill.app_id,
            name=skill.name,
            description=skill.description,
            instructions=skill.instructions,
            license=skill.license,
            author=skill.author,
            category=skill.category,
            tags=tags_list,
            trigger_phrases=trigger_phrases_list,
            version=version_str,
            sha256_hash=skill.sha256_hash,
            hitl_tier=skill.hitl_tier,
            json_schema=json_schema,
            metadata=metadata_dict,
            references=resources_dict,
            examples=examples_dict,
            created_at=skill.created_at,
            updated_at=skill.updated_at,
            similarity_score=similarity_score,
        )

    @classmethod
    def create_skill(
        cls,
        session: Session,
        app_id: str,
        request: SkillCreateRequest,
        embedding_vector: Optional[List[float]] = None,
        model_name: str = "text-embedding-004",
    ) -> SkillResponse:
        """Persists a new skill, compiling schema and storing optional embeddings."""
        raw_def = SkillDefinition(
            name=request.name,
            description=request.description,
            instructions=request.instructions,
            version=request.version or "1.0.0",
            license=request.license,
            author=request.author,
            category=request.category,
            tags=request.tags,
            trigger_phrases=request.trigger_phrases,
        )
        compiled_ref = cls.compiler.compile(raw_def)

        skill = Skill(
            app_id=app_id,
            name=request.name,
            description=request.description,
            instructions=request.instructions,
            license=request.license,
            author=request.author,
            category=request.category,
            sha256_hash=compiled_ref.sha256_hash,
            hitl_tier=compiled_ref.hitl_tier.value,
            tags_json=json.dumps(request.tags),
            trigger_phrases_json=json.dumps(request.trigger_phrases),
        )
        session.add(skill)
        session.commit()
        session.refresh(skill)

        schema_json = json.dumps(compiled_ref.json_schema)
        version_entry = SkillVersion(
            skill_id=skill.id,
            version=request.version or "1.0.0",
            json_schema_json=schema_json,
            sha256_hash=compiled_ref.sha256_hash,
        )
        session.add(version_entry)

        for key, value in request.metadata.items():
            session.add(SkillMetadata(skill_id=skill.id, key=key, value=value))

        for name, content in request.references.items():
            session.add(SkillResource(skill_id=skill.id, name=name, content=content))

        for name, content in request.examples.items():
            session.add(SkillExample(skill_id=skill.id, name=name, content=content))

        if embedding_vector:
            session.add(
                SkillEmbedding(
                    skill_id=skill.id,
                    embedding_json=json.dumps(embedding_vector),
                    model_name=model_name,
                )
            )

        session.commit()
        session.refresh(skill)
        return cls._build_skill_response(session, skill)

    @classmethod
    def get_skill(cls, session: Session, skill_id_or_name: str) -> SkillResponse:
        """Retrieves a skill by ID or unique name."""
        statement = select(Skill).where(
            (Skill.id == skill_id_or_name) | (Skill.name == skill_id_or_name)
        )
        skill = session.exec(statement).first()

        if not skill:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Skill '{skill_id_or_name}' not found.",
            )

        return cls._build_skill_response(session, skill)

    @classmethod
    def list_skills(
        cls,
        session: Session,
        query: Optional[str] = None,
        query_vector: Optional[List[float]] = None,
        cosine_similarity_fn: Optional[Callable[[List[float], List[float]], float]] = None,
    ) -> List[SkillResponse]:
        """Lists all skills, performing semantic vector matching or keyword matching."""
        skills = session.exec(select(Skill)).all()
        if not skills:
            return []

        if not query or not query.strip():
            return [cls._build_skill_response(session, s) for s in skills]

        if query_vector:
            sim_fn = cosine_similarity_fn or cls.cosine_similarity
            scored_skills: List[Tuple[Skill, float]] = []
            for skill in skills:
                emb_stmt = select(SkillEmbedding).where(SkillEmbedding.skill_id == skill.id)
                emb_record = session.exec(emb_stmt).first()
                if emb_record:
                    vector = json.loads(emb_record.embedding_json)
                    score = sim_fn(query_vector, vector)
                else:
                    score = 0.0
                scored_skills.append((skill, score))

            scored_skills.sort(key=lambda x: x[1], reverse=True)
            return [
                cls._build_skill_response(session, item[0], similarity_score=item[1])
                for item in scored_skills
            ]

        q_lower = query.lower()
        matched = [
            s for s in skills
            if q_lower in s.name.lower()
            or q_lower in s.description.lower()
            or q_lower in s.instructions.lower()
            or (s.category and q_lower in s.category.lower())
        ]
        return [cls._build_skill_response(session, s) for s in matched]

    @classmethod
    def update_skill(
        cls,
        session: Session,
        skill_id: str,
        app_id: str,
        request: SkillUpdateRequest,
        full_replace: bool = False,
        embedding_vector: Optional[List[float]] = None,
        model_name: str = "text-embedding-004",
    ) -> SkillResponse:
        """Updates or replaces an existing skill, enforcing app ownership and updating embeddings."""
        statement = select(Skill).where(Skill.id == skill_id)
        skill = session.exec(statement).first()

        if not skill:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Skill '{skill_id}' not found.",
            )

        if skill.app_id != app_id:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Unauthorized: Skill belongs to a different application.",
            )

        if request.description is not None or full_replace:
            skill.description = request.description or skill.description
        if request.instructions is not None or full_replace:
            skill.instructions = request.instructions or skill.instructions
        if request.license is not None or full_replace:
            skill.license = request.license
        if request.category is not None or full_replace:
            skill.category = request.category
        if request.tags is not None or full_replace:
            skill.tags_json = json.dumps(request.tags or [])
        if request.trigger_phrases is not None or full_replace:
            skill.trigger_phrases_json = json.dumps(request.trigger_phrases or [])

        skill.updated_at = datetime.now(timezone.utc)

        current_tags = json.loads(skill.tags_json) if hasattr(skill, "tags_json") and skill.tags_json else []
        current_trigger_phrases = json.loads(skill.trigger_phrases_json) if hasattr(skill, "trigger_phrases_json") and skill.trigger_phrases_json else []

        raw_def = SkillDefinition(
            name=skill.name,
            description=skill.description,
            instructions=skill.instructions,
            license=skill.license,
            author=skill.author,
            category=skill.category,
            tags=current_tags,
            trigger_phrases=current_trigger_phrases,
            version=request.version or "1.1.0",
        )
        compiled_ref = cls.compiler.compile(raw_def)
        skill.sha256_hash = compiled_ref.sha256_hash
        skill.hitl_tier = compiled_ref.hitl_tier.value

        session.add(skill)

        if request.version:
            session.add(
                SkillVersion(
                    skill_id=skill.id,
                    version=request.version,
                    json_schema_json=json.dumps(compiled_ref.json_schema),
                    sha256_hash=compiled_ref.sha256_hash,
                )
            )

        if request.metadata is not None:
            old_meta = session.exec(select(SkillMetadata).where(SkillMetadata.skill_id == skill.id)).all()
            for m in old_meta:
                session.delete(m)
            for k, v in request.metadata.items():
                session.add(SkillMetadata(skill_id=skill.id, key=k, value=v))

        if request.references is not None:
            old_res = session.exec(select(SkillResource).where(SkillResource.skill_id == skill.id)).all()
            for r in old_res:
                session.delete(r)
            for name, content in request.references.items():
                session.add(SkillResource(skill_id=skill.id, name=name, content=content))

        if request.examples is not None:
            old_ex = session.exec(select(SkillExample).where(SkillExample.skill_id == skill.id)).all()
            for e in old_ex:
                session.delete(e)
            for name, content in request.examples.items():
                session.add(SkillExample(skill_id=skill.id, name=name, content=content))

        if embedding_vector:
            old_emb = session.exec(select(SkillEmbedding).where(SkillEmbedding.skill_id == skill.id)).first()
            if old_emb:
                session.delete(old_emb)

            session.add(
                SkillEmbedding(
                    skill_id=skill.id,
                    embedding_json=json.dumps(embedding_vector),
                    model_name=model_name,
                )
            )

        session.commit()
        session.refresh(skill)
        return cls._build_skill_response(session, skill)

    @classmethod
    def delete_skill(cls, session: Session, skill_id: str, app_id: str) -> Dict[str, Any]:
        """Deletes a skill and all linked sub-entities."""
        statement = select(Skill).where(Skill.id == skill_id)
        skill = session.exec(statement).first()

        if not skill:
            raise HTTPException(
                status_code=status.HTTP_404_NOT_FOUND,
                detail=f"Skill '{skill_id}' not found.",
            )

        if skill.app_id != app_id:
            raise HTTPException(
                status_code=status.HTTP_403_FORBIDDEN,
                detail="Unauthorized: Skill belongs to a different application.",
            )

        for table in [SkillVersion, SkillMetadata, SkillResource, SkillExample, SkillEmbedding]:
            records = session.exec(select(table).where(table.skill_id == skill.id)).all()
            for r in records:
                session.delete(r)

        session.delete(skill)
        session.commit()
        return {"status": "success", "message": f"Skill '{skill_id}' deleted successfully."}


# Backwards-compatibility alias
SkillsController = SkillsRepository
