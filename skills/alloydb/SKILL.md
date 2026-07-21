---
name: alloydb
description: Elite Google Cloud AlloyDB for PostgreSQL & pgvector SDLC. Covers Testcontainers TDD, migration CI/CD validation, HTTP 429 connection rate backoff, private IP VPC security (CWE-200), and SSL enforcement.
---

# Google Cloud AlloyDB & pgvector SDLC Skill

This skill prescribes best practices for connecting to, scaling, testing, securing, and managing relational workloads and vector similarity search on **Google Cloud AlloyDB for PostgreSQL**.

## Prerequisites & Pre-Flight Checklist

1. Google Cloud VPC configured with AlloyDB cluster provisioned on Private IP.
2. `vector` extension enabled in PostgreSQL database.

## HTTP 429 Rate Limit & Connection Quota Invariants

- Outbound database connection attempts during pool saturation must use exponential backoff to handle rate limits and connection exhaustion.

## Security Checkpoints & CWE Invariants

- **CWE-200 (Exposure of Sensitive Information)**: AlloyDB clusters MUST use Private IP inside the GCP VPC; public IP endpoints are strictly prohibited.
- **CWE-319 (Cleartext Transmission)**: Require `sslmode=require` or `verify-ca` for all application connections.
- **CWE-400 (Uncontrolled Resource Consumption)**: Pin connection pool bounds (`pool_size=10`, `max_overflow=20`) to prevent database connection exhaustion.

## 3-Phase Execution Protocol

### Phase 1: Enable pgvector & HNSW Indexing
Create vector table with 768-dimension embeddings and build HNSW cosine distance index.

### Phase 2: Implement TDD Suite with Testcontainers
Write integration tests against live PostgreSQL/pgvector instances spun up via **Testcontainers**.

### Phase 3: Validate Migration in CI/CD & Deploy
```bash
uv run pytest tests/test_database.py
alembic upgrade head
```

## Progressive Disclosure References

- **AlloyDB & pgvector Guide**: Read [`references/alloydb_pgvector.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/alloydb/references/alloydb_pgvector.md).
- **Reference Database Setup**: View [`examples/db.py`](file:///Users/rmcguinness/Projects/skill-builder/skills/alloydb/examples/db.py).
