---
name: opentelemetry-google
description: Elite OpenTelemetry & Google Cloud Trace SDLC. Covers InMemorySpanExporter TDD, CI/CD verification, HTTP 429 quota backoff, PII scrubbing security (CWE-200), and batch flush lifecycles.
license: Apache-2.0
metadata:
  author: Ryan McGuinness
  version: "1.0"
---

# OpenTelemetry & Google Cloud Trace SDLC Skill

This skill prescribes best practices for instrumenting enterprise Python applications, FastAPI services, and Google ADK agents with **OpenTelemetry**, exporting traces to **Google Cloud Trace**, and applying strict SDLC standards.

## Prerequisites & Pre-Flight Checklist

1. Google Cloud project with Cloud Trace API enabled.
2. IAM role `roles/cloudtrace.agent` assigned to execution service account.

## HTTP 429 Rate Limit & Trace Quota Invariants

- Outbound trace batch exporters must implement exponential backoff with full randomized jitter to survive Google Cloud Trace HTTP 429 quota exhaustion.

## Security Checkpoints & CWE Invariants

- **CWE-200 (Information Exposure)**: MUST scrub sensitive user data, passwords, and PII from span attributes before exporting traces to Cloud Trace.
- **CWE-400 (Resource Exhaustion)**: Always invoke `provider.force_flush()` or `provider.shutdown()` on application exit to prevent memory leaks and dropped trace spans.

## 3-Phase Execution Protocol

### Phase 1: Initialize TracerProvider & BatchSpanProcessor
Configure `TracerProvider` with `CloudTraceSpanExporter` wrapped in a `BatchSpanProcessor`.

### Phase 2: Implement Telemetry TDD Assertions
Write unit tests using `InMemorySpanExporter` to verify span creation, naming, and attribute injection.

### Phase 3: Execute in CI/CD & Deploy
```bash
uv run pytest tests/test_telemetry.py
```

## Progressive Disclosure References

- **OpenTelemetry GCP Trace Guide**: Read [`references/gcp_trace_otel.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/opentelemetry-google/references/gcp_trace_otel.md).
- **Reference Telemetry Setup**: View [`examples/telemetry.py`](file:///Users/rmcguinness/Projects/skill-builder/skills/opentelemetry-google/examples/telemetry.py).
