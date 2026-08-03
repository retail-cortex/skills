# AlloyDB & pgvector Optimization Patterns

## Vector Indexing (HNSW vs IVFFlat)

For production vector similarity in AlloyDB:
- Use **HNSW** (Hierarchical Navigable Small World) for maximum query speed and high recall.
- Use **IVFFlat** when memory overhead must be minimized.

```sql
-- Create vector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Define table with 768-dimension embeddings (matches text-embedding-004)
CREATE TABLE skill_vectors (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    embedding vector(768)
);

-- Cosine distance index
CREATE INDEX idx_skill_vector_hnsw 
ON skill_vectors 
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);
```

## Cosine Distance Query Syntax

Query top 5 nearest neighbors using the `<=>` cosine distance operator:

```sql
SELECT id, name, description, 1 - (embedding <=> '[0.012, -0.045, ...]') AS similarity
FROM skill_vectors
ORDER BY embedding <=> '[0.012, -0.045, ...]'
LIMIT 5;
```
