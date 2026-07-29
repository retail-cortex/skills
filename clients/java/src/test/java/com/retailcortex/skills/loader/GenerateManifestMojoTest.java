package com.retailcortex.skills.loader;

import org.apache.maven.model.Resource;
import org.apache.maven.project.MavenProject;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.junit.jupiter.api.io.TempDir;
import org.mockito.ArgumentCaptor;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.io.File;
import java.nio.file.Files;
import java.nio.file.Path;

import static org.assertj.core.api.Assertions.assertThat;
import static org.mockito.Mockito.verify;

@ExtendWith(MockitoExtension.class)
@DisplayName("GenerateManifestMojo Unit Tests")
class GenerateManifestMojoTest {

    @Mock
    private MavenProject mavenProject;

    @Test
    @DisplayName("Should execute pre-processor Mojo and add resource to Maven project")
    void testMojoExecution(@TempDir Path tempDir) throws Exception {
        File outputDir = tempDir.resolve("target/generated-resources/skills").toFile();

        GenerateManifestMojo mojo = new GenerateManifestMojo();
        mojo.setProject(mavenProject);
        mojo.setSkillsRoot(SkillLoader.findRegistryRoot().toFile());
        mojo.setOutputDirectory(outputDir);
        mojo.setOutputFilename("skills_manifest.json");

        mojo.execute();

        Path generatedManifest = outputDir.toPath().resolve("skills_manifest.json");
        assertThat(Files.isRegularFile(generatedManifest)).isTrue();

        ArgumentCaptor<Resource> resourceCaptor = ArgumentCaptor.forClass(Resource.class);
        verify(mavenProject).addResource(resourceCaptor.capture());

        Resource addedResource = resourceCaptor.getValue();
        assertThat(addedResource.getDirectory()).isEqualTo(outputDir.getAbsolutePath());
    }
}
