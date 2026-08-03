# Skills Service (`apps/skills-service`)

Enterprise FastAPI + SQLModel + FastMCP server for persisting, compiling, and searching AI Agent skills with Gemini vector embeddings and Google Cloud OpenTelemetry instrumentation.

---

## Features

1. **REST CRUD APIs** (`/api/v1/skills`):
   - `GET /api/v1/skills` – List skills with optional Gemini semantic vector search (`?s={query}`).
   - `GET /api/v1/skills/{skill_id_or_name}` – Get skill by ID or name with complete sub-entities (versions, metadata, resources, examples).
   - `POST /api/v1/skills` – Register and compile new skill. Correlates `app_id` and computes Gemini embeddings.
   - `PUT /api/v1/skills/{skill_id}` – Replace skill and create new version entry.
   - `PATCH /api/v1/skills/{skill_id}` – Partial skill update.
   - `DELETE /api/v1/skills/{skill_id}` – Delete skill and sub-entities.

2. **App Registration & Verification** (`/api/v1/apps`):
   - `POST /api/v1/apps/register` – Register application with `app_name` and `email`. Issues an `app_id`, secure API key (`sk_live_...`), and verification link.
   - `GET /api/v1/apps/verify?token=...` – Verifies email and activates the application.
   - All skill mutation endpoints (`POST`, `PUT`, `PATCH`, `DELETE`) require header `X-API-Key`.

3. **Gemini Vector Embeddings & Credential Fallback**:
   - Attempts Google Cloud Project credentials (ADC / Vertex AI) first.
   - Falls back to `GEMINI_API_KEY` from `.env`.
   - Fails gracefully to keyword matching if no credentials are configured.

4. **FastMCP Server Sub-App**:
   - Mounted at `/mcp` exposing tools: `search_skills`, `get_skill`, `register_app`, `verify_app`, `register_skill`.

5. **OpenTelemetry GCP Trace**:
   - Integrated OpenTelemetry exporter sending traces to Google Cloud Trace.

---

## Quickstart

### Run via Bazel
```bash
bazel run //apps/skills-service
```

### Run Tests via Bazel
```bash
bazel test //apps/skills-service:test_skills_service
```
