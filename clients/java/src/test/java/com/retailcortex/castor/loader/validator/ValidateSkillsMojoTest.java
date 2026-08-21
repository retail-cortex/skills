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
import org.apache.maven.plugin.MojoFailureException;
import org.apache.maven.project.MavenProject;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.api.io.TempDir;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.io.File;
import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;

import static org.assertj.core.api.Assertions.assertThat;
import static org.assertj.core.api.Assertions.assertThatThrownBy;

@ExtendWith(MockitoExtension.class)
@DisplayName("ValidateSkillsMojo Unit Tests")
class ValidateSkillsMojoTest {

    @Mock
    private MavenProject mavenProject;

    @Test
    @DisplayName("Should execute validation Mojo successfully on valid workspace")
    void testValidateMojoSuccess(@TempDir Path tempDir) throws Exception {
        File reportFile = tempDir.resolve("target/validator_report.json").toFile();

        ValidateSkillsMojo mojo = new ValidateSkillsMojo();
        mojo.setProject(mavenProject);
        mojo.setSkillsRoot(SkillLoader.findRegistryRoot().toFile());
        mojo.setReportFile(reportFile);
        mojo.setFailOnError(true);

        mojo.execute();

        assertThat(Files.isRegularFile(reportFile.toPath())).isTrue();
    }

    @Test
    @DisplayName("Should throw MojoFailureException when audit fails and failOnError is true")
    void testValidateMojoFailure(@TempDir Path tempDir) throws IOException {
        Path invalidSkillDir = tempDir.resolve("skills/broken-skill");
        Files.createDirectories(invalidSkillDir);
        Files.writeString(invalidSkillDir.resolve("SKILL.md"), "Broken content without frontmatter");

        ValidateSkillsMojo mojo = new ValidateSkillsMojo();
        mojo.setProject(mavenProject);
        mojo.setSkillsRoot(tempDir.toFile());
        mojo.setReportFile(tempDir.resolve("target/validator_report.json").toFile());
        mojo.setFailOnError(true);

        assertThatThrownBy(mojo::execute)
                .isInstanceOf(MojoFailureException.class)
                .hasMessageContaining("SDLC Skill Validation Failed");
    }
}
