---
name: docker-containers
description: Elite multi-stage Docker containerization SDLC. Covers container health check TDD, GitHub Actions CI with Trivy scanning (CWE-1395), HTTP 429 registry rate backoff, and non-root security (CWE-250).
---

# Docker Multi-Stage Containerization SDLC Skill

This skill prescribes best practices for creating secure, minimal, reproducible, and verifiable **Docker container images** for enterprise microservices and AI agent backends.

## Prerequisites & Pre-Flight Checklist

1. Docker Engine 26+ installed.
2. Trivy container vulnerability scanner installed.
3. Access to Google Artifact Registry.

## HTTP 429 Rate Limit & Registry Backoff Invariants

- Configure Docker daemon and CI build-push actions with exponential backoff retries to survive container registry HTTP 429 rate limit spikes during bulk monorepo image pushes.

## Security Checkpoints & CWE Invariants

- **CWE-250 (Execution with Unnecessary Privileges)**: Container runtimes MUST execute as non-privileged users (`USER 65534:65534` or `nonroot:nonroot`).
- **CWE-1395 (Vulnerable Third-Party Components)**: Fail CI pipelines on any CRITICAL or HIGH severity CVEs discovered by Trivy scanning.
- **Minimal Base Images**: Use `scratch` or `distroless` for Go binaries, and slim distributions (`python:3.13-slim`, `alpine`) for Python/Java to minimize CVE exposure.

## 3-Phase Execution Protocol

### Phase 1: Multi-Stage Dockerfile Definition
Separate build environment (SDKs, compilers) from final runtime image.

### Phase 2: Container TDD & Trivy Security Scan
```bash
docker build -t service:v1.0.0 .
trivy image --severity HIGH,CRITICAL service:v1.0.0
```

### Phase 3: Tagged Push to Artifact Registry
```bash
docker tag service:v1.0.0 gcr.io/my-gcp-project/service:v1.0.0
docker push gcr.io/my-gcp-project/service:v1.0.0
```

## Progressive Disclosure References

- **Dockerfile Standards Guide**: Read [`references/dockerfile_standards.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/docker-containers/references/dockerfile_standards.md).
- **Reference Dockerfile**: View [`examples/Dockerfile`](file:///Users/rmcguinness/Projects/skill-builder/skills/docker-containers/examples/Dockerfile).
