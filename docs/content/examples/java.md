---
title: Java Integration Examples
description: Standalone Java client integration with System Property resolution, Maven build lifecycle plugins, and Java enterprise skill packages.
weight: 30
---

# Java Integration Examples

This section details the Java integration examples in `examples/java/`, covering standalone Java client integration, Maven build lifecycle pre-processing plugins, and Java enterprise skill packages.

---

## 1. Standalone Java Client Example

Located at `examples/java/client`, this example demonstrates how a standalone Java application uses system properties for server discovery, wires `castor-client` into `pom.xml`, and parses skill URIs via `SkillLoader`.

### Key Features & Design
- **Standard Maven Layout**: Features an explicit `pom.xml` requiring no local Bazel toolchain.
- **Maven Lifecycle Integration**: Configures `castor-client:generate-manifest` to run during the `generate-resources` phase, automatically scanning local skill directories and bundling `skills_manifest.json` into `target/classes`.
- **System Property Resolution**: Implements `System.getProperty("castor.server.url")` falling back to environment variables (`CASTOR_SERVER_URL`).

### Project Layout

```text
examples/java/client/
├── BUILD.bazel              # Bazel test rule (test_java_client_example)
├── pom.xml                  # Maven POM declaring castor-client
└── src/
    ├── main/java/com/company/example/
    │   └── Application.java # Main application resolving properties & parsing URIs
    └── test/java/com/company/example/
        └── ApplicationTest.java # JUnit 5 test suite
```

### Maven Plugin Lifecycle Configuration (`pom.xml`)

```xml
<build>
    <plugins>
        <plugin>
            <groupId>com.retailcortex.castor</groupId>
            <artifactId>castor-client</artifactId>
            <version>1.0.0</version>
            <executions>
                <execution>
                    <phase>generate-resources</phase>
                    <goals>
                        <goal>generate-manifest</goal>
                    </goals>
                </execution>
            </executions>
        </plugin>
    </plugins>
</build>
```

### Application Walkthrough (`Application.java`)

```java
package com.company.example;

import com.retailcortex.castor.loader.SkillLoader;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class Application {

    private static final Logger logger = LoggerFactory.getLogger(Application.class);

    public static String getCastorServerUrl() {
        String prop = System.getProperty("castor.server.url", System.getProperty("skm.server.url"));
        if (prop != null && !prop.isBlank()) {
            return prop;
        }
        String env = System.getenv("CASTOR_SERVER_URL");
        if (env == null || env.isBlank()) {
            env = System.getenv("CSTR_SERVER_URL");
        }
        if (env == null || env.isBlank()) {
            env = System.getenv("SKM_SERVER_URL");
        }
        return env != null && !env.isBlank() ? env : "http://localhost:8080";
    }

    public static String getCastorApiKey() {
        String prop = System.getProperty("castor.api.key", System.getProperty("skm.api.key"));
        if (prop != null && !prop.isBlank()) {
            return prop;
        }
        String env = System.getenv("CASTOR_API_KEY");
        if (env == null || env.isBlank()) {
            env = System.getenv("CSTR_API_KEY");
        }
        if (env == null || env.isBlank()) {
            env = System.getenv("SKM_API_KEY");
        }
        return env != null && !env.isBlank() ? env : "java-secret-key-99999";
    }

    public static void main(String[] args) {
        logger.info("Initializing Castor Java Client Standalone Example...");

        String serverUrl = getCastorServerUrl();
        String apiKey = getCastorApiKey();

        logger.info("Loaded System properties: serverUrl={}, apiKey={}", serverUrl, apiKey);

        var parsed = SkillLoader.parseSkillRootUri("castor://skills/example.com/testing/test-skill/1.0.0");
        logger.info("Parsed URI: scheme={}, target={}, ref={}", parsed.scheme(), parsed.target(), parsed.ref());
    }
}
```

### Execution Commands

```bash
# Native Maven Build & Test (triggers generate-resources pre-processor Mojo)
cd examples/java/client
mvn test

# Bazel Workspace Integration Test
bazel test //examples/java/client:test_java_client_example
```

---

## 2. Java Enterprise Skills Package

Located at `examples/java/skills`, this package provides enterprise Java SDLC skills:

| Skill Directory | Skill Name | Category | Description |
| :--- | :--- | :--- | :--- |
| `src/retailcortex_skills_java/skills/java-enterprise` | `java-enterprise` | Java | Javalin lightweight server, Jackson JSON mapping, Weld CDI, and Hibernate ORM. |
| `src/retailcortex_skills_java/skills/java-project-setup` | `java-project-setup` | Java | Standard Maven directory layout (`src/main/java`, `src/test/java`), `<dependencyManagement>`, SpotBugs, and Checkstyle plugins. |
