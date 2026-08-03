"""SQLModel schema for skills, versions, metadata, resources, examples, and embeddings."""

from datetime import datetime, timezone
import uuid
from typing import Any, Dict, List, Optional
from sqlmodel import Field, SQLModel


class Skill(SQLModel, table=True):
    """Primary skill persistence record linked to registering app."""

    __tablename__ = "skills"

    id: str = Field(default_factory=lambda: str(uuid.uuid4()), primary_key=True)
    app_id: str = Field(foreign_key="registered_apps.app_id", index=True)
    name: str = Field(index=True)
    description: str
    instructions: str
    license: Optional[str] = Field(default=None)
    author: Optional[str] = Field(default=None)
    category: Optional[str] = Field(default=None, index=True)
    sha256_hash: str = Field(index=True)
    hitl_tier: str = Field(default="TIER_1_AUTO_READ")
    tags_json: str = Field(default="[]")
    trigger_phrases_json: str = Field(default="[]")
    created_at: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))
    updated_at: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))


class SkillVersion(SQLModel, table=True):
    """Historical version entry for a skill."""

    __tablename__ = "skill_versions"

    id: str = Field(default_factory=lambda: str(uuid.uuid4()), primary_key=True)
    skill_id: str = Field(foreign_key="skills.id", index=True)
    version: str = Field(index=True)
    json_schema_json: str
    sha256_hash: str
    created_at: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))


class SkillMetadata(SQLModel, table=True):
    """Key-value metadata attribute linked to a skill."""

    __tablename__ = "skill_metadata"

    id: str = Field(default_factory=lambda: str(uuid.uuid4()), primary_key=True)
    skill_id: str = Field(foreign_key="skills.id", index=True)
    key: str = Field(index=True)
    value: str


class SkillResource(SQLModel, table=True):
    """Resource or reference file content linked to a skill."""

    __tablename__ = "skill_resources"

    id: str = Field(default_factory=lambda: str(uuid.uuid4()), primary_key=True)
    skill_id: str = Field(foreign_key="skills.id", index=True)
    name: str = Field(index=True)
    content: str


class SkillExample(SQLModel, table=True):
    """Usage example content linked to a skill."""

    __tablename__ = "skill_examples"

    id: str = Field(default_factory=lambda: str(uuid.uuid4()), primary_key=True)
    skill_id: str = Field(foreign_key="skills.id", index=True)
    name: str = Field(index=True)
    content: str


class SkillEmbedding(SQLModel, table=True):
    """Gemini vector embedding record for semantic search."""

    __tablename__ = "skill_embeddings"

    id: str = Field(default_factory=lambda: str(uuid.uuid4()), primary_key=True)
    skill_id: str = Field(foreign_key="skills.id", index=True)
    embedding_json: str
    model_name: str = Field(default="text-embedding-004")
    created_at: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))


# API Data Transfer Objects (DTOs)


class SkillCreateRequest(SQLModel):
    """Input payload for creating or registering a skill."""

    model_config = {"protected_namespaces": ()}

    name: str
    description: str
    instructions: str
    license: Optional[str] = None
    author: Optional[str] = None
    version: Optional[str] = "1.0.0"
    category: Optional[str] = None
    tags: List[str] = []
    trigger_phrases: List[str] = []
    metadata: Dict[str, str] = {}
    references: Dict[str, str] = {}
    examples: Dict[str, str] = {}


class SkillUpdateRequest(SQLModel):
    """Input payload for updating an existing skill."""

    model_config = {"protected_namespaces": ()}

    description: Optional[str] = None
    instructions: Optional[str] = None
    license: Optional[str] = None
    category: Optional[str] = None
    tags: Optional[List[str]] = None
    trigger_phrases: Optional[List[str]] = None
    version: Optional[str] = None
    metadata: Optional[Dict[str, str]] = None
    references: Optional[Dict[str, str]] = None
    examples: Optional[Dict[str, str]] = None


class SkillResponse(SQLModel):
    """Response DTO for skill serialization."""

    model_config = {"protected_namespaces": ()}

    id: str
    app_id: str
    name: str
    description: str
    instructions: str
    license: Optional[str]
    author: Optional[str]
    category: Optional[str]
    tags: List[str]
    trigger_phrases: List[str]
    version: str
    sha256_hash: str
    hitl_tier: str
    json_schema: Dict[str, Any]
    metadata: Dict[str, str]
    references: Dict[str, str]
    examples: Dict[str, str]
    created_at: datetime
    updated_at: datetime
    similarity_score: Optional[float] = None
