---
name: java-enterprise
description: Elite Java 25 (LTS) microservice SDLC. Covers Google OAuth2 JWT verification, Javalin 6+, HTTP 429 rate limiting (Resilience4j), SLF4J 2.x, and JUnit 5 TDD with 85% JaCoCo coverage.
license: Apache-2.0
metadata:
  author: Ryan McGuinness
  version: "1.0"
authors:
  - name: Retail Cortex Engineering
    url: https://github.com/retail-cortex/skills
category: java
tags:
  - java
  - javalin
  - maven
  - enterprise
trigger_phrases:
  - "build Java enterprise service"
  - "Javalin REST setup"
  - "Java JUnit 5 TDD"
execution_hints:
  preferred_model: "gemini-3.1-pro"
  requires_human_approval: false
  environment_variables:
    - JAVA_HOME
  timeout_seconds: 240
---
# Java 25 (LTS) Enterprise Microservices SDLC Skill

This skill prescribes best practices for designing, testing, securing, and deploying modern **Java 25 (LTS)** enterprise microservices using **Javalin 6+**, **Google OAuth2 Authentication**, **Resilience4j HTTP 429 Rate Limiting**, and **Bazel** (`rules_java`).

## Prerequisites & Pre-Flight Checklist

1. JDK 25 (LTS) runtime installed (Eclipse Temurin 25).
2. Google Cloud OAuth2 Web Client ID configured.
3. Maven 3.9+ and Bazelisk on system PATH.

## HTTP 429 Rate Limiting & Resilience4j Invariants

1. **Bucket4j Inbound Filter**:
   - Throttle requests and return HTTP 429 status codes with `Retry-After` headers.
2. **Resilience4j Outbound Retries & Quota Backoff**:
   - Wrap outbound GCP and AI calls in `Resilience4j` Retry and Circuit Breaker modules configured with exponential backoff and randomized jitter to handle 429 rate limits.
3. **429 Negative JUnit 5 TDD Assertions**:
   - Write `@ParameterizedTest` methods simulating external 429 responses.

## Security Checkpoints & CWE Invariants

- **CWE-532 (Sensitive Info in Log Files)**: Standardize exclusively on SLF4J 2.x; NEVER use `printStackTrace()` or `System.out.println()`.
- **CWE-502 (Deserialization of Untrusted Data)**: Configure Jackson with explicit typing and disable default polymorphism on unverified payloads.
- **CWE-250 (Execution with Unnecessary Privileges)**: Package shaded fat JARs inside minimal JRE alpine images executing as non-root.

## Defensive Error Handling & Null Pointer Safety Invariants

- **Strict Null Pointer Prevention**: Methods returning absent values MUST return `java.util.Optional<T>`. Annotate method parameters with `@NonNull` or `@Nullable`.
- **Defensive Domain Exceptions**: Create typed domain runtime exceptions and register global Javalin exception handlers.

## 3-Phase Execution Protocol

### Phase 1: Scaffold Maven POM with Google Cloud BOM & Auth SDK
Configure Java 25 compiler release (`<maven.compiler.release>25</maven.compiler.release>`) and import the latest Google Cloud BOM.

### Phase 2: Implement Positive & Negative TDD Suite (85% Coverage)
Write parameterized tests in JUnit 5 (Jupiter) mocking Google Auth verifiers and evaluating happy path, 429 rate limit backoff, and exception propagation.

### Phase 3: Verify with JaCoCo, SpotBugs, Checkstyle & Bazel
```bash
mvn clean verify
bazel build //...
bazel test //...
```

## Progressive Disclosure References

- **Javalin & Maven Architecture**: Read [`references/javalin_maven_patterns.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/java-enterprise/references/javalin_maven_patterns.md).
- **Reference POM**: View [`examples/pom.xml`](file:///Users/rmcguinness/Projects/skill-builder/skills/java-enterprise/examples/pom.xml).
