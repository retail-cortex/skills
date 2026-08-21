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

package com.retailcortex.castor.loader.validator;

import com.retailcortex.castor.loader.SkillLoader;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.io.TempDir;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;

import static org.assertj.core.api.Assertions.assertThat;

@DisplayName("SkillAuditor 5-Point SDLC Audit Unit Tests")
class SkillAuditorTest {

    @Test
    @DisplayName("Should pass 5-point SDLC audit on valid skill directory")
    void testAuditSkillDirectorySuccess(@TempDir Path tempDir) throws IOException {
        Path skillDir = tempDir.resolve("valid-skill");
        Files.createDirectories(skillDir.resolve("references"));
        Files.createDirectories(skillDir.resolve("examples"));

        String skillMdContent = """
                ---
                name: valid-skill
                description: Valid enterprise skill description for testing.
                license: Apache-2.0
                ---
                # Overview
                Follow security guidelines.
                
                ## Security Checkpoints
                - CWE-200: Sandboxing enabled for prompt injection defence.
                - Rate Limit: HTTP 429 backoff retry logic.
                
                See reference documentation in [utils.py](file:///path/to/utils.py#L1-L10).
                """;

        Files.writeString(skillDir.resolve("SKILL.md"), skillMdContent);
        Files.writeString(skillDir.resolve("references/ref1.md"), "Reference content");
        Files.writeString(skillDir.resolve("examples/ex1.java"), "Example content");

        SkillAuditResult result = SkillAuditor.auditSkillDirectory(skillDir);
        assertThat(result.isFrontmatterValid()).isTrue();
        assertThat(result.isL3TreeValid()).isTrue();
        assertThat(result.isCweSecurityValid()).isTrue();
        assertThat(result.isRateLimit429Valid()).isTrue();
        assertThat(result.isClickableLinksValid()).isTrue();
        assertThat(result.isPassed()).isTrue();
        assertThat(result.getErrors()).isEmpty();
    }

    @Test
    @DisplayName("Should detect non-kebab-case name and missing directories")
    void testAuditSkillDirectoryFailures(@TempDir Path tempDir) throws IOException {
        Path skillDir = tempDir.resolve("invalidSkill");
        Files.createDirectories(skillDir);

        String skillMdContent = """
                ---
                name: invalidSkill
                description: Invalid skill
                ---
                # Overview
                Plain text content with no invariants.
                """;

        Files.writeString(skillDir.resolve("SKILL.md"), skillMdContent);

        SkillAuditResult result = SkillAuditor.auditSkillDirectory(skillDir);
        assertThat(result.isFrontmatterValid()).isFalse();
        assertThat(result.isL3TreeValid()).isFalse();
        assertThat(result.isCweSecurityValid()).isFalse();
        assertThat(result.isRateLimit429Valid()).isFalse();
        assertThat(result.isClickableLinksValid()).isFalse();
        assertThat(result.isPassed()).isFalse();
        assertThat(result.getErrors()).isNotEmpty();
    }

    @Test
    @DisplayName("Should audit all 23 skills in the workspace cleanly")
    void testAuditAllSkillsInWorkspace() {
        Path workspaceRoot = SkillLoader.findRegistryRoot();
        AuditSummary summary = SkillAuditor.auditAllSkills(workspaceRoot);

        assertThat(summary.getTotalSkills()).isGreaterThanOrEqualTo(20);
        assertThat(summary.getPassedSkills()).isGreaterThanOrEqualTo(20);
        assertThat(summary.getFailedSkills()).isEqualTo(0);
    }
}
