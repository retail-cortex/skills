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

import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.yaml.snakeyaml.Yaml;

import java.io.File;
import java.io.IOException;
import java.io.InputStream;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.nio.charset.StandardCharsets;
import java.nio.file.*;
import java.nio.file.attribute.BasicFileAttributes;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.time.Duration;
import java.util.*;

import java.util.regex.Matcher;
import java.util.regex.Pattern;
import java.util.stream.Collectors;
import java.util.stream.Stream;
import java.util.zip.ZipEntry;
import java.util.zip.ZipInputStream;

/**
 * Dynamic skill scanner and loader for enterprise AI agent skills compatible with Google ADK.
 */
public class SkillLoader {

    private static final Logger logger = LoggerFactory.getLogger(SkillLoader.class);
    private static final ObjectMapper objectMapper = new ObjectMapper().enable(SerializationFeature.INDENT_OUTPUT);
    private static final Pattern FRONTMATTER_PATTERN = Pattern.compile("(?s)^---\\s*\\n(.*?)\\n---\\s*\\n(.*)$");

    public record ParsedUri(String scheme, String target, String ref, String subpath) {}

    public record FrontmatterResult(Map<String, String> data, String body) {}

    private static HttpClient httpClient = null;

    /**
     * For unit testing: inject a custom or mock HttpClient.
     */
    public static void setHttpClient(HttpClient client) {
        httpClient = client;
    }

    private static HttpClient getHttpClient() {
        if (httpClient != null) {
            return httpClient;
        }
        return HttpClient.newBuilder().connectTimeout(Duration.ofSeconds(15)).build();
    }

    /**

     * Discovers the root workspace directory containing enterprise skill packages.
     */
    public static Path findRegistryRoot() {
        String buildWs = System.getenv("BUILD_WORKSPACE_DIRECTORY");
        if (buildWs != null && !buildWs.isBlank()) {
            Path ws = Paths.get(buildWs);
            if (Files.isDirectory(ws.resolve("clients")) || Files.isDirectory(ws.resolve("cmd")) || Files.isDirectory(ws.resolve("pkg")) || Files.isDirectory(ws.resolve("skills")) || Files.isDirectory(ws.resolve("examples"))) {
                return ws;
            }
        }

        Path curr = Paths.get("").toAbsolutePath();
        while (curr != null) {
            String str = curr.toString();
            if (!str.contains("bazel-out") && !str.contains(".runfiles") && !str.contains("sandbox")) {
                if (Files.isDirectory(curr.resolve("clients")) || Files.isDirectory(curr.resolve("cmd")) || Files.isDirectory(curr.resolve("pkg")) || Files.isDirectory(curr.resolve("skills")) || Files.isDirectory(curr.resolve("examples"))) {
                    return curr;
                }
            }
            curr = curr.getParent();
        }

        String testSrcDir = System.getenv("TEST_SRCDIR");
        if (testSrcDir != null && !testSrcDir.isBlank()) {
            Path base = Paths.get(testSrcDir);
            String wsName = System.getenv("TEST_WORKSPACE");
            List<Path> candidates = new ArrayList<>();
            if (wsName != null && !wsName.isBlank()) {
                candidates.add(base.resolve(wsName));
            }
            candidates.add(base.resolve("_main"));
            candidates.add(base.resolve("castor"));
            candidates.add(base.resolve("skill_builder"));
            candidates.add(base);

            for (Path cand : candidates) {
                if (Files.isDirectory(cand.resolve("packages")) || Files.isDirectory(cand.resolve("skills")) || Files.isDirectory(cand.resolve("examples"))) {
                    return cand;
                }
            }
        }

        return Paths.get("").toAbsolutePath();
    }

    /**
     * Returns persistent directory for downloaded skills.
     */
    public static Path getLoaderSkillsDir() {
        Path loaderDir = findRegistryRoot().resolve(".loader_skills");
        try {
            Files.createDirectories(loaderDir);
        } catch (IOException e) {
            logger.warn("Failed to create .loader_skills directory: {}", e.getMessage());
        }
        return loaderDir;
    }

    /**
     * Parses YAML frontmatter block from SKILL.md content.
     */
    public static FrontmatterResult parseFrontmatter(String content) {
        if (content == null || content.isBlank()) {
            return new FrontmatterResult(Collections.emptyMap(), content != null ? content : "");
        }

        Matcher matcher = FRONTMATTER_PATTERN.matcher(content);
        if (!matcher.find()) {
            return new FrontmatterResult(Collections.emptyMap(), content);
        }

        String yamlText = matcher.group(1);
        String body = matcher.group(2);

        Map<String, String> data = new HashMap<>();
        try {
            Yaml yaml = new Yaml();
            Map<String, Object> loaded = yaml.load(yamlText);
            if (loaded != null) {
                for (Map.Entry<String, Object> entry : loaded.entrySet()) {
                    if (entry.getValue() instanceof Map<?, ?> subMap) {
                        for (Map.Entry<?, ?> subEntry : subMap.entrySet()) {
                            if (subEntry.getValue() != null) {
                                data.put(String.valueOf(subEntry.getKey()), String.valueOf(subEntry.getValue()));
                            }
                        }
                    } else if (entry.getValue() != null) {
                        data.put(entry.getKey(), String.valueOf(entry.getValue()));
                    }
                }
            }

        } catch (Exception e) {
            logger.debug("SnakeYAML parsing failed, falling back to line parsing: {}", e.getMessage());
            for (String line : yamlText.split("\n")) {
                line = line.trim();
                if (line.isEmpty() || line.startsWith("#") || !line.contains(":")) {
                    continue;
                }
                String[] parts = line.split(":", 2);
                String k = parts[0].trim();
                String v = parts[1].trim().replaceAll("^['\"]|['\"]$", "");
                data.put(k, v);
            }
        }

        return new FrontmatterResult(data, body);
    }

    /**
     * Parses key-value environment variables from a dotenv file.
     */
    public static Map<String, String> parseDotenvFile(Path dotenvPath) {
        Map<String, String> envVars = new HashMap<>();
        if (dotenvPath == null || !Files.isRegularFile(dotenvPath)) {
            return envVars;
        }

        try {
            List<String> lines = Files.readAllLines(dotenvPath);
            for (String line : lines) {
                line = line.trim();
                if (line.isEmpty() || line.startsWith("#") || !line.contains("=")) {
                    continue;
                }
                String[] parts = line.split("=", 2);
                String k = parts[0].trim();
                String v = parts[1].trim().replaceAll("^['\"]|['\"]$", "");
                envVars.put(k, v);
            }
        } catch (IOException e) {
            logger.debug("Failed to read dotenv file: {}", e.getMessage());
        }
        return envVars;
    }

    /**
     * Parses qualified URI into scheme, target, ref, subpath.
     */
    public static ParsedUri parseSkillRootUri(String uri) {
        if (uri == null) {
            return new ParsedUri("file", ".", null, null);
        }
        String clean = uri.trim();
        for (String p : List.of("castors://", "castor://", "cstrs://", "cstr://", "skms://", "skm://")) {
            if (clean.startsWith(p)) {
                String raw = clean.substring(p.length());
                if (raw.startsWith("skills/")) {
                    raw = raw.substring("skills/".length());
                }
                String schemeName = p.substring(0, p.indexOf("://"));
                return new ParsedUri(schemeName, raw, null, null);
            }
        }
        if (clean.startsWith("file://")) {
            return new ParsedUri("file", clean.substring("file://".length()), null, null);
        }
        if (clean.startsWith("pkg://") || clean.startsWith("package://")) {
            String prefix = clean.startsWith("pkg://") ? "pkg://" : "package://";
            return new ParsedUri("pkg", clean.substring(prefix.length()), null, null);
        }

        if (clean.startsWith("maven://") || clean.startsWith("mvn://")) {
            String prefix = clean.startsWith("maven://") ? "maven://" : "mvn://";
            String raw = clean.substring(prefix.length());
            String target = raw;
            String subpath = null;
            String ref = null;
            int slashIdx = raw.indexOf('/');
            if (slashIdx != -1) {
                target = raw.substring(0, slashIdx);
                subpath = raw.substring(slashIdx + 1);
            }
            if (target.contains(":")) {
                String[] parts = target.split(":");
                if (parts.length >= 3) {
                    ref = parts[2];
                }
            }
            return new ParsedUri("maven", target, ref, subpath);
        }

        if (clean.startsWith("mod://") || clean.startsWith("go://")) {
            String prefix = clean.startsWith("mod://") ? "mod://" : "go://";
            String raw = clean.substring(prefix.length());
            String target = raw;
            String ref = null;
            String subpath = null;
            if (raw.contains("@")) {
                String[] parts = raw.split("@", 2);
                target = parts[0];
                String refSub = parts[1];
                if (refSub.contains("/")) {
                    String[] bits = refSub.split("/", 2);
                    ref = bits[0];
                    subpath = bits[1];
                } else {
                    ref = refSub;
                }
            } else if (raw.contains("/")) {
                String[] bits = raw.split("/");
                if (bits.length >= 3) {
                    target = bits[0] + "/" + bits[1] + "/" + bits[2];
                    if (bits.length > 3) {
                        StringBuilder sb = new StringBuilder();
                        for (int i = 3; i < bits.length; i++) {
                            if (i > 3) sb.append("/");
                            sb.append(bits[i]);
                        }
                        subpath = sb.toString();
                    }
                }
            }
            return new ParsedUri("mod", target, ref, subpath);
        }

        if (clean.startsWith("github://") || clean.startsWith("https://github.com/")) {
            String prefix = clean.startsWith("github://") ? "github://" : "https://github.com/";
            String target = clean.substring(prefix.length());
            if (target.endsWith("/")) {
                target = target.substring(0, target.length() - 1);
            }

            String ref = null;
            String subpath = null;

            if (target.contains("/tree/")) {
                String[] parts = target.split("/tree/", 2);
                target = parts[0];
                String[] treeBits = parts[1].split("/", 2);
                ref = treeBits[0];
                if (treeBits.length > 1) {
                    subpath = treeBits[1];
                }
            } else if (target.contains(":")) {
                int lastColon = target.lastIndexOf(':');
                String part2 = target.substring(lastColon + 1);
                if (!part2.contains("/")) {
                    ref = part2;
                    target = target.substring(0, lastColon);
                } else {
                    String[] parts = target.split(":", 2);
                    target = parts[0];
                    String refSub = parts[1];
                    if (refSub.contains("/")) {
                        String[] bits = refSub.split("/", 2);
                        ref = bits[0];
                        subpath = bits[1];
                    } else {
                        ref = refSub;
                    }
                }
            } else if (target.contains("@")) {
                String[] parts = target.split("@", 2);
                target = parts[0];
                String refSub = parts[1];
                if (refSub.contains("/")) {
                    String[] bits = refSub.split("/", 2);
                    ref = bits[0];
                    subpath = bits[1];
                } else {
                    ref = refSub;
                }
            }

            if (subpath == null) {
                String[] parts = target.split("/");
                if (parts.length > 2) {
                    target = parts[0] + "/" + parts[1];
                    subpath = String.join("/", Arrays.copyOfRange(parts, 2, parts.length));
                }
            }

            return new ParsedUri("github", target, ref != null ? ref : "main", subpath);
        }

        return new ParsedUri("file", clean, null, null);
    }

    /**
     * Loads a single skill definition from a directory.
     */
    public static SkillDefinition loadSkillFromDir(Path skillDir) {
        Path skillMd = skillDir.resolve("SKILL.md");
        if (!Files.isRegularFile(skillMd)) {
            return null;
        }

        try {
            String content = Files.readString(skillMd);
            FrontmatterResult fm = parseFrontmatter(content);
            Map<String, String> fmData = fm.data();

            String name = fmData.getOrDefault("name", skillDir.getFileName().toString());
            String desc = fmData.getOrDefault("description", "Enterprise skill for " + name);
            String allowedTools = fmData.get("allowed-tools");
            if (allowedTools == null) {
                allowedTools = fmData.get("allowed_tools");
            }

            Set<String> knownKeys = Set.of(
                    "name", "description", "license", "author", "authors", "version", "compatibility",
                    "allowed-tools", "allowed_tools", "category", "tags", "trigger_phrases", "execution_hints",
                    "metadata", "source_uri", "src", "source"
            );
            Map<String, String> meta = new HashMap<>();
            for (Map.Entry<String, String> entry : fmData.entrySet()) {
                if (!knownKeys.contains(entry.getKey()) && entry.getValue() != null && !entry.getValue().isBlank()) {
                    meta.put(entry.getKey(), entry.getValue());
                }
            }

            String author = fmData.get("author");
            if (author == null) {
                author = meta.get("author");
            } else if (!meta.containsKey("author")) {
                meta.put("author", author);
            }

            String version = fmData.get("version");
            if (version == null) {
                version = meta.get("version");
            } else if (!meta.containsKey("version")) {
                meta.put("version", version);
            }

            Path regRoot = findRegistryRoot();
            String relPath;
            try {
                relPath = regRoot.toAbsolutePath().relativize(skillDir.toAbsolutePath()).toString();
            } catch (Exception e) {
                relPath = skillDir.toString();
            }

            String sourceUri = fmData.get("source_uri");
            if (sourceUri == null) {
                sourceUri = fmData.get("src");
            }
            if (sourceUri == null) {
                sourceUri = fmData.get("source");
            }
            if (sourceUri == null) {
                sourceUri = "file://" + relPath;
            }

            Map<String, String> references = new TreeMap<>();
            Path refDir = skillDir.resolve("references");
            if (Files.isDirectory(refDir)) {
                try (Stream<Path> stream = Files.list(refDir)) {
                    stream.filter(p -> Files.isRegularFile(p) && p.getFileName().toString().endsWith(".md"))
                            .sorted(Comparator.comparing(p -> p.getFileName().toString()))
                            .forEach(p -> {
                                try {
                                    references.put(p.getFileName().toString(), Files.readString(p));
                                } catch (IOException e) {
                                    logger.debug("Failed reading reference file: {}", e.getMessage());
                                }
                            });
                }
            }

            Map<String, String> examples = new TreeMap<>();
            Path exDir = skillDir.resolve("examples");
            if (Files.isDirectory(exDir)) {
                try (Stream<Path> stream = Files.list(exDir)) {
                    stream.filter(p -> Files.isRegularFile(p) && !p.getFileName().toString().startsWith("."))
                            .sorted(Comparator.comparing(p -> p.getFileName().toString()))
                            .forEach(p -> {
                                try {
                                    examples.put(p.getFileName().toString(), Files.readString(p));
                                } catch (IOException e) {
                                    logger.debug("Failed reading example file: {}", e.getMessage());
                                }
                            });
                }
            }

            String sha256Hex = "";
            try {
                java.security.MessageDigest md = java.security.MessageDigest.getInstance("SHA-256");
                String payload = name + ":" + (version != null ? version : "") + ":" + fm.body().trim() + ":" + desc;
                byte[] hash = md.digest(payload.getBytes(java.nio.charset.StandardCharsets.UTF_8));
                StringBuilder sb = new StringBuilder();
                for (byte b : hash) {
                    sb.append(String.format("%02x", b));
                }
                sha256Hex = sb.toString();
            } catch (Exception ignored) {
            }

            return new SkillDefinition(
                    name,
                    desc,
                    fm.body().trim(),
                    fmData.get("license"),
                    author,
                    version,
                    fmData.get("compatibility"),
                    allowedTools,
                    meta,
                    references,
                    examples,
                    relPath,
                    sourceUri,
                    sha256Hex
            );
        } catch (Exception e) {
            logger.warn("Failed loading skill from {}: {}", skillDir, e.getMessage());
            return null;
        }


    }

    /**
     * Scans and loads skill directories.
     */
    public static Map<String, SkillDefinition> loadAllSkills(Path skillsRoot, List<String> skillFilter) {
        Path root = skillsRoot != null ? skillsRoot : findRegistryRoot();
        Set<String> filterSet = skillFilter != null ? new HashSet<>(skillFilter) : Collections.emptySet();
        boolean hasFilter = !filterSet.isEmpty();
        Map<String, SkillDefinition> loaded = new LinkedHashMap<>();

        if (Files.isRegularFile(root.resolve("SKILL.md"))) {
            SkillDefinition single = loadSkillFromDir(root);
            if (single != null && (filterSet.isEmpty() || filterSet.contains(single.getName()))) {
                loaded.put(single.getName(), single);
                return loaded;
            }
        }

        // 1. Direct scan for skill subdirectories containing SKILL.md
        try (Stream<Path> walk = Files.walk(root, 6, FileVisitOption.FOLLOW_LINKS)) {
            walk.filter(p -> Files.isDirectory(p) && Files.isRegularFile(p.resolve("SKILL.md")))
                .sorted(Comparator.comparing(Path::toString))
                .forEach(skillDir -> {
                    String skillName = skillDir.getFileName().toString();
                    if (hasFilter && !filterSet.contains(skillName)) {
                        return;
                    }
                    SkillDefinition def = loadSkillFromDir(skillDir);
                    if (def != null && (!hasFilter || filterSet.contains(def.getName()))) {
                        loaded.put(def.getName(), def);
                    }
                });
        } catch (IOException e) {
            logger.debug("Error walking root for skills: {}", e.getMessage());
        }

        // 2. Scan packages and examples directories

        List<Path> searchDirs = new ArrayList<>();
        if (root.getFileName() != null && (root.getFileName().toString().equals("packages") || root.getFileName().toString().equals("examples"))) {
            searchDirs.add(root);
        } else {
            searchDirs.add(root.resolve("packages"));
            searchDirs.add(root.resolve("examples"));
        }
        for (Path parentDir : searchDirs) {
            if (Files.isDirectory(parentDir)) {
                try (Stream<Path> stream = Files.walk(parentDir, 5, FileVisitOption.FOLLOW_LINKS)) {

                    List<Path> pkgPaths = stream
                        .filter(p -> Files.isDirectory(p) && p.resolve("src").toFile().exists())
                        .sorted(Comparator.comparing(Path::toString))
                        .toList();

                    for (Path pkg : pkgPaths) {
                        Path srcDir = pkg.resolve("src");

                        if (Files.isDirectory(srcDir)) {
                            try (Stream<Path> walk = Files.walk(srcDir, 4)) {
                                walk.filter(Files::isDirectory)
                                        .filter(p -> p.getFileName().toString().equals("skills"))
                                        .flatMap(skillsFolder -> {
                                            try {
                                                return Files.list(skillsFolder);
                                            } catch (IOException e) {
                                                return Stream.empty();
                                            }
                                        })
                                        .filter(Files::isDirectory)
                                        .sorted(Comparator.comparing(Path::toString))
                                        .forEach(skillDir -> {
                                            String skillName = skillDir.getFileName().toString();
                                            if (hasFilter && !filterSet.contains(skillName)) {
                                                return;
                                            }
                                            SkillDefinition def = loadSkillFromDir(skillDir);
                                            if (def != null) {
                                                if (!hasFilter || filterSet.contains(def.getName())) {
                                                    loaded.put(def.getName(), def);
                                                }
                                            }
                                        });
                            } catch (IOException e) {
                                logger.debug("Error walking package src directory: {}", e.getMessage());
                            }
                        }
                    }
                } catch (IOException e) {

                    logger.debug("Error listing packages directory: {}", e.getMessage());
                }
            }
        }

        // 2. Scan skills directory (and subcategory folders)
        List<Path> skillsDirs = new ArrayList<>();
        if (root.getFileName() != null && root.getFileName().toString().equals("skills")) {
            skillsDirs.add(root);
        } else {
            skillsDirs.add(root.resolve("skills"));
            skillsDirs.add(root.resolve("examples").resolve("skills"));
        }
        for (Path sDir : skillsDirs) {
            if (Files.isDirectory(sDir)) {
                try (Stream<Path> stream = Files.walk(sDir, FileVisitOption.FOLLOW_LINKS)) {
                    stream.filter(Files::isRegularFile)
                            .filter(p -> p.getFileName().toString().equals("SKILL.md"))
                            .filter(p -> !p.toString().contains(".venv") && !p.toString().contains(".git"))
                            .map(Path::getParent)
                            .sorted(Comparator.comparing(p -> p.getFileName().toString()))
                            .forEach(skillDir -> {
                                String skillName = skillDir.getFileName().toString();
                                if (loaded.containsKey(skillName)) return;
                                if (!filterSet.isEmpty() && !filterSet.contains(skillName)) return;

                                SkillDefinition def = loadSkillFromDir(skillDir);
                                if (def != null && (filterSet.isEmpty() || filterSet.contains(def.getName()))) {
                                    loaded.put(def.getName(), def);
                                }
                            });
                } catch (IOException e) {
                    logger.debug("Error scanning skills directory: {}", e.getMessage());
                }
            }
        }

        // 3. Scan standard cross-client .agents/skills directories (project-level & user-level)
        List<Path> agentDirs = new ArrayList<>();
        agentDirs.add(root.resolve(".agents").resolve("skills"));
        String userHome = System.getProperty("user.home");
        if (userHome != null && !userHome.isBlank()) {
            agentDirs.add(Paths.get(userHome).resolve(".agents").resolve("skills"));
        }

        for (Path agDir : agentDirs) {
            if (Files.isDirectory(agDir)) {
                try (Stream<Path> stream = Files.list(agDir)) {
                    stream.filter(Files::isDirectory)
                            .filter(p -> !p.getFileName().toString().startsWith("."))
                            .sorted(Comparator.comparing(p -> p.getFileName().toString()))
                            .forEach(skillDir -> {
                                String skillName = skillDir.getFileName().toString();
                                if (loaded.containsKey(skillName)) return;
                                if (!filterSet.isEmpty() && !filterSet.contains(skillName)) return;

                                SkillDefinition def = loadSkillFromDir(skillDir);
                                if (def != null && (filterSet.isEmpty() || filterSet.contains(def.getName()))) {
                                    loaded.put(def.getName(), def);
                                }
                            });
                } catch (IOException e) {
                    logger.debug("Error scanning .agents/skills directory: {}", e.getMessage());
                }
            }
        }

        return loaded;
    }

    /**
     * Loads skills from a package name.
     */
    public static Map<String, SkillDefinition> loadSkillsFromPackage(String packageName, List<String> filter) {
        if (packageName == null || packageName.isBlank()) {
            return Collections.emptyMap();
        }
        String cleanPkg = packageName.trim();
        Path root = findRegistryRoot();
        List<Path> searchDirs = new ArrayList<>();
        if (root.getFileName() != null && (root.getFileName().toString().equals("packages") || root.getFileName().toString().equals("examples"))) {
            searchDirs.add(root);
        } else {
            searchDirs.add(root.resolve("packages"));
            searchDirs.add(root.resolve("examples"));
        }

        Map<String, SkillDefinition> loaded = new HashMap<>();
        for (Path parentDir : searchDirs) {
            if (Files.isDirectory(parentDir)) {
                try (Stream<Path> stream = Files.walk(parentDir, 5, FileVisitOption.FOLLOW_LINKS)) {
                    List<Path> allDirs = stream.filter(Files::isDirectory).toList();
                    List<Path> pPaths = allDirs.stream().filter(p -> p.resolve("src").toFile().exists()).toList();
                    for (Path p : pPaths) {



                        Path srcDir = p.resolve("src");
                        if (Files.isDirectory(srcDir)) {
                            Path pkgDir = srcDir.resolve(cleanPkg);
                            if (!Files.isDirectory(pkgDir)) {
                                String altName = cleanPkg.startsWith("retailcortex_skills_") ?
                                    cleanPkg.substring("retailcortex_skills_".length()) :
                                    "retailcortex_skills_" + cleanPkg.replace("-", "_");
                                Path altDir = srcDir.resolve(altName);
                                if (Files.isDirectory(altDir)) {
                                    pkgDir = altDir;
                                }
                            }
                            if (Files.isDirectory(pkgDir)) {
                                Path skillsSub = pkgDir.resolve("skills");
                                Path target = Files.isDirectory(skillsSub) ? skillsSub : pkgDir;
                                loaded.putAll(loadAllSkills(target, filter));
                            }
                        }
                    }
                } catch (IOException e) {
                    logger.debug("Error searching packages: {}", e.getMessage());
                }
            }
        }

        System.err.println("DEBUG loadSkillsFromPackage cleanPkg=" + cleanPkg + " root=" + root + " searchDirs=" + searchDirs + " loaded=" + loaded.keySet());
        return loaded;
    }


    /**
     * Loads skills from a Maven artifact coordinates or classpath.
     */
    public static Map<String, SkillDefinition> loadSkillsFromMaven(String target, String ref, List<String> roots, List<String> filter) {
        if (target == null || target.isBlank()) {
            return Collections.emptyMap();
        }
        String[] parts = target.split(":");
        String groupId = parts.length >= 1 ? parts[0] : "";
        String artifactId = parts.length >= 2 ? parts[1] : target;
        String version = parts.length >= 3 ? parts[2] : (ref != null ? ref : "");

        String homeDir = System.getProperty("user.home");
        Path jarPath = null;
        if (!groupId.isBlank() && !artifactId.isBlank() && !version.isBlank()) {
            String groupPath = groupId.replace('.', '/');
            Path cand = Paths.get(homeDir, ".m2", "repository", groupPath, artifactId, version, artifactId + "-" + version + ".jar");
            if (Files.isRegularFile(cand)) {
                jarPath = cand;
            }
        }

        if (jarPath != null) {
            String cacheKey = groupId + "-" + artifactId + "-" + version;
            Path extractedDir = getLoaderSkillsDir().resolve("maven").resolve(cacheKey);
            if (!Files.isDirectory(extractedDir)) {
                try {
                    unzip(jarPath, extractedDir);
                } catch (IOException e) {
                    logger.warn("Failed to extract maven jar {}: {}", jarPath, e.getMessage());
                }
            }
            if (roots != null && !roots.isEmpty() && roots.get(0) != null) {
                Path sub = extractedDir.resolve(roots.get(0));
                if (Files.isDirectory(sub)) {
                    extractedDir = sub;
                }
            }
            return loadAllSkills(extractedDir, filter);
        }

        Path root = findRegistryRoot();
        String cleanId = artifactId.startsWith("skills-") ? artifactId.substring(7) : artifactId;
        List<Path> candidates = List.of(
            root.resolve("examples").resolve(cleanId).resolve("skills"),
            root.resolve("examples").resolve("skills-" + cleanId),
            root.resolve("examples").resolve(cleanId),
            root.resolve("packages").resolve("skills-" + cleanId),
            root.resolve("packages").resolve(cleanId),
            root.resolve("clients").resolve("java")
        );

        for (Path cand : candidates) {
            if (Files.isDirectory(cand)) {
                Map<String, SkillDefinition> skills = loadAllSkills(cand, filter);
                if (!skills.isEmpty()) {
                    return skills;
                }
            }
        }
        return Collections.emptyMap();
    }

    /**
     * Loads skills from a Go module URI or GOPATH cache.
     */
    public static Map<String, SkillDefinition> loadSkillsFromGoModule(String target, String ref, List<String> roots, List<String> filter) {
        if (target == null || target.isBlank()) {
            return Collections.emptyMap();
        }
        String cleanMod = target.trim();
        String gopath = System.getenv("GOPATH");
        if (gopath == null || gopath.isBlank()) {
            gopath = Paths.get(System.getProperty("user.home"), "go").toString();
        }

        Path modDir = null;
        if (ref != null && !ref.isBlank() && !gopath.isBlank()) {
            Path cand = Paths.get(gopath, "pkg", "mod", cleanMod.toLowerCase() + "@" + ref);
            if (Files.isDirectory(cand)) {
                modDir = cand;
            }
        }

        if (modDir != null) {
            Path candidateDir = modDir;
            if (roots != null && !roots.isEmpty() && roots.get(0) != null) {
                Path sub = modDir.resolve(roots.get(0));
                if (Files.isDirectory(sub)) {
                    candidateDir = sub;
                }
            }
            return loadAllSkills(candidateDir, filter);
        }

        Path root = findRegistryRoot();
        String artName = Paths.get(cleanMod).getFileName().toString();
        String cleanName = artName.startsWith("skills-") ? artName.substring(7) : artName;
        List<Path> candidates = List.of(
            root.resolve("examples").resolve(cleanName).resolve("skills"),
            root.resolve("examples").resolve("skills-" + cleanName),
            root.resolve("examples").resolve(cleanName),
            root.resolve("packages").resolve("skills-" + cleanName),
            root.resolve("packages").resolve(cleanName),
            root.resolve("clients").resolve("go")
        );

        for (Path cand : candidates) {
            if (Files.isDirectory(cand)) {
                Map<String, SkillDefinition> skills = loadAllSkills(cand, filter);
                if (!skills.isEmpty()) {
                    return skills;
                }
            }
        }
        return Collections.emptyMap();
    }

    /**
     * Loads skills from a remote GitHub repository.
     */
    public static Map<String, SkillDefinition> loadSkillsFromGithub(String repo, String ref, List<String> roots,
                                                                    List<String> filter, String token, Path dotenvPath) {
        Path envFile = dotenvPath != null ? dotenvPath : Paths.get("").toAbsolutePath().resolve(".env");
        Map<String, String> dotenvVars = parseDotenvFile(envFile);

        String cleanRepo = repo.trim();
        if (cleanRepo.startsWith("https://github.com/")) {
            cleanRepo = cleanRepo.substring("https://github.com/".length());
            if (cleanRepo.endsWith("/")) cleanRepo = cleanRepo.substring(0, cleanRepo.length() - 1);
        }
        if (cleanRepo.endsWith(".git")) {
            cleanRepo = cleanRepo.substring(0, cleanRepo.length() - 4);
        }

        List<String> urlRoots = null;
        if (cleanRepo.contains("/tree/")) {
            String[] parts = cleanRepo.split("/tree/", 2);
            cleanRepo = parts[0].replaceAll("/$", "");
            String[] treeParts = parts[1].split("/", 2);
            if (ref == null) ref = treeParts[0];
            if (treeParts.length > 1 && !treeParts[1].isBlank()) {
                urlRoots = List.of(treeParts[1].trim());
            }
        }

        String gitToken = token != null ? token : System.getenv("GITHUB_TOKEN");
        if (gitToken == null) gitToken = System.getenv("GH_TOKEN");
        if (gitToken == null) gitToken = dotenvVars.get("GITHUB_TOKEN");
        if (gitToken == null) gitToken = dotenvVars.get("GH_TOKEN");

        String gitRef = ref != null ? ref : System.getenv("GITHUB_REF");
        if (gitRef == null) gitRef = dotenvVars.get("GITHUB_REF");
        if (gitRef == null) gitRef = "main";

        List<String> rootPaths;
        if (roots != null && !roots.isEmpty()) {
            rootPaths = roots;
        } else if (urlRoots != null && !urlRoots.isEmpty()) {
            rootPaths = urlRoots;
        } else if (System.getenv("SKILLS_ROOTS") != null) {
            rootPaths = Arrays.stream(System.getenv("SKILLS_ROOTS").split(","))
                    .map(String::trim).filter(s -> !s.isEmpty()).collect(Collectors.toList());
        } else if (dotenvVars.containsKey("SKILLS_ROOTS")) {
            rootPaths = Arrays.stream(dotenvVars.get("SKILLS_ROOTS").split(","))
                    .map(String::trim).filter(s -> !s.isEmpty()).collect(Collectors.toList());
        } else {
            rootPaths = List.of("skills", ".");
        }

        String repoSlug = cleanRepo.replace("/", "_");
        Path loaderBase = getLoaderSkillsDir();
        Path persistentRepoDir = loaderBase.resolve("github").resolve(repoSlug).resolve(gitRef);
        try {
            Files.createDirectories(persistentRepoDir);
        } catch (IOException ignored) {}

        Path tmpDir = null;
        Path repoTargetDir = persistentRepoDir;
        try {
            tmpDir = Files.createTempDirectory("skills-loader-gh-");
            Path zipPath = tmpDir.resolve("repo.zip");
            String archiveUrl = String.format("https://api.github.com/repos/%s/zipball/%s", cleanRepo, gitRef);

            HttpClient client = getHttpClient();
            HttpRequest.Builder reqBuilder = HttpRequest.newBuilder().uri(URI.create(archiveUrl))
                    .header("User-Agent", "skills-loader-java/1.0.0")
                    .GET();
            if (gitToken != null && !gitToken.isBlank()) {
                reqBuilder.header("Authorization", "token " + gitToken);
            }

            HttpResponse<Path> resp = client.send(reqBuilder.build(), HttpResponse.BodyHandlers.ofFile(zipPath));
            if (resp.statusCode() == 200) {
                unzip(zipPath, tmpDir);
                try (Stream<Path> stream = Files.list(tmpDir)) {
                    Optional<Path> extracted = stream.filter(Files::isDirectory).findFirst();
                    if (extracted.isPresent()) {
                        repoTargetDir = extracted.get();
                    }
                }
            }
        } catch (Exception e) {
            logger.debug("Failed downloading GitHub archive: {}", e.getMessage());
        }

        Map<String, SkillDefinition> loaded = new HashMap<>();
        for (String rootRel : rootPaths) {
            Path candidate = rootRel.equals(".") ? repoTargetDir : repoTargetDir.resolve(rootRel);
            if (Files.isDirectory(candidate)) {
                SkillDefinition single = loadSkillFromDir(candidate);
                if (single != null) {
                    if (filter == null || filter.contains(single.getName())) {
                        loaded.put(single.getName(), single);
                    }
                } else {
                    loaded.putAll(loadAllSkills(candidate, filter));
                }
            }
        }

        if (tmpDir != null) {
            deleteRecursively(tmpDir);
        }

        return loaded;
    }

    /**
     * Loads skills across multiple qualified URIs.
     */
    public static Map<String, SkillDefinition> loadSkillsFromRoots(List<String> roots, List<String> filter,
                                                                  String token, Path dotenvPath) {
        Path envFile = dotenvPath != null ? dotenvPath : Paths.get("").toAbsolutePath().resolve(".env");
        Map<String, String> dotenvVars = parseDotenvFile(envFile);

        List<String> rootURIs = roots;
        if (rootURIs == null || rootURIs.isEmpty()) {
            if (System.getenv("SKILLS_ROOTS") != null) {
                rootURIs = Arrays.stream(System.getenv("SKILLS_ROOTS").split(","))
                        .map(String::trim).filter(s -> !s.isEmpty()).collect(Collectors.toList());
            } else if (dotenvVars.containsKey("SKILLS_ROOTS")) {
                rootURIs = Arrays.stream(dotenvVars.get("SKILLS_ROOTS").split(","))
                        .map(String::trim).filter(s -> !s.isEmpty()).collect(Collectors.toList());
            } else {
                rootURIs = List.of("file://.");
            }
        }

        Map<String, SkillDefinition> loaded = new LinkedHashMap<>();
        for (String uri : rootURIs) {
            ParsedUri parsed = parseSkillRootUri(uri);
            switch (parsed.scheme()) {
                case "file" -> {
                    Path p = Paths.get(parsed.target());
                    if (!p.isAbsolute()) {
                        Path base = findRegistryRoot();
                        Path cand = base.resolve(parsed.target());
                        if (Files.isDirectory(cand) || Files.isRegularFile(cand)) {
                            p = cand;
                        } else {
                            p = p.toAbsolutePath();
                        }
                    }
                    if (Files.isDirectory(p)) {
                        SkillDefinition single = loadSkillFromDir(p);
                        if (single != null) {
                            if (filter == null || filter.contains(single.getName())) {
                                loaded.put(single.getName(), single);
                            }
                        } else {
                            loaded.putAll(loadAllSkills(p, filter));
                        }
                    }
                }
                case "pkg" -> loaded.putAll(loadSkillsFromPackage(parsed.target(), filter));
                case "maven", "mvn" -> {
                    List<String> mavenRoots = parsed.subpath() != null ? List.of(parsed.subpath()) : null;
                    loaded.putAll(loadSkillsFromMaven(parsed.target(), parsed.ref(), mavenRoots, filter));
                }
                case "mod", "go" -> {
                    List<String> modRoots = parsed.subpath() != null ? List.of(parsed.subpath()) : null;
                    loaded.putAll(loadSkillsFromGoModule(parsed.target(), parsed.ref(), modRoots, filter));
                }
                case "github" -> {
                    List<String> ghRoots = parsed.subpath() != null ? List.of(parsed.subpath()) : null;
                    loaded.putAll(loadSkillsFromGithub(parsed.target(), parsed.ref(), ghRoots, filter, token, dotenvPath));
                }
            }
        }
        return loaded;
    }

    /**
     * Builds pre-compiled JSON skills manifest file.
     */
    public static Path buildSkillsManifest(Path skillsRoot, Path outputPath) throws IOException {
        Map<String, SkillDefinition> skills = loadAllSkills(skillsRoot, null);
        Path outFile = outputPath != null ? outputPath : getLoaderSkillsDir().resolve("skills_manifest.json");

        Path manifestDir = outFile.getParent() != null ? outFile.getParent().toAbsolutePath().normalize() : Paths.get(".").toAbsolutePath().normalize();
        Files.createDirectories(manifestDir);

        Path registryRoot = findRegistryRoot().toAbsolutePath().normalize();

        Map<String, Map<String, Object>> manifestData = new LinkedHashMap<>();
        for (Map.Entry<String, SkillDefinition> entry : skills.entrySet()) {
            SkillDefinition s = entry.getValue();
            Path skillAbs = Paths.get(s.getPath());
            if (!skillAbs.isAbsolute()) {
                skillAbs = registryRoot.resolve(skillAbs).toAbsolutePath().normalize();
            }

            String relPath;
            try {
                relPath = manifestDir.relativize(skillAbs).toString();
            } catch (Exception e) {
                relPath = s.getPath();
            }

            String srcUri = s.getSourceUri();
            if (srcUri == null || srcUri.isEmpty() || srcUri.startsWith("file://")) {
                srcUri = "file://" + relPath;
            }

            Map<String, Object> map = s.toMap();
            map.put("path", relPath);
            map.put("source_uri", srcUri);
            map.put("references", s.getReferences());
            map.put("examples", s.getExamples());
            manifestData.put(entry.getKey(), map);
        }

        objectMapper.writeValue(outFile.toFile(), manifestData);
        return outFile;
    }

    /**
     * Loads skill definitions from a JSON manifest file.
     */
    public static Map<String, SkillDefinition> loadSkillsFromManifest(Path manifestPath) {
        if (manifestPath == null || !Files.isRegularFile(manifestPath)) {
            return Collections.emptyMap();
        }

        try {
            Map<String, SkillDefinition> result = objectMapper.readValue(
                    manifestPath.toFile(), new TypeReference<Map<String, SkillDefinition>>() {});
            return result != null ? result : Collections.emptyMap();
        } catch (IOException e) {
            logger.warn("Failed reading skills manifest from {}: {}", manifestPath, e.getMessage());
            return Collections.emptyMap();
        }
    }

    // --- Helper Utilities ---

    private static void unzip(Path zipFile, Path targetDir) throws IOException {
        try (ZipInputStream zis = new ZipInputStream(Files.newInputStream(zipFile))) {
            ZipEntry entry;
            while ((entry = zis.getNextEntry()) != null) {
                Path destPath = targetDir.resolve(entry.getName()).normalize();
                if (!destPath.startsWith(targetDir)) {
                    throw new IOException("Zip entry attempted directory traversal: " + entry.getName());
                }
                if (entry.isDirectory()) {
                    Files.createDirectories(destPath);
                } else {
                    if (destPath.getParent() != null) {
                        Files.createDirectories(destPath.getParent());
                    }
                    Files.copy(zis, destPath, StandardCopyOption.REPLACE_EXISTING);
                }
                zis.closeEntry();
            }
        }
    }

    private static void deleteRecursively(Path dir) {
        try {
            Files.walkFileTree(dir, new SimpleFileVisitor<>() {
                @Override
                public FileVisitResult visitFile(Path file, BasicFileAttributes attrs) throws IOException {
                    Files.delete(file);
                    return FileVisitResult.CONTINUE;
                }
                @Override
                public FileVisitResult postVisitDirectory(Path dir, IOException exc) throws IOException {
                    Files.delete(dir);
                    return FileVisitResult.CONTINUE;
                }
            });
        } catch (IOException ignored) {}
    }

    public record ManifestLockEntry(String skill_name, String uri, String sha256) {}

    public record VerificationResult(String skillName, String uri, String status, String expectedSha256, String actualSha256, String error) {}

    public record VerificationReport(String targetDir, int totalSkills, int verifiedCount, int modifiedCount, int missingCount, List<VerificationResult> results) {}

    public static String calculateSkillChecksum(Path skillDir) throws IOException {
        MessageDigest digest;
        try {
            digest = MessageDigest.getInstance("SHA-256");
        } catch (NoSuchAlgorithmException e) {
            throw new RuntimeException("SHA-256 algorithm not found", e);
        }
        Path absDir = skillDir.toAbsolutePath().normalize();
        if (!Files.isDirectory(absDir)) {
            throw new IllegalArgumentException("Skill directory does not exist: " + absDir);
        }

        List<Path> filePaths = new ArrayList<>();
        try (Stream<Path> stream = Files.walk(absDir)) {
            stream.filter(Files::isRegularFile)
                    .filter(p -> !p.getFileName().toString().equals(".DS_Store") && !p.getFileName().toString().equals(".manifest.lock"))
                    .forEach(filePaths::add);
        }

        filePaths.sort(Comparator.comparing(p -> absDir.relativize(p).toString().replace('\\', '/')));
        for (Path p : filePaths) {
            String relPath = absDir.relativize(p).toString().replace('\\', '/');
            digest.update(relPath.getBytes(StandardCharsets.UTF_8));
            digest.update(Files.readAllBytes(p));
        }

        byte[] hash = digest.digest();
        StringBuilder hexString = new StringBuilder();
        for (byte b : hash) {
            String hex = Integer.toHexString(0xff & b);
            if (hex.length() == 1) hexString.append('0');
            hexString.append(hex);
        }
        return hexString.toString();
    }

    public static Map<String, Object> readManifestLock(Path destDir) throws IOException {
        Path lockFile = destDir.resolve(".manifest.lock");
        if (!Files.isRegularFile(lockFile)) {
            throw new NoSuchFileException(".manifest.lock file not found in " + destDir);
        }
        return objectMapper.readValue(lockFile.toFile(), new TypeReference<Map<String, Object>>() {});
    }

    public static Path writeManifestLock(Path destDir, Map<String, Object> lockData) throws IOException {
        Files.createDirectories(destDir);
        Path lockFile = destDir.resolve(".manifest.lock");
        Map<String, Object> copy = new HashMap<>(lockData);
        if (!copy.containsKey("version")) {
            copy.put("version", "1.0.0");
        }
        if (!copy.containsKey("skills")) {
            copy.put("skills", new HashMap<String, Object>());
        }
        objectMapper.writeValue(lockFile.toFile(), copy);
        return lockFile;
    }

    public static Path updateManifestLock(Path destDir, String skillName, String uri, String checksum) throws IOException {
        Map<String, Object> lockData;
        try {
            lockData = readManifestLock(destDir);
        } catch (Exception e) {
            lockData = new HashMap<>();
            lockData.put("version", "1.0.0");
            lockData.put("skills", new HashMap<String, Object>());
        }

        if (checksum == null || checksum.isBlank()) {
            checksum = calculateSkillChecksum(destDir.resolve(skillName));
        }

        @SuppressWarnings("unchecked")
        Map<String, Object> skills = (Map<String, Object>) lockData.computeIfAbsent("skills", k -> new HashMap<String, Object>());

        Map<String, String> entry = new HashMap<>();
        entry.put("skill_name", skillName);
        entry.put("uri", uri);
        entry.put("sha256", checksum);
        skills.put(skillName, entry);

        return writeManifestLock(destDir, lockData);
    }

    public static VerificationReport verifyManifestLock(Path destDir) throws IOException {
        Map<String, Object> lockData = readManifestLock(destDir);
        @SuppressWarnings("unchecked")
        Map<String, Map<String, Object>> skills = (Map<String, Map<String, Object>>) lockData.getOrDefault("skills", Collections.emptyMap());

        List<VerificationResult> results = new ArrayList<>();
        int verifiedCount = 0;
        int modifiedCount = 0;
        int missingCount = 0;

        List<String> sortedNames = new ArrayList<>(skills.keySet());
        Collections.sort(sortedNames);

        for (String name : sortedNames) {
            Map<String, Object> entry = skills.get(name);
            String uri = String.valueOf(entry.getOrDefault("uri", ""));
            String expectedSha = String.valueOf(entry.getOrDefault("sha256", ""));

            Path skillDir = destDir.resolve(name);
            if (!Files.isDirectory(skillDir)) {
                missingCount++;
                results.add(new VerificationResult(name, uri, "missing", expectedSha, null, "skill directory missing"));
                continue;
            }

            try {
                String currentSha = calculateSkillChecksum(skillDir);
                if (currentSha.equalsIgnoreCase(expectedSha)) {
                    verifiedCount++;
                    results.add(new VerificationResult(name, uri, "verified", expectedSha, currentSha, null));
                } else {
                    modifiedCount++;
                    results.add(new VerificationResult(name, uri, "modified", expectedSha, currentSha, "checksum mismatch (skill files modified)"));
                }
            } catch (Exception e) {
                modifiedCount++;
                results.add(new VerificationResult(name, uri, "modified", expectedSha, null, "failed to compute checksum: " + e.getMessage()));
            }
        }

        return new VerificationReport(destDir.toAbsolutePath().toString(), skills.size(), verifiedCount, modifiedCount, missingCount, results);
    }
}

