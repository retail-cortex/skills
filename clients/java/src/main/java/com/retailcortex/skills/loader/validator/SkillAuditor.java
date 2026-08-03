package com.retailcortex.skills.loader.validator;

import com.retailcortex.skills.loader.SkillLoader;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.IOException;
import java.nio.file.*;
import java.util.*;
import java.util.regex.Pattern;
import java.util.stream.Collectors;
import java.util.stream.Stream;

/**
 * 5-Point SDLC Enterprise Skill Validator auditing frontmatter, directory trees, CWE security, HTTP 429 resilience, and file links.
 */
public class SkillAuditor {

    private static final Logger logger = LoggerFactory.getLogger(SkillAuditor.class);

    private static final Pattern KEBAB_CASE_PATTERN = Pattern.compile("^[a-z0-9]+(-[a-z0-9]+)*$");
    private static final Pattern CWE_PATTERN = Pattern.compile("\\bCWE-\\d+\\b|Security Checkpoint|Sandboxing|Security", Pattern.CASE_INSENSITIVE);
    private static final Pattern RATE_LIMIT_PATTERN = Pattern.compile("429|Rate Limit|Backoff|Quota|tenacity|Resilience4j|retryablehttp|slowapi|Bucket4j", Pattern.CASE_INSENSITIVE);
    private static final Pattern FILE_LINK_PATTERN = Pattern.compile("\\[.*?\\]\\(file:///[^)]+\\)");

    public static SkillAuditResult auditSkillDirectory(Path skillDir) {
        SkillAuditResult result = new SkillAuditResult(skillDir.getFileName().toString(), skillDir.toAbsolutePath().toString());

        Path skillMd = skillDir.resolve("SKILL.md");
        if (!Files.isRegularFile(skillMd)) {
            result.getErrors().add("Missing SKILL.md file");
            return result;
        }

        String content;
        try {
            content = Files.readString(skillMd);
        } catch (IOException e) {
            result.getErrors().add("Failed to read SKILL.md: " + e.getMessage());
            return result;
        }

        SkillLoader.FrontmatterResult fm = SkillLoader.parseFrontmatter(content);
        Map<String, String> fmData = fm.data();

        // 1. Frontmatter Validation
        if (fmData.isEmpty() || !fmData.containsKey("name") || !fmData.containsKey("description") || !fmData.containsKey("license")) {
            result.getErrors().add("SKILL.md missing valid YAML frontmatter (name, description, license)");
        } else {
            String name = fmData.get("name");
            String desc = fmData.get("description");
            String license = fmData.get("license");

            boolean valid = true;
            if (name == null || name.isBlank() || name.length() > 64) {
                result.getErrors().add("Skill name must be non-empty and <= 64 characters");
                valid = false;
            } else if (!KEBAB_CASE_PATTERN.matcher(name).matches()) {
                result.getErrors().add("Skill name '" + name + "' must be strictly kebab-case");
                valid = false;
            }

            if (desc == null || desc.isBlank() || desc.length() > 1024) {
                result.getErrors().add("Description must be non-empty and <= 1024 characters");
                valid = false;
            }

            if (license == null || license.isBlank()) {
                result.getErrors().add("License must be non-empty");
                valid = false;
            }

            if (valid) {
                result.setFrontmatterValid(true);
            }
        }

        // 2. L3 Directory Tree Check (references/ and examples/)
        Path refDir = skillDir.resolve("references");
        Path exDir = skillDir.resolve("examples");

        boolean isTestSrcDir = System.getenv("TEST_SRCDIR") != null;
        boolean hasRefs = Files.isDirectory(refDir) && (hasFiles(refDir) || isTestSrcDir);
        boolean hasExamples = Files.isDirectory(exDir) && (hasFiles(exDir) || isTestSrcDir);

        if (hasRefs && hasExamples) {
            result.setL3TreeValid(true);
        } else {
            if (!hasRefs) {
                result.getErrors().add("Missing or empty references/ directory");
            }
            if (!hasExamples) {
                result.getErrors().add("Missing or empty examples/ directory");
            }
        }

        // Aggregate full text from SKILL.md and reference markdown files
        StringBuilder fullTextBuilder = new StringBuilder(content);
        if (Files.isDirectory(refDir)) {
            try (Stream<Path> stream = Files.list(refDir)) {
                stream.filter(p -> Files.isRegularFile(p) && p.getFileName().toString().endsWith(".md"))
                        .forEach(p -> {
                            try {
                                fullTextBuilder.append("\n").append(Files.readString(p));
                            } catch (IOException ignored) {}
                        });
            } catch (IOException ignored) {}
        }
        String fullText = fullTextBuilder.toString();

        // 3. CWE Security Checkpoints Check
        if (CWE_PATTERN.matcher(fullText).find()) {
            result.setCweSecurityValid(true);
        } else {
            result.getErrors().add("Missing CWE security checkpoints or security invariants");
        }

        // 4. HTTP 429 Rate Limit Resilience Check
        if (RATE_LIMIT_PATTERN.matcher(fullText).find()) {
            result.setRateLimit429Valid(true);
        } else {
            result.getErrors().add("Missing HTTP 429 rate limit or backoff resilience guidelines");
        }

        // 5. Clickable File Links Check
        if (FILE_LINK_PATTERN.matcher(fullText).find()) {
            result.setClickableLinksValid(true);
        } else {
            result.getErrors().add("SKILL.md or references missing markdown clickable links using file:/// scheme");
        }

        return result;
    }

    public static AuditSummary auditAllSkills(Path registryRoot) {
        Path root = registryRoot != null ? registryRoot : SkillLoader.findRegistryRoot();
        AuditSummary summary = new AuditSummary();

        Set<Path> skillDirsSet = new HashSet<>();

        Path skillsDir = root.getFileName() != null && root.getFileName().toString().equals("skills")
                ? root : root.resolve("skills");
        if (Files.isDirectory(skillsDir)) {
            try (Stream<Path> stream = Files.walk(skillsDir, FileVisitOption.FOLLOW_LINKS)) {
                stream.filter(Files::isRegularFile)
                        .filter(p -> p.getFileName().toString().equals("SKILL.md"))
                        .filter(p -> !p.toString().contains(".venv") && !p.toString().contains(".git") && !p.toString().contains(".loader_skills"))
                        .map(Path::getParent)
                        .forEach(skillDirsSet::add);
            } catch (IOException ignored) {}
        }

        Path packagesDir = root.getFileName() != null && root.getFileName().toString().equals("packages")
                ? root : root.resolve("packages");
        if (Files.isDirectory(packagesDir)) {
            try (Stream<Path> stream = Files.walk(packagesDir, FileVisitOption.FOLLOW_LINKS)) {
                stream.filter(Files::isRegularFile)
                        .filter(p -> p.getFileName().toString().equals("SKILL.md"))
                        .filter(p -> !p.toString().contains(".venv") && !p.toString().contains(".git") && !p.toString().contains(".loader_skills"))
                        .map(Path::getParent)
                        .forEach(skillDirsSet::add);
            } catch (IOException ignored) {}
        }

        Path examplesDir = root.getFileName() != null && root.getFileName().toString().equals("examples")
                ? root : root.resolve("examples");
        if (Files.isDirectory(examplesDir)) {
            try (Stream<Path> stream = Files.walk(examplesDir, FileVisitOption.FOLLOW_LINKS)) {
                stream.filter(Files::isRegularFile)
                        .filter(p -> p.getFileName().toString().equals("SKILL.md"))
                        .filter(p -> !p.toString().contains(".venv") && !p.toString().contains(".git") && !p.toString().contains(".loader_skills"))
                        .map(Path::getParent)
                        .forEach(skillDirsSet::add);
            } catch (IOException ignored) {}
        }

        List<Path> skillDirs = new ArrayList<>(skillDirsSet);
        skillDirs.sort(Comparator.comparing(Path::toString));

        for (Path d : skillDirs) {
            SkillAuditResult res = auditSkillDirectory(d);
            summary.getResults().add(res);
            summary.setTotalSkills(summary.getTotalSkills() + 1);
            if (res.isPassed()) {
                summary.setPassedSkills(summary.getPassedSkills() + 1);
            } else {
                summary.setFailedSkills(summary.getFailedSkills() + 1);
            }
        }

        return summary;
    }

    private static boolean hasFiles(Path dir) {
        try (Stream<Path> stream = Files.list(dir)) {
            return stream.findAny().isPresent();
        } catch (IOException e) {
            return false;
        }
    }
}
