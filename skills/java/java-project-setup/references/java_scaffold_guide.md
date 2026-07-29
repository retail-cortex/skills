# Java 25 (LTS) Enterprise Scaffolding & Configuration Architecture

## Directory Scaffolding Script

```bash
#!/usr/bin/env bash
set -euo pipefail

# 1. Directory Tree
mkdir -p src/main/java/com/enterprise/service/controllers
mkdir -p src/main/resources configs/terraform .github/workflows

# 2. Touch baseline files
touch src/main/java/com/enterprise/service/Application.java
touch src/main/resources/logback.xml
touch configs/.env.toml configs/.env.local.toml
touch configs/terraform/main.tf
touch pom.xml BUILD.bazel
```

## Maven Java 25 BOM Invariant (`pom.xml`)

Always configure compiler release `25` and import the latest Google Cloud `libraries-bom`:

```xml
<properties>
    <maven.compiler.release>25</maven.compiler.release>
    <project.build.sourceEncoding>UTF-8</project.build.sourceEncoding>
</properties>
```
