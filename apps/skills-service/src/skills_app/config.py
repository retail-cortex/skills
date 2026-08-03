"""Configuration management for Skills Service using dotenv."""

from pathlib import Path
from pydantic_settings import BaseSettings, SettingsConfigDict

BASE_DIR: Path = Path(__file__).resolve().parent.parent.parent


class Settings(BaseSettings):
    """Application settings loaded from .env file or environment variables."""

    model_config = SettingsConfigDict(
        env_file=str(BASE_DIR / ".env"),
        env_file_encoding="utf-8",
        extra="ignore",
    )

    port: int = 8080
    grpc_port: int = 50051
    host: str = "0.0.0.0"
    base_url: str = "http://localhost:8080"

    database_url: str = "sqlite:///./skills.db"

    gcp_project_id: str = ""
    gemini_api_key: str = ""
    embedding_model: str = "text-embedding-004"

    enable_opentelemetry: bool = False
    otel_service_name: str = "skills-service"


settings = Settings()
