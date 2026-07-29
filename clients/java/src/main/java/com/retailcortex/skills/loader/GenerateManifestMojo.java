package com.retailcortex.skills.loader;

import org.apache.maven.model.Resource;
import org.apache.maven.plugin.AbstractMojo;
import org.apache.maven.plugin.MojoExecutionException;
import org.apache.maven.plugin.MojoFailureException;
import org.apache.maven.plugins.annotations.LifecyclePhase;
import org.apache.maven.plugins.annotations.Mojo;
import org.apache.maven.plugins.annotations.Parameter;
import org.apache.maven.project.MavenProject;

import java.io.File;
import java.nio.file.Path;
import java.util.List;
import java.util.Map;

/**
 * Maven Plugin Mojo that acts as a pre-processor build step to generate the skills_manifest.json resource artifact.
 */
@Mojo(name = "generate-manifest", defaultPhase = LifecyclePhase.GENERATE_RESOURCES, threadSafe = true)
public class GenerateManifestMojo extends AbstractMojo {

    @Parameter(defaultValue = "${project}", readonly = true, required = true)
    private MavenProject project;

    @Parameter(property = "skillsRoot", defaultValue = "${project.basedir}")
    private File skillsRoot;

    @Parameter(property = "outputDirectory", defaultValue = "${project.build.directory}/generated-resources/skills")
    private File outputDirectory;

    @Parameter(property = "outputFilename", defaultValue = "skills_manifest.json")
    private String outputFilename;

    @Parameter(property = "roots")
    private List<String> roots;

    @Parameter(property = "filter")
    private List<String> filter;

    @Parameter(property = "dotenvPath")
    private File dotenvPath;

    @Parameter(property = "githubToken")
    private String githubToken;

    @Parameter(property = "skip", defaultValue = "false")
    private boolean skip;

    @Override
    public void execute() throws MojoExecutionException, MojoFailureException {
        if (skip) {
            getLog().info("Skipping enterprise skills manifest generation.");
            return;
        }

        try {
            getLog().info("Initializing Enterprise Skills Loader Pre-processor...");

            Path rootPath = skillsRoot != null ? skillsRoot.toPath() : SkillLoader.findRegistryRoot();
            Path outDir = outputDirectory.toPath();
            Path outFile = outDir.resolve(outputFilename);
            Path dotEnv = dotenvPath != null ? dotenvPath.toPath() : null;

            Map<String, SkillDefinition> skills;
            if (roots != null && !roots.isEmpty()) {
                getLog().info("Loading skills from qualified root URIs: " + roots);
                skills = SkillLoader.loadSkillsFromRoots(roots, filter, githubToken, dotEnv);
            } else {
                getLog().info("Scanning workspace skills from root: " + rootPath);
                skills = SkillLoader.loadAllSkills(rootPath, filter);
            }

            getLog().info("Discovered " + skills.size() + " enterprise skill definitions.");

            Path generatedFile = SkillLoader.buildSkillsManifest(rootPath, outFile);
            getLog().info("Successfully generated skills manifest at: " + generatedFile.toAbsolutePath());

            // Add generated directory as Maven resource directory
            Resource resource = new Resource();
            resource.setDirectory(outDir.toAbsolutePath().toString());
            project.addResource(resource);
            getLog().info("Added generated resource directory to Maven build lifecycle: " + outDir.toAbsolutePath());

        } catch (Exception e) {
            getLog().error("Failed to generate skills manifest", e);
            throw new MojoExecutionException("Error generating skills manifest during Maven build", e);
        }
    }

    public void setProject(MavenProject project) {
        this.project = project;
    }

    public void setSkillsRoot(File skillsRoot) {
        this.skillsRoot = skillsRoot;
    }

    public void setOutputDirectory(File outputDirectory) {
        this.outputDirectory = outputDirectory;
    }

    public void setOutputFilename(String outputFilename) {
        this.outputFilename = outputFilename;
    }

    public void setRoots(List<String> roots) {
        this.roots = roots;
    }

    public void setFilter(List<String> filter) {
        this.filter = filter;
    }

    public void setDotenvPath(File dotenvPath) {
        this.dotenvPath = dotenvPath;
    }

    public void setGithubToken(String githubToken) {
        this.githubToken = githubToken;
    }

    public void setSkip(boolean skip) {
        this.skip = skip;
    }
}
