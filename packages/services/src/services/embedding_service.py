"""Gemini Embedding Service using google-genai SDK with credential fallback."""

import logging
import math
import os
from typing import List, Optional

logger = logging.getLogger("skills_service.embedding_service")


class EmbeddingService:
    """Service for generating Gemini text embeddings and computing vector similarity."""

    def __init__(
        self,
        model_name: Optional[str] = None,
        gemini_api_key: Optional[str] = None,
    ) -> None:
        self.model_name = model_name or os.getenv("EMBEDDING_MODEL", "text-embedding-004")
        self.gemini_api_key = gemini_api_key or os.getenv("GEMINI_API_KEY")
        self.client = self._init_client()

    def _init_client(self) -> Optional[object]:
        """Initializes google-genai client with GCP project credentials first, then GEMINI_API_KEY, and fails gracefully."""
        try:
            from google import genai

            try:
                client = genai.Client()
                logger.info("Initialized Gemini GenAI client using Google Application Default Credentials.")
                return client
            except Exception as adc_err:
                logger.debug("ADC credential initialization failed, trying GEMINI_API_KEY: %s", adc_err)

            if self.gemini_api_key:
                client = genai.Client(api_key=self.gemini_api_key)
                logger.info("Initialized Gemini GenAI client using GEMINI_API_KEY.")
                return client

            logger.warning("Neither GCP project credentials nor GEMINI_API_KEY are configured. Embedding service disabled.")
            return None
        except Exception as exc:
            logger.warning("Failed to initialize google-genai client gracefully: %s", exc)
            return None

    def generate_embedding(self, text: str) -> Optional[List[float]]:
        """Generates embedding vector for text using Gemini text-embedding model."""
        if not self.client or not text.strip():
            return None

        try:
            response = self.client.models.embed_content(
                model=self.model_name,
                contents=text,
            )
            if hasattr(response, "embedding") and response.embedding and hasattr(response.embedding, "values"):
                return list(response.embedding.values)
            elif hasattr(response, "embeddings") and response.embeddings and len(response.embeddings) > 0:
                return list(response.embeddings[0].values)
            return None
        except Exception as exc:
            logger.warning("Error generating embedding from Gemini API: %s", exc)
            return None

    @staticmethod
    def cosine_similarity(vec1: List[float], vec2: List[float]) -> float:
        """Computes cosine similarity between two embedding vectors."""
        if not vec1 or not vec2 or len(vec1) != len(vec2):
            return 0.0
        dot_product = sum(a * b for a, b in zip(vec1, vec2))
        norm1 = math.sqrt(sum(a * a for a in vec1))
        norm2 = math.sqrt(sum(b * b for b in vec2))
        if norm1 == 0.0 or norm2 == 0.0:
            return 0.0
        return dot_product / (norm1 * norm2)


embedding_service = EmbeddingService()
