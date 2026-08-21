---
name: protobuf-grpc
description: Elite Multi-language Protocol Buffers & gRPC SDLC. Covers in-process gRPC TDD, GitHub Actions Buf breaking change detection, HTTP 429 / ResourceExhausted backoff, and mTLS/auth security (CWE-306).
license: Apache-2.0
metadata:
  author: Ryan McGuinness
  version: "1.0"
authors:
  - name: Retail Cortex Engineering
    url: https://github.com/retail-cortex/castor
category: protobuf
tags:
  - protobuf
  - grpc
  - schema
  - bazel
trigger_phrases:
  - "define Protobuf gRPC service"
  - "Buf breaking change check"
  - "gRPC proto3 schema"
execution_hints:
  preferred_model: "gemini-3.1-pro"
  requires_human_approval: false
  environment_variables: []
  timeout_seconds: 240
---
# Protocol Buffers & gRPC SDLC Skill

This skill prescribes best practices for designing schema-first API contracts with **Protocol Buffers**, generating strongly typed gRPC stubs across Go, Python, TypeScript, and Java, and enforcing complete SDLC verification.

## Prerequisites & Pre-Flight Checklist

1. `protoc` compiler or `buf` CLI installed.
2. Bazelisk installed for hermetic compilation via `rules_proto_grpc`.

## HTTP 429 & gRPC ResourceExhausted Resilience

- gRPC client stubs across all languages MUST attach retry interceptors configured with exponential backoff and randomized jitter to handle `codes.ResourceExhausted` (HTTP 429 equivalent).

## Security Checkpoints & CWE Invariants

- **CWE-306 (Missing Authentication for Critical Function)**: Enforce **mTLS (Mutual TLS)** and authorization interceptors for all inter-service gRPC communication.
- **Field Stability Invariant**: Field numbers are immutable; once assigned, they must never be altered or reused.

## 3-Phase Execution Protocol

### Phase 1: Define Proto Schema & Ruleset
Write proto3 schema definitions in `api/proto/` and configure Bazel `rules_proto_grpc` targets.

### Phase 2: In-Process gRPC TDD Testing
Test service handlers using in-process gRPC test channels (`grpc/test/bufconn` in Go, in-process servers in Java/Python).

### Phase 3: Breaking Change Audit & Build
```bash
buf breaking --against '.git#branch=main'
bazel test //api/...
```

## Progressive Disclosure References

- **Proto & gRPC Compilation Guide**: Read [`references/proto_grpc_rules.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/protobuf-grpc/references/proto_grpc_rules.md).
- **Reference Proto Schema**: View [`examples/service.proto`](file:///Users/rmcguinness/Projects/skill-builder/skills/protobuf-grpc/examples/service.proto).
