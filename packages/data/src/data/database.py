"""Database engine and session management using SQLModel."""

import os
from typing import Any, Dict, Generator, Optional
from sqlmodel import SQLModel, Session, create_engine
from sqlalchemy.engine import Engine

_engine: Optional[Engine] = None


def get_database_url() -> str:
    """Returns database URL from environment or fallback default."""
    return os.getenv("DATABASE_URL", "sqlite:///./skills.db")


def get_engine(database_url: Optional[str] = None, echo: bool = False) -> Engine:
    """Returns or creates the SQLModel/SQLAlchemy engine."""
    global _engine
    url = database_url or get_database_url()
    connect_args: Dict[str, Any] = {"check_same_thread": False} if "sqlite" in url else {}
    if _engine is None or str(_engine.url) != url:
        _engine = create_engine(url, echo=echo, connect_args=connect_args)
    return _engine


def reset_engine() -> None:
    """Resets the cached engine instance."""
    global _engine
    if _engine is not None:
        _engine.dispose()
        _engine = None


def init_db(engine_instance: Optional[Engine] = None) -> None:
    """Creates database tables if they do not exist."""
    eng = engine_instance or get_engine()
    SQLModel.metadata.create_all(eng)


def get_session() -> Generator[Session, None, None]:
    """Dependency generator returning a database session."""
    eng = get_engine()
    with Session(eng) as session:
        yield session


# Module-level default engine for backwards compatibility
engine = get_engine()
