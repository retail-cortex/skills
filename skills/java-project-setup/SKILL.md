---
name: java-project-setup
description: Elite meta-skill for scaffolding enterprise Java 25 (LTS) microservices using Maven POM with Google Cloud BOM, wrapped in Bazel rules_java. Enforces JUnit 5 TDD, HTTP 429 rate limit retries, and GCP Terraform.
---

# Java 25 (LTS) Project Setup Meta-Skill (Maven & Bazel Standard)

This meta-skill provides automated scaffolding instructions and templates for initializing enterprise **Java 25 (LTS)** microservices using **Maven** (`pom.xml`), **Bazel** (`rules_java`), **Javalin 6+**, and a dedicated `configs/` directory.

## Prerequisites & Pre-Flight Checklist

1. JDK 25 (LTS) installed locally (e.g. Eclipse Temurin 25).
2. Maven 3.9+ and Bazelisk available on system PATH.
3. Access to GCP environment for Cloud Run / GKE deployment.

## HTTP 429 Rate Limit & Quota Resilience Invariants

- Outbound API calls to Google Cloud and AI backends must be wrapped in `Resilience4j` retry and circuit breakers to handle HTTP 429 quota exhaustion.

## Security Checkpoints & CWE Invariants

- **CWE-532 (Sensitive Info in Log Files)**: Standardize exclusively on SLF4J 2.x; strictly prohibit `printStackTrace()` or logging plain-text passwords.
- **CWE-502 (Deserialization of Untrusted Data)**: Configure Jackson with explicit typing and disable default polymorphism on unverified payloads.
- **CWE-250 (Execution with Unnecessary Privileges)**: Package shaded JARs inside minimal JRE alpine images executing as non-root.

## 3-Phase Execution Protocol

### Phase 1: Initialize Maven & Bazel Directory Layout
```bash
mkdir -p src/main/java/com/enterprise/service/controllers src/main/resources
mkdir -p src/test/java/com/enterprise/service src/test/resources
mkdir -p configs/terraform .github/workflows
```

### Phase 2: Configure Maven POM with GCP BOM & Bazel
Populate `pom.xml` with Java 25 compiler release (`<maven.compiler.release>25</maven.compiler.release>`) and Google Cloud BOM. Wrap in Bazel `rules_java`.

### Phase 3: Run TDD Suite (85% JaCoCo Coverage) & SpotBugs
```bash
mvn clean verify
bazel test //...
```

## Progressive Disclosure References

- **Java Scaffold Guide**: Read [`references/java_scaffold_guide.md`](file:///Users/rmcguinness/Projects/skill-builder/skills/java-project-setup/references/java_scaffold_guide.md).
- **Reference POM**: View [`examples/pom.xml`](file:///Users/rmcguinness/Projects/skill-builder/skills/java-project-setup/examples/pom.xml).
- **Reference Bazel Build**: View [`examples/BUILD.bazel`](file:///Users/rmcguinness/Projects/skill-builder/skills/java-project-setup/examples/BUILD.bazel).
- **Reference Terraform**: View [`examples/configs/terraform/main.tf`](file:///Users/rmcguinness/Projects/skill-builder/skills/java-project-setup/examples/configs/terraform/main.tf).
- **Reference Base Config**: View [`examples/configs/.env.toml`](file:///Users/rmcguinness/Projects/skill-builder/skills/java-project-setup/examples/configs/.env.toml).
