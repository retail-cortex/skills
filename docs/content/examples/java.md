---
title: Java Integration Examples
description: Standalone Java client integration with System Property resolution, Maven build lifecycle plugins, and Java enterprise skill packages.
weight: 30
---

# Java Integration Examples

This section details the Java integration examples in `examples/java/`, covering standalone Java client integration, Maven build lifecycle pre-processing plugins, and Java enterprise skill packages.

---

## 1. Standalone Java Client Example

Located at `examples/java/client`, this example demonstrates how a standalone Java application uses system properties for server discovery, wires `skills-loader-maven-plugin` into `pom.xml`, and loads pre-compiled `skills_manifest.json` resources.

### Key Features & Design
- **Standard Maven Layout**: Features an explicit `pom.xml` requiring no local Bazel toolchain.
- **Maven Lifecycle Integration**: Configures `skills-loader-maven-plugin:generate-manifest` to run during the `generate-resources` phase, automatically scanning local skill directories and bundling `skills_manifest.json` into `target/classes`.
- **System Property Resolution**: Implements `System.getProperty("castor.server.url")` falling back to environment variables (`CASTOR_SERVER_URL`).

### Project Layout

```text
examples/java/client/
├── BUILD.bazel              # Bazel test rule (test_java_client_example)
├── pom.xml                  # Maven POM declaring skills-loader-maven-plugin
└── src/
    ├── main/java/com/company/example/
    │   └── Application.java # Main application resolving properties & loading skills
    └── test/java/com/company/example/
        └── ApplicationTest.java # JUnit 5 test suite
```

### Maven Plugin Lifecycle Configuration (`pom.xml`)

```xml
<build>
    <plugins>
        <plugin>
            <groupId>com.retailcortex.castor</groupId>
            <artifactId>skills-loader-maven-plugin</artifactId>
            <version>1.0.0-SNAPSHOT</version>
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

import com.retailcortex.castor.loader.SkillDefinition;
import com.retailcortex.castor.loader.SkillLoader;
import java.util.List;
import java.util.Map;

public class Application {

    public String resolveServerUrl() {
        String prop = System.getProperty("castor.server.url");
        if (prop != null && !prop.isBlank()) {
            return prop;
        }
        String env = System.getenv("CASTOR_SERVER_URL");
        return (env != null && !env.isBlank()) ? env : "http://localhost:8080";
    }

    public static void main(String[] args) {
        Application app = new Application();
        System.out.println("Connected to Castor Server at: " + app.resolveServerUrl());

        // Load Java Enterprise Skills
        Map<String, SkillDefinition> skills = SkillLoader.loadSkillsFromPackage(
            "retailcortex_skills_java", List.of("java-enterprise")
        );

        for (Map.Entry<String, SkillDefinition> entry : skills.entrySet()) {
            System.out.println("Loaded Java Skill [" + entry.getKey() + "]: " + entry.getValue().getDescription());
        }
    }
}
```

### Execution Commands

```bash
# Native Maven Build & Test (triggers generate-resources pre-processor Mojo)
cd examples/java/client
mvn test

# Bazel Workspace Integration Test
bazel test //examples/java/client:src/test/java/com/company/example/ApplicationTest
```

---

## 2. Java Enterprise Skills Package

Located at `examples/java/skills`, this package provides enterprise Java SDLC skills:

| Skill Directory | Skill Name | Category | Description |
| :--- | :--- | :--- | :--- |
| `src/retailcortex_skills_java/skills/java-enterprise` | `java-enterprise` | Java | Javalin lightweight server, Jackson JSON mapping, Weld CDI, and Hibernate ORM. |
| `src/retailcortex_skills_java/skills/java-project-setup` | `java-project-setup` | Java | Standard Maven directory layout (`src/main/java`, `src/test/java`), `<dependencyManagement>`, SpotBugs, and Checkstyle plugins. |
