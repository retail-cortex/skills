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

"""Semantic Skill Discovery Engine using local TF-IDF / BM25 and Cosine Similarity.

Implements the Two-Step Directory Protocol to allow dynamic runtime tool retrieval without prompt context bloat.
"""

import math
import re
from typing import Any, Dict, List, Optional, Set, Tuple

from .compiler import SkillCompiler
from .types import CompiledSkillReference, SkillDefinition, SkillDirectorySearchResult


def tokenize(text: str) -> List[str]:
    """Tokenizes natural language text into normalized term tokens."""
    cleaned = re.sub(r"[^a-zA-Z0-9\s-]", " ", text.lower())
    tokens = [t for t in cleaned.split() if len(t) > 1]
    return tokens


class TFIDFVectorIndex:
    """Local TF-IDF / BM25 & Cosine Similarity search index over skill definitions."""

    def __init__(self):
        self.documents: Dict[str, CompiledSkillReference] = {}
        self.doc_tokens: Dict[str, List[str]] = {}
        self.df: Dict[str, int] = {}  # Document frequency of terms
        self.idf: Dict[str, float] = {}  # Inverse document frequency
        self.doc_vectors: Dict[str, Dict[str, float]] = {}
        self.vocab: Set[str] = set()
        self.num_docs: int = 0

    def index_skill(self, skill: SkillDefinition, compiled_ref: CompiledSkillReference):
        """Indexes a compiled skill reference using its metadata, tags, and description."""
        skill_id = compiled_ref.skill_id
        self.documents[skill_id] = compiled_ref

        # Synthesize document text from name, description, tags, trigger phrases, and category
        doc_text_parts = [
            skill.name * 3,  # Boost skill name
            skill.description * 2,
            " ".join(skill.tags) * 3,
            " ".join(skill.trigger_phrases) * 3,
            skill.category or "",
        ]
        doc_text = " ".join(doc_text_parts)
        tokens = tokenize(doc_text)
        self.doc_tokens[skill_id] = tokens

    def build(self):
        """Computes TF-IDF vectors for all indexed skills."""
        self.num_docs = len(self.documents)
        if self.num_docs == 0:
            return

        self.df = {}
        self.vocab = set()

        for skill_id, tokens in self.doc_tokens.items():
            unique_terms = set(tokens)
            for term in unique_terms:
                self.df[term] = self.df.get(term, 0) + 1
                self.vocab.add(term)

        # Calculate IDF values
        self.idf = {
            term: math.log((self.num_docs + 1) / (df_count + 0.5)) + 1.0
            for term, df_count in self.df.items()
        }

        # Calculate normalized TF-IDF vector for each document
        self.doc_vectors = {}
        for skill_id, tokens in self.doc_tokens.items():
            tf: Dict[str, int] = {}
            for t in tokens:
                tf[t] = tf.get(t, 0) + 1

            vec: Dict[str, float] = {}
            norm_sq = 0.0
            for term, count in tf.items():
                tfidf_val = (1 + math.log(count)) * self.idf.get(term, 1.0)
                vec[term] = tfidf_val
                norm_sq += tfidf_val * tfidf_val

            norm = math.sqrt(norm_sq) if norm_sq > 0 else 1.0
            self.doc_vectors[skill_id] = {k: v / norm for k, v in vec.items()}

    def search(self, intent_query: str, top_k: int = 5) -> List[Tuple[CompiledSkillReference, float]]:
        """Performs cosine similarity search against intent query."""
        if not self.doc_vectors:
            self.build()

        query_tokens = tokenize(intent_query)
        if not query_tokens:
            return []

        # Build query vector
        query_tf: Dict[str, int] = {}
        for t in query_tokens:
            query_tf[t] = query_tf.get(t, 0) + 1

        query_vec: Dict[str, float] = {}
        norm_sq = 0.0
        for term, count in query_tf.items():
            if term in self.idf:
                val = (1 + math.log(count)) * self.idf[term]
                query_vec[term] = val
                norm_sq += val * val

        if norm_sq == 0:
            return []

        norm = math.sqrt(norm_sq)
        query_vec = {k: v / norm for k, v in query_vec.items()}

        # Calculate dot product with document vectors
        scores: List[Tuple[CompiledSkillReference, float]] = []
        for skill_id, doc_vec in self.doc_vectors.items():
            score = 0.0
            for term, q_val in query_vec.items():
                if term in doc_vec:
                    score += q_val * doc_vec[term]

            if score > 0.01:
                scores.append((self.documents[skill_id], score))

        scores.sort(key=lambda x: x[1], reverse=True)
        return scores[:top_k]


class SkillDiscoveryEngine:
    """Orchestrates Two-Step Directory Protocol and intent search."""

    def __init__(self, compiler: Optional[SkillCompiler] = None):
        self.compiler = compiler or SkillCompiler()
        self.index = TFIDFVectorIndex()
        self.skills: Dict[str, SkillDefinition] = {}

    def register_skill(self, skill: SkillDefinition) -> CompiledSkillReference:
        """Registers and compiles a skill into the local search index."""
        compiled_ref = self.compiler.compile(skill)
        self.skills[skill.name] = skill
        self.index.index_skill(skill, compiled_ref)
        return compiled_ref

    def build_index(self):
        """Finalizes the TF-IDF vector index."""
        self.index.build()

    def search_skills(self, intent: str, top_k: int = 5) -> SkillDirectorySearchResult:
        """Executes natural language intent discovery search."""
        matches_with_scores = self.index.search(intent, top_k=top_k)
        matched_refs = [m[0] for m in matches_with_scores]
        return SkillDirectorySearchResult(
            query_intent=intent,
            matches=matched_refs,
            total_found=len(matched_refs),
        )

    def get_tool_declaration(self) -> Dict[str, Any]:
        """Returns tool schema definition for skill_directory_search."""
        return {
            "name": "skill_directory_search",
            "description": "Searches the secure registry for available skills matching execution intent. Returns compiled skill reference IDs.",
            "parameters": {
                "type": "object",
                "properties": {
                    "intent": {
                        "type": "string",
                        "description": "Natural language description of what you are trying to achieve (e.g., 'query postgres database').",
                    }
                },
                "required": ["intent"],
                "additionalProperties": False,
            },
        }
