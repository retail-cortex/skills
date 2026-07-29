package com.retailcortex.skills.loader;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.util.List;
import java.util.Map;

import static org.assertj.core.api.Assertions.assertThat;

@DisplayName("SkillLoader Core Engine Unit Tests")
class SkillLoaderTest {

    @Test
    @DisplayName("Should find valid workspace registry root")
    void testFindRegistryRoot() {
        Path root = SkillLoader.findRegistryRoot();
        assertThat(root).isNotNull();
        assertThat(Files.exists(root)).isTrue();
        assertThat(Files.isDirectory(root.resolve("packages")) || Files.isDirectory(root.resolve("skills"))).isTrue();
    }

    @Test
    @DisplayName("Should correctly parse frontmatter and body content")
    void testParseFrontmatterValid() {
        String sample = """
                ---
                name: test-skill
                description: Test desc
                ---
                # Rules
                """;
        SkillLoader.FrontmatterResult result = SkillLoader.parseFrontmatter(sample);
        assertThat(result.data()).containsEntry("name", "test-skill");
        assertThat(result.data()).containsEntry("description", "Test desc");
        assertThat(result.body().trim()).isEqualTo("# Rules");
    }

    @Test
    @DisplayName("Should handle content without frontmatter")
    void testParseFrontmatterEmpty() {
        String sample = "# Header\nNo frontmatter";
        SkillLoader.FrontmatterResult result = SkillLoader.parseFrontmatter(sample);
        assertThat(result.data()).isEmpty();
        assertThat(result.body()).isEqualTo(sample);
    }


    @ParameterizedTest
    @DisplayName("Should correctly parse skill root URIs")
    @CsvSource(delimiter = '|', textBlock = """
        file://skills/custom | file | skills/custom | |
        pkg://retailcortex_skills_python | pkg | retailcortex_skills_python | |
        github://google/skills/skills/cloud/gemini-api:main | github | google/skills | main | skills/cloud/gemini-api
        github://owner/repo:v2.5.0 | github | owner/repo | v2.5.0 |
        github://owner/repo@v1.2.0/skills/cloud | github | owner/repo | v1.2.0 | skills/cloud
        github://google/skills/tree/main/skills/cloud/gemini-api | github | google/skills | main | skills/cloud/gemini-api
        """)
    void testParseSkillRootUri(String uri, String expScheme, String expTarget, String expRef, String expSubpath) {
        SkillLoader.ParsedUri parsed = SkillLoader.parseSkillRootUri(uri);
        assertThat(parsed.scheme()).isEqualTo(expScheme);
        assertThat(parsed.target()).isEqualTo(expTarget);
        if (expRef != null && !expRef.isEmpty()) {
            assertThat(parsed.ref()).isEqualTo(expRef);
        }
        if (expSubpath != null && !expSubpath.isEmpty()) {
            assertThat(parsed.subpath()).isEqualTo(expSubpath);
        }
    }

    @Test
    @DisplayName("Should parse dotenv key-value environment file")
    void testParseDotenvFile(@TempDir Path tempDir) throws IOException {
        Path envPath = tempDir.resolve(".env");
        Files.writeString(envPath, "GITHUB_TOKEN=ghp_secret_123\nGITHUB_REF=v2.1.0\nSKILLS_ROOTS=skills,custom\n# comment\n");

        Map<String, String> vars = SkillLoader.parseDotenvFile(envPath);
        assertThat(vars).containsEntry("GITHUB_TOKEN", "ghp_secret_123");
        assertThat(vars).containsEntry("GITHUB_REF", "v2.1.0");
        assertThat(vars).containsEntry("SKILLS_ROOTS", "skills,custom");
    }

    @Test
    @DisplayName("Should load all skills from workspace")
    void testLoadAllSkills() {
        Map<String, SkillDefinition> skills = SkillLoader.loadAllSkills(null, null);
        assertThat(skills).hasSizeGreaterThanOrEqualTo(20);
        assertThat(skills).containsKey("python-core");
        assertThat(skills).containsKey("go-lang");
    }

    @Test
    @DisplayName("Should query SkillRegistry for skills, search, and domain filtering")
    void testSkillRegistryQueries() {
        SkillRegistry registry = new SkillRegistry();
        assertThat(registry.getSkills()).hasSizeGreaterThanOrEqualTo(20);

        SkillDefinition pythonSkill = registry.get("python-core");
        assertThat(pythonSkill).isNotNull();
        assertThat(pythonSkill.getName()).isEqualTo("python-core");
        assertThat(pythonSkill.getInstructions()).isNotEmpty();

        List<SkillDefinition> goMatches = registry.search("Go");
        assertThat(goMatches).isNotEmpty();

        List<SkillDefinition> domainMatches = registry.getDomainSkills("fastapi");
        assertThat(domainMatches).isNotEmpty();

        List<SkillSummary> summaries = registry.listSkills();
        assertThat(summaries).hasSizeGreaterThanOrEqualTo(20);
    }

    @Test
    @DisplayName("Should parse allowed-tools, metadata, and provide content getters")
    void testFrontmatterMetadataMapping(@TempDir Path tempDir) throws IOException {
        Path skillDir = tempDir.resolve("test-meta-skill");
        Files.createDirectories(skillDir.resolve("references"));
        Files.writeString(skillDir.resolve("references/ref1.md"), "Reference 1 Content");

        String content = """
                ---
                name: test-meta-skill
                description: Meta test
                license: MIT
                author: Jane Doe
                version: 2.0.0
                compatibility: ADK-v2
                allowed-tools: Bash(git:*) Read
                custom_field: custom_val
                ---
                # Instructions
                """;
        Files.writeString(skillDir.resolve("SKILL.md"), content);

        SkillDefinition skill = SkillLoader.loadSkillFromDir(skillDir);
        assertThat(skill).isNotNull();
        assertThat(skill.getLicense()).isEqualTo("MIT");
        assertThat(skill.getAuthor()).isEqualTo("Jane Doe");
        assertThat(skill.getVersion()).isEqualTo("2.0.0");
        assertThat(skill.getCompatibility()).isEqualTo("ADK-v2");
        assertThat(skill.getAllowedTools()).isEqualTo("Bash(git:*) Read");
        assertThat(skill.getMetadata()).containsEntry("custom_field", "custom_val");
        assertThat(skill.getMetadata()).containsEntry("author", "Jane Doe");
        assertThat(skill.getReferenceContent("ref1.md")).isEqualTo("Reference 1 Content");
        assertThat(skill.getReferenceContent("nonexistent.md")).isNull();
    }

    @Test
    @DisplayName("Should discover skills in .agents/skills directory")
    void testAgentsSkillsDirectoryScanning(@TempDir Path tempDir) throws IOException {
        Path agDir = tempDir.resolve(".agents").resolve("skills").resolve("ag-java-skill");
        Files.createDirectories(agDir);
        String content = """
                ---
                name: ag-java-skill
                description: Cross client agent skill for Java
                ---
                # Instructions
                """;
        Files.writeString(agDir.resolve("SKILL.md"), content);

        Map<String, SkillDefinition> skills = SkillLoader.loadAllSkills(tempDir, null);
        assertThat(skills).containsKey("ag-java-skill");
        assertThat(skills.get("ag-java-skill").getDescription()).isEqualTo("Cross client agent skill for Java");
    }

    @Test
    @DisplayName("Should generate and load pre-compiled skills manifest JSON")
    void testBuildAndLoadManifest(@TempDir Path tempDir) throws IOException {
        Path manifestPath = tempDir.resolve("skills_manifest.json");
        Path generated = SkillLoader.buildSkillsManifest(null, manifestPath);
        assertThat(Files.isRegularFile(generated)).isTrue();

        Map<String, SkillDefinition> loaded = SkillLoader.loadSkillsFromManifest(generated);
        assertThat(loaded).hasSizeGreaterThanOrEqualTo(20);
        assertThat(loaded).containsKey("python-core");
    }

    @Test
    @DisplayName("Should load skills from package")
    void testLoadSkillsFromPackage() {
        Map<String, SkillDefinition> skills = SkillLoader.loadSkillsFromPackage("retailcortex_skills_python", List.of("python-core"));
        assertThat(skills).containsKey("python-core");
    }

    @Test
    @DisplayName("Should load skills from roots with file and pkg schemes")
    void testLoadSkillsFromRoots() {
        Map<String, SkillDefinition> skills = SkillLoader.loadSkillsFromRoots(List.of("file://.", "pkg://retailcortex_skills_python"), List.of("python-core"), null, null);
        assertThat(skills).containsKey("python-core");
    }

    @Test
    @DisplayName("Should test SkillSummary records and getters")
    void testSkillSummary() {
        SkillSummary summary = new SkillSummary("name", "desc", 1, 2, "/path");
        assertThat(summary.getName()).isEqualTo("name");
        assertThat(summary.getDescription()).isEqualTo("desc");
        assertThat(summary.getReferenceCount()).isEqualTo(1);
        assertThat(summary.getExampleCount()).isEqualTo(2);
        assertThat(summary.getPath()).isEqualTo("/path");

        summary.setName("name2");
        summary.setDescription("desc2");
        summary.setReferenceCount(3);
        summary.setExampleCount(4);
        summary.setPath("/path2");

        assertThat(summary.getName()).isEqualTo("name2");
        assertThat(summary.getDescription()).isEqualTo("desc2");
        assertThat(summary.getReferenceCount()).isEqualTo(3);
        assertThat(summary.getExampleCount()).isEqualTo(4);
        assertThat(summary.getPath()).isEqualTo("/path2");
    }
}
