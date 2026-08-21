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

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.retailcortex.castor.loader.SkillLoader;
import org.apache.maven.plugin.AbstractMojo;
import org.apache.maven.plugin.MojoExecutionException;
import org.apache.maven.plugin.MojoFailureException;
import org.apache.maven.plugins.annotations.LifecyclePhase;
import org.apache.maven.plugins.annotations.Mojo;
import org.apache.maven.plugins.annotations.Parameter;
import org.apache.maven.project.MavenProject;

import java.io.File;
import java.nio.file.Files;
import java.nio.file.Path;

/**
 * Maven Plugin Mojo executing the 5-Point SDLC Skill Audit during Maven's 'validate' build phase.
 */
@Mojo(name = "validate", defaultPhase = LifecyclePhase.VALIDATE, threadSafe = true)
public class ValidateSkillsMojo extends AbstractMojo {

    private static final ObjectMapper objectMapper = new ObjectMapper().enable(SerializationFeature.INDENT_OUTPUT);

    @Parameter(defaultValue = "${project}", readonly = true, required = true)
    private MavenProject project;

    @Parameter(property = "skillsRoot", defaultValue = "${project.basedir}")
    private File skillsRoot;

    @Parameter(property = "failOnError", defaultValue = "true")
    private boolean failOnError;

    @Parameter(property = "reportFile", defaultValue = "${project.build.directory}/validator_report.json")
    private File reportFile;

    @Parameter(property = "skip", defaultValue = "false")
    private boolean skip;

    @Override
    public void execute() throws MojoExecutionException, MojoFailureException {
        if (skip) {
            getLog().info("Skipping enterprise skills validation audit.");
            return;
        }

        try {
            getLog().info("================================================================");
            getLog().info("Starting 5-Point SDLC Enterprise Skills Validation Audit...");
            getLog().info("================================================================");

            Path rootPath = skillsRoot != null ? skillsRoot.toPath() : SkillLoader.findRegistryRoot();
            AuditSummary summary = SkillAuditor.auditAllSkills(rootPath);

            getLog().info(String.format("Audit Completed: %d skills evaluated | %d PASSED | %d FAILED",
                    summary.getTotalSkills(), summary.getPassedSkills(), summary.getFailedSkills()));

            for (SkillAuditResult res : summary.getResults()) {
                if (res.isPassed()) {
                    getLog().info(String.format(" [PASS] %-30s (%s)", res.getSkillName(), res.getDirectoryPath()));
                } else {
                    getLog().error(String.format(" [FAIL] %-30s (%s)", res.getSkillName(), res.getDirectoryPath()));
                    for (String err : res.getErrors()) {
                        getLog().error("   --> Error: " + err);
                    }
                }
            }

            if (reportFile != null) {
                Path rFile = reportFile.toPath();
                if (rFile.getParent() != null) {
                    Files.createDirectories(rFile.getParent());
                }
                objectMapper.writeValue(rFile.toFile(), summary);
                getLog().info("Saved audit report to: " + rFile.toAbsolutePath());
            }

            if (summary.getFailedSkills() > 0 && failOnError) {
                throw new MojoFailureException(String.format(
                        "SDLC Skill Validation Failed: %d out of %d skills failed audit standards.",
                        summary.getFailedSkills(), summary.getTotalSkills()));
            }

        } catch (MojoFailureException e) {
            throw e;
        } catch (Exception e) {
            getLog().error("Error executing skill validation audit", e);
            throw new MojoExecutionException("Critical error during 5-point SDLC skill audit", e);
        }
    }

    public void setProject(MavenProject project) {
        this.project = project;
    }

    public void setSkillsRoot(File skillsRoot) {
        this.skillsRoot = skillsRoot;
    }

    public void setFailOnError(boolean failOnError) {
        this.failOnError = failOnError;
    }

    public void setReportFile(File reportFile) {
        this.reportFile = reportFile;
    }

    public void setSkip(boolean skip) {
        this.skip = skip;
    }
}
