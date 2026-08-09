# Copyright 2026 Ryan McGuinness
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import os
from typing import List, Optional
from uuid import UUID, uuid4
from sqlmodel import SQLModel, Field, Session, create_engine, select
from pgvector.sqlalchemy import Vector
from sqlalchemy import Column

class DocumentChunk(SQLModel, table=True):
    __tablename__ = "document_chunks"

    id: UUID = Field(default_factory=uuid4, primary_key=True)
    content: str = Field(description="Raw text content of the chunk")
    # 768-dimensional embedding from Vertex AI text-embedding-004
    embedding: List[float] = Field(sa_column=Column(Vector(768)))

# Build AlloyDB / PostgreSQL DSN
DB_USER = os.environ.get("ALLOYDB_USER", "postgres")
DB_PASS = os.environ.get("ALLOYDB_PASS", "secret")
DB_HOST = os.environ.get("ALLOYDB_HOST", "10.0.0.1") # Private IP
DB_PORT = os.environ.get("ALLOYDB_PORT", "5432")
DB_NAME = os.environ.get("ALLOYDB_NAME", "enterprise_db")

DATABASE_URL = f"postgresql://{DB_USER}:{DB_PASS}@{DB_HOST}:{DB_PORT}/{DB_NAME}"

engine = create_engine(
    DATABASE_URL,
    pool_size=10,
    max_overflow=20,
    pool_recycle=1800,
)

def search_similar_chunks(query_embedding: List[float], limit: int = 5) -> List[DocumentChunk]:
    with Session(engine) as session:
        statement = (
            select(DocumentChunk)
            .order_by(DocumentChunk.embedding.cosine_distance(query_embedding))
            .limit(limit)
        )
        return session.exec(statement).all()
