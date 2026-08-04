---
title: "Java Client & Maven Plugin"
weight: 20
---

# Java Client & Maven Build Integration (`skills-loader-maven-plugin`)

The Java client library (`com.retailcortex.skills:skills-java`) is implemented as both a runtime library and a **native Maven Plugin** (`skills-loader-maven-plugin`). 

By hooking directly into Maven's build lifecycle (`generate-resources`, `compile`), the plugin validates skill dependencies, enforces 5-point SDLC compliance, and packages pre-compiled `skills_manifest.json` resources into target application JARs automatically.

---

## 1. Native Maven Build Integration (`pom.xml`)

Configure the plugin in your application's `pom.xml`:

```xml
<project>
    <modelVersion>4.0.0</modelVersion>
    <groupId>com.company.agent</groupId>
    <artifactId>agent-service</artifactId>
    <version>1.0.0</version>

    <build>
        <plugins>
            <!-- Enterprise Skills Loader Maven Plugin -->
            <plugin>
                <groupId>com.retailcortex.skills</groupId>
                <artifactId>skills-loader-maven-plugin</artifactId>
                <version>1.0.0</version>
                <executions>
                    <execution>
                        <phase>generate-resources</phase>
                        <goals>
                            <goal>generate-manifest</goal>
                        </goals>
                    </execution>
                </executions>
                <configuration>
                    <!-- Qualified skill root URIs to resolve during build -->
                    <roots>
                        <root>skm://skills/sk-9b1deb4d</root>
                        <root>github://google/skills@main/tree/main/skills/cloud/gemini-api</root>
                        <root>file://${project.basedir}/skills</root>
                    </roots>
                    <!-- Target directory for generated resources -->
                    <outputDirectory>${project.build.directory}/generated-resources/skills</outputDirectory>
                    <outputFilename>skills_manifest.json</outputFilename>
                </configuration>
            </plugin>
        </plugins>
    </build>

    <dependencies>
        <!-- Runtime Client Library -->
        <dependency>
            <groupId>com.retailcortex.skills</groupId>
            <artifactId>skills-java</artifactId>
            <version>1.0.0</version>
        </dependency>
    </dependencies>
</project>
```

### What Happens During `mvn compile` or `mvn package`:

1. **Pre-Processing Execution**: The `GenerateManifestMojo` fires during the `generate-resources` lifecycle phase.
2. **Skill Resolution**: Resolves all specified `roots` from central SKM servers (`skm://`), GitHub (`github://`), local filesystem (`file://`), or local Maven artifacts (`~/.m2`).
3. **Build Validation**: Validates SDLC invariants (YAML frontmatter, CWE rules, retry policies). If validation fails, **the Maven build fails**.
4. **Resource Injection**: Generates `skills_manifest.json` and automatically registers `${project.build.directory}/generated-resources/skills` into the Maven project resources so it is bundled directly inside the final target `.jar`.

---

## 2. Hermetic Bazel Integration (`rules_jvm_external`)

In `MODULE.bazel`:

```starlark
maven = use_extension("@rules_jvm_external//:extensions.bzl", "maven")
maven.install(
    artifacts = [
        "com.retailcortex.skills:skills-java:1.0.0",
        "com.fasterxml.jackson.core:jackson-databind:2.17.1",
        "org.slf4j:slf4j-api:2.0.12",
    ],
)
use_repo(maven, "maven")
```

In `BUILD.bazel`:

```starlark
java_library(
    name = "agent_service_java",
    srcs = glob(["src/main/java/**/*.java"]),
    resources = ["//docs:site"],
    deps = [
        "@maven//:com_retailcortex_skills_skills_java",
        "@maven//:org_slf4j_slf4j_api",
    ],
)
```

---

## 3. Runtime Manifest Consumption & Zero-I/O Loading

Because the Maven plugin automatically injects `skills_manifest.json` into classloader resources, your Spring Boot or Javalin runtime loads skills with **zero file I/O**:

```java
package com.company.agent;

import com.retailcortex.skills.loader.SkillLoader;
import com.retailcortex.skills.loader.SkillDefinition;

import java.io.InputStream;
import java.util.Map;

public class AgentApplication {
    public static void main(String[] args) throws Exception {
        // Load pre-compiled manifest bundled directly in JAR classpath
        try (InputStream is = AgentApplication.class.getResourceAsStream("/skills_manifest.json")) {
            if (is != null) {
                Map<String, SkillDefinition> skills = SkillLoader.loadSkillsFromStream(is);
                System.out.printf("Instantly loaded %d skills from classpath manifest.%n", skills.size());
            }
        }
    }
}
```

---

## 4. Unit Testing & HTTP Mocking with Mockito

`SkillLoader` includes a test helper hook `setHttpClient(HttpClient client)` for isolated JUnit 5 unit tests:

```java
import com.retailcortex.skills.loader.SkillLoader;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.mockito.Mockito;

import java.net.http.HttpClient;
import java.net.http.HttpResponse;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;

class SkillLoaderTest {

    private HttpClient mockHttpClient;
    private HttpResponse<String> mockResponse;

    @BeforeEach
    void setUp() {
        mockHttpClient = Mockito.mock(HttpClient.class);
        mockResponse = Mockito.mock(HttpResponse.class);
        SkillLoader.setHttpClient(mockHttpClient);
    }

    @Test
    void testLoadFromSKMServerMocked() throws Exception {
        when(mockResponse.statusCode()).thenReturn(200);
        when(mockResponse.body()).thenReturn("""
            {
              "id": "sk-9b1deb4d",
              "name": "mocked-skill",
              "description": "Mocked skill for testing",
              "version": "1.0.0"
            }
            """);
        when(mockHttpClient.send(any(), any())).thenReturn(mockResponse);

        var skills = SkillLoader.loadSkillsFromSKMServer("sk-9b1deb4d", null, "http://localhost:8080", "key");
        assertThat(skills).containsKey("mocked-skill");
    }
}
```

---

## 5. Integrating Skills with Google ADK Agents

Loaded `SkillDefinition` records map directly to Google ADK system instructions and agent prompt configurations. Server connection properties are loaded via Java System properties (`System.getProperty`) with environment variable fallbacks:

```java
package com.company.agent;

import com.retailcortex.skills.loader.SkillLoader;
import com.retailcortex.skills.loader.SkillDefinition;

import com.google.adk.agent.Agent;

import java.io.InputStream;
import java.util.Map;

public class ADKAgentRunner {
    public static void main(String[] args) throws Exception {
        // 1. Resolve server and authentication System properties (-Dskm.server.url=... -Dskm.api.key=...)
        String serverUrl = System.getProperty("skm.server.url",
                System.getenv().getOrDefault("SKM_SERVER_URL", "http://localhost:8080"));
        String apiKey = System.getProperty("skm.api.key",
                System.getenv().getOrDefault("SKM_API_KEY", ""));

        // 2. Load skill definition from classpath manifest or remote SKM server
        Map<String, SkillDefinition> skills;
        try (InputStream is = ADKAgentRunner.class.getResourceAsStream("/skills_manifest.json")) {
            if (is != null) {
                skills = SkillLoader.loadSkillsFromStream(is);
            } else {
                skills = SkillLoader.loadSkillsFromSKMServer("sk-9b1deb4d", null, serverUrl, apiKey);
            }
        }

        SkillDefinition skill = skills.get("gemini-api");

        // 3. Instantiate Google ADK Agent grounded in skill instructions
        Agent agent = Agent.builder()
                .name(skill.name())
                .model("gemini-2.0-flash")
                .systemInstruction(skill.instructions())
                .build();

        // 4. Execute ADK agent request
        String response = agent.execute("Generate BigQuery analytics statement for retail orders");
        System.out.println("ADK Agent Response:\n" + response);
    }
}
```


---

## Best Practices for Enterprise Java Services

1. **Logging Compliance**: `SkillLoader` strictly utilizes `SLF4J` for all diagnostic logging. Never use `e.printStackTrace()` or stdout for error logging.
2. **Immutable Thread Safety**: `SkillDefinition` objects are immutable Java records, enabling safe multi-threaded sharing across Spring Singletons and Weld CDI beans.
3. **Fail-Fast Build Integrity**: Configure `<failOnError>true</failOnError>` in the plugin to ensure corrupt skill definitions abort CI/CD pipelines before deployment.

