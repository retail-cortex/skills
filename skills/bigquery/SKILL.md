---
name: bigquery
description: Elite Google Cloud BigQuery & Conversational Analytics (CAPI) SDLC. Covers dry-run TDD, HTTP 429 rate limit backoff, GE OAuth security delegation (CWE-269), Protobuf normalization, and troubleshooting.
---

# Google Cloud BigQuery & CAPI SDLC Skill

This skill prescribes best practices for querying enterprise data warehouses using the **BigQuery Python SDK** (`google-cloud-bigquery`), integrating **BigQuery Conversational Analytics (CAPI)** via `geminidataanalytics`, and enforcing complete SDLC rigor.

## Prerequisites & Pre-Flight Checklist

1. Google Cloud project with BigQuery API and Gemini Data Analytics API enabled.
2. IAM roles `roles/bigquery.user` and `roles/bigquery.dataViewer` assigned.
3. Verify that all SQL queries reference **fully qualified table names**: `<gcp-project>.<dataset>.<table-name>`.

## HTTP 429 Rate Limit & Quota Resilience Invariants

- BigQuery and CAPI natural language query jobs must use `tenacity` exponential backoff with full randomized jitter to survive HTTP 429 / `RESOURCE_EXHAUSTED` query concurrency quota spikes.

## Security Checkpoints & CWE Invariants

- **CWE-269 (Improper Privilege Management)**: CAPI queries MUST execute as the end user by delegating Gemini Enterprise OAuth tokens from session state (`resolve_user_credentials`).
- **CWE-89 (SQL Injection)**: SQL queries MUST always reference fully qualified notation: `` `<gcp-project>.<dataset>.<table-name>` `` and use query parameter bindings.

## 3-Phase Execution Protocol

### Phase 1: Dry-Run Query TDD Verification
Test SQL query syntax and estimate bytes processed without incurring cost (`query_job_config.dry_run = True`).

### Phase 2: Conversational Analytics (CAPI) Execution
Run natural language CAPI queries using the global endpoint invariant (`parent="projects/{p}/locations/global"`).

### Phase 3: Protobuf Normalization & CI/CD Linting
```bash
uv run sqlfluff lint queries/
uv run pytest tests/test_bigquery.py
```

## Troubleshooting & Remediation Matrix

| Symptom / Error | Root Cause | Exact Remediation |
| :--- | :--- | :--- |
| `404 Not Found: Location 'us-central1' not found` | Regional endpoint used for CAPI client | CAPI requires `parent="projects/{project}/locations/global"`. Set location to `global`. |
| `SerializationError: Protobuf object is not JSON serializable` | CAPI stream contains nested protobuf structs | Pass response through `_to_plain()` recursive dictionary converter before serialization. |
| `Table not found: dataset.table` | Missing GCP project ID in SQL query | Prefix with project ID: `` `<gcp-project>.dataset.table` ``. |
| `403 Access Denied: User does not have bigquery.jobs.create` | Tool using service account credentials instead of user OAuth | Retrieve end-user token from `tool_context.state` and instantiate credentials via `google.oauth2.credentials.Credentials`. |

## Progressive Disclosure References

- **BigQuery & CAPI Architecture**: Read [`references/bigquery_capi.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/bigquery/references/bigquery_capi.md).
- **Reference BigQuery Client**: View [`examples/bq_client.py`](file:///Users/rmcguinness/Projects/skill-builder/skills/bigquery/examples/bq_client.py).
