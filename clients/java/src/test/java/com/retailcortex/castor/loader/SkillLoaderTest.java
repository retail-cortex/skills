// Copyright 2026 Ryan McGuinness
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package com.retailcortex.castor.loader;

import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;
import org.junit.jupiter.params.ParameterizedTest;
import org.junit.jupiter.params.provider.CsvSource;

import org.mockito.Mockito;
import org.mockito.ArgumentMatchers;

import java.io.FileOutputStream;
import java.io.IOException;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.nio.file.StandardCopyOption;
import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.stream.Stream;
import java.util.zip.ZipEntry;
import java.util.zip.ZipOutputStream;

import static org.assertj.core.api.Assertions.assertThat;

@DisplayName("SkillLoader Core Engine Unit Tests")
class SkillLoaderTest {

    @Test
    @DisplayName("Should find valid workspace registry root")
    void testFindRegistryRoot() {
        Path root = SkillLoader.findRegistryRoot();
        assertThat(root).isNotNull();
        assertThat(Files.exists(root)).isTrue();
        assertThat(Files.isDirectory(root.resolve("clients")) || Files.isDirectory(root.resolve("cmd")) || Files.isDirectory(root.resolve("pkg")) || Files.isDirectory(root.resolve("skills")) || Files.isDirectory(root.resolve("examples"))).isTrue();
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
        castor://skills/example.com/testing/test-skill/1.0.0 | castor | example.com/testing/test-skill/1.0.0 | |
        cstr://skills/example.com/testing/test-skill | cstr | example.com/testing/test-skill | |
        skm://skills/example.com/testing/test-skill | skm | example.com/testing/test-skill | |
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

        List<SkillDefinition> suggested = registry.suggestSkills("python testing framework", 3, null);
        assertThat(suggested).isNotEmpty().hasSizeLessThanOrEqualTo(3);

        List<SkillDefinition> emptySuggested = registry.suggestSkills("", 2, null);
        assertThat(emptySuggested).isNotEmpty().hasSizeLessThanOrEqualTo(2);
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

    @Test
    @DisplayName("Should create .manifest.lock and verify skill integrity")
    void testManifestLockAndVerification(@TempDir Path tempDir) throws IOException {
        Path skillDir = tempDir.resolve("test-lock-skill");
        Files.createDirectories(skillDir.resolve("references"));
        Files.writeString(skillDir.resolve("SKILL.md"), "---\nname: test-lock-skill\n---\nHello World");
        Files.writeString(skillDir.resolve("references/guide.md"), "# Guide");

        String checksum = SkillLoader.calculateSkillChecksum(skillDir);
        assertThat(checksum).isNotEmpty();

        String uri = "file://" + skillDir;
        SkillLoader.updateManifestLock(tempDir, "test-lock-skill", uri, checksum);

        SkillLoader.VerificationReport report = SkillLoader.verifyManifestLock(tempDir);
        assertThat(report.totalSkills()).isEqualTo(1);
        assertThat(report.verifiedCount()).isEqualTo(1);
        assertThat(report.modifiedCount()).isEqualTo(0);
        assertThat(report.missingCount()).isEqualTo(0);

        // Modify skill file
        Files.writeString(skillDir.resolve("SKILL.md"), "---\nname: test-lock-skill\n---\nModified Content");
        SkillLoader.VerificationReport reportMod = SkillLoader.verifyManifestLock(tempDir);
        assertThat(reportMod.totalSkills()).isEqualTo(1);
        assertThat(reportMod.verifiedCount()).isEqualTo(0);
        assertThat(reportMod.modifiedCount()).isEqualTo(1);
        assertThat(reportMod.results().get(0).status()).isEqualTo("modified");

        // Delete skill directory
        SkillLoaderTestUtils.deleteDir(skillDir);
        SkillLoader.VerificationReport reportMiss = SkillLoader.verifyManifestLock(tempDir);
        assertThat(reportMiss.totalSkills()).isEqualTo(1);
        assertThat(reportMiss.missingCount()).isEqualTo(1);
        assertThat(reportMiss.results().get(0).status()).isEqualTo("missing");
    }

    @Test
    @DisplayName("Should mock HTTP calls with Mockito to test loading skills from remote GitHub repository")
    void testLoadSkillsFromGithubWithMockito(@TempDir Path tempDir) throws Exception {
        Path mockZip = tempDir.resolve("mock_repo.zip");
        createMockSkillZip(mockZip, "test-remote-skill", "A mocked remote skill from GitHub");

        HttpClient mockClient = Mockito.mock(HttpClient.class);
        @SuppressWarnings("unchecked")
        HttpResponse<Path> mockResponse = (HttpResponse<Path>) Mockito.mock(HttpResponse.class);
        Mockito.when(mockResponse.statusCode()).thenReturn(200);

        Mockito.doAnswer(inv -> {
            Path tmpRoot = Paths.get(System.getProperty("java.io.tmpdir"));
            try (Stream<Path> s = Files.list(tmpRoot)) {
                s.filter(p -> p.getFileName().toString().startsWith("skills-loader-gh-"))
                 .max(Comparator.comparingLong(p -> p.toFile().lastModified()))
                 .ifPresent(p -> {
                     try {
                         Files.copy(mockZip, p.resolve("repo.zip"), StandardCopyOption.REPLACE_EXISTING);
                     } catch (IOException e) {
                         throw new RuntimeException(e);
                     }
                 });
            }
            return mockResponse;
        }).when(mockClient).send(ArgumentMatchers.any(HttpRequest.class), ArgumentMatchers.any());

        SkillLoader.setHttpClient(mockClient);
        try {
            Map<String, SkillDefinition> loaded = SkillLoader.loadSkillsFromGithub("owner/repo", "main", List.of("skills", "."), null, "secret-token", null);
            assertThat(loaded).containsKey("test-remote-skill");

            Map<String, SkillDefinition> filtered = SkillLoader.loadSkillsFromGithub("https://github.com/owner/repo.git", "v1.0.0", List.of("."), List.of("test-remote-skill"), null, null);
            assertThat(filtered).containsKey("test-remote-skill");

            Map<String, SkillDefinition> fromRoots = SkillLoader.loadSkillsFromRoots(List.of("github://owner/repo@main/skills"), null, "token", null);
            assertThat(fromRoots).containsKey("test-remote-skill");
        } finally {
            SkillLoader.setHttpClient(null);
        }
    }

    @Test
    @DisplayName("Should handle GitHub HTTP error gracefully")
    void testLoadSkillsFromGithubHttpError() throws Exception {
        HttpClient mockClient = Mockito.mock(HttpClient.class);
        HttpResponse<Path> mockResponse = (HttpResponse<Path>) Mockito.mock(HttpResponse.class);
        Mockito.when(mockResponse.statusCode()).thenReturn(404);
        Mockito.doReturn(mockResponse).when(mockClient).send(ArgumentMatchers.any(HttpRequest.class), ArgumentMatchers.any());

        SkillLoader.setHttpClient(mockClient);
        try {
            Map<String, SkillDefinition> loaded = SkillLoader.loadSkillsFromGithub("owner/nonexistent", "main", List.of("."), null, null, null);
            assertThat(loaded).isEmpty();
        } finally {
            SkillLoader.setHttpClient(null);
        }
    }

    @Test
    @DisplayName("Should handle GitHub network exception gracefully")
    void testLoadSkillsFromGithubNetworkException() throws Exception {
        HttpClient mockClient = Mockito.mock(HttpClient.class);
        Mockito.doThrow(new IOException("Simulated network timeout")).when(mockClient).send(ArgumentMatchers.any(HttpRequest.class), ArgumentMatchers.any());

        SkillLoader.setHttpClient(mockClient);
        try {
            Map<String, SkillDefinition> loaded = SkillLoader.loadSkillsFromGithub("owner/repo", "main", List.of("."), null, null, null);
            assertThat(loaded).isEmpty();
        } finally {
            SkillLoader.setHttpClient(null);
        }
    }

    @Test
    @DisplayName("Should load skills from Maven local m2 repository")
    void testLoadSkillsFromMavenLocalM2(@TempDir Path tempDir) throws IOException {
        Path m2Repo = tempDir.resolve(".m2").resolve("repository").resolve("com").resolve("test").resolve("my-artifact").resolve("1.0.0");
        Files.createDirectories(m2Repo);
        Path jarPath = m2Repo.resolve("my-artifact-1.0.0.jar");
        createMockSkillZip(jarPath, "maven-skill", "Mocked skill inside Maven JAR", "");

        String origHome = System.getProperty("user.home");
        try {
            System.setProperty("user.home", tempDir.toAbsolutePath().toString());
            Map<String, SkillDefinition> loaded = SkillLoader.loadSkillsFromMaven("com.test:my-artifact:1.0.0", null, List.of("."), null);
            assertThat(loaded).containsKey("maven-skill");
        } finally {
            if (origHome != null) {
                System.setProperty("user.home", origHome);
            }
        }
    }

    @Test
    @DisplayName("Should load skills from Go module local GOPATH")
    void testLoadSkillsFromGoModuleLocalGopath(@TempDir Path tempDir) throws IOException {
        Path goModDir = tempDir.resolve("go").resolve("pkg").resolve("mod").resolve("github.com").resolve("owner").resolve("repo@v1.0.0");
        Path skillDir = goModDir.resolve("skills").resolve("go-skill");
        Files.createDirectories(skillDir);
        Files.writeString(skillDir.resolve("SKILL.md"), "---\nname: go-skill\ndescription: Mock Go mod skill\n---\n# Go skill\n");

        String origHome = System.getProperty("user.home");
        try {
            System.setProperty("user.home", tempDir.toAbsolutePath().toString());
            Map<String, SkillDefinition> loaded = SkillLoader.loadSkillsFromGoModule("github.com/owner/repo", "v1.0.0", List.of("skills"), null);
            assertThat(loaded).containsKey("go-skill");
        } finally {
            if (origHome != null) {
                System.setProperty("user.home", origHome);
            }
        }
    }

    @Test
    @DisplayName("Should compute valid skill checksum and get loader dir")
    void testCalculateSkillChecksumAndLoaderDir(@TempDir Path tempDir) throws IOException {
        Path skillDir = tempDir.resolve("checksum-skill");
        Files.createDirectories(skillDir);
        Files.writeString(skillDir.resolve("SKILL.md"), "---\nname: checksum-skill\n---\n# Checksum\n");

        String checksum = SkillLoader.calculateSkillChecksum(skillDir);
        assertThat(checksum).isNotEmpty();

        Path loaderDir = SkillLoader.getLoaderSkillsDir();
        assertThat(loaderDir).isNotNull();
    }

    @Test
    @DisplayName("Should test manifest lock read, write, update, and verify")
    void testManifestLockMethods(@TempDir Path tempDir) throws IOException {
        Map<String, Object> lockData = new java.util.HashMap<>(Map.of(
                "version", "1.0.0",
                "skills", new java.util.HashMap<>(Map.of("skill-a", Map.of("uri", "file://a", "sha256", "sha256-111")))
        ));
        Path lockPath = SkillLoader.writeManifestLock(tempDir, lockData);
        assertThat(Files.isRegularFile(lockPath)).isTrue();

        Map<String, Object> readBack = SkillLoader.readManifestLock(tempDir);
        @SuppressWarnings("unchecked")
        Map<String, Object> readSkills = (Map<String, Object>) readBack.get("skills");
        assertThat(readSkills).containsKey("skill-a");

        SkillLoader.updateManifestLock(tempDir, "skill-b", "file://b", "sha256-222");
        Map<String, Object> updated = SkillLoader.readManifestLock(tempDir);
        @SuppressWarnings("unchecked")
        Map<String, Object> updatedSkills = (Map<String, Object>) updated.get("skills");
        assertThat(updatedSkills).containsKeys("skill-a", "skill-b");
    }

    @Test
    @DisplayName("Should test SkillRegistry fromRoots and fromGithub factories")
    void testSkillRegistryFromMethods(@TempDir Path tempDir) throws Exception {
        Path skillDir = tempDir.resolve("skills").resolve("reg-skill");
        Files.createDirectories(skillDir);
        Files.writeString(skillDir.resolve("SKILL.md"), "---\nname: reg-skill\ndescription: Reg\n---\n# Reg\n");

        SkillRegistry regRoots = SkillRegistry.fromRoots(List.of("file://" + tempDir.resolve("skills")), null, null, null);
        assertThat(regRoots.getSkills()).containsKey("reg-skill");

        HttpClient mockClient = Mockito.mock(HttpClient.class);
        @SuppressWarnings("unchecked")
        HttpResponse<Path> mockResponse = (HttpResponse<Path>) Mockito.mock(HttpResponse.class);
        Mockito.when(mockResponse.statusCode()).thenReturn(404);
        Mockito.doReturn(mockResponse).when(mockClient).send(ArgumentMatchers.any(HttpRequest.class), ArgumentMatchers.any());

        SkillLoader.setHttpClient(mockClient);
        try {
            SkillRegistry regGh = SkillRegistry.fromGithub("owner/nonexistent", "main", List.of("."), null, null, null);
            assertThat(regGh.getSkills()).isEmpty();
        } finally {
            SkillLoader.setHttpClient(null);
        }
    }

    private void createMockSkillZip(Path zipPath, String skillName, String skillDescription) throws IOException {
        createMockSkillZip(zipPath, skillName, skillDescription, "extracted-repo/");
    }

    private void createMockSkillZip(Path zipPath, String skillName, String skillDescription, String prefix) throws IOException {
        try (ZipOutputStream zos = new ZipOutputStream(new FileOutputStream(zipPath.toFile()))) {
            ZipEntry entry = new ZipEntry(prefix + "skills/" + skillName + "/SKILL.md");
            zos.putNextEntry(entry);
            String content = "---\nname: " + skillName + "\ndescription: " + skillDescription + "\n---\n# " + skillName + "\n";
            zos.write(content.getBytes(StandardCharsets.UTF_8));
            zos.closeEntry();
        }
    }
}

class SkillLoaderTestUtils {
    static void deleteDir(Path dir) throws IOException {
        if (!Files.exists(dir)) return;
        try (Stream<Path> walk = Files.walk(dir)) {
            walk.sorted(Comparator.reverseOrder()).forEach(p -> {
                try {
                    Files.delete(p);
                } catch (IOException ignored) {}
            });
        }
    }
}

