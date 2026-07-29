package com.retailcortex.skills.loader;

import com.fasterxml.jackson.annotation.JsonInclude;
import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.Collections;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Objects;
import java.util.stream.Collectors;

/**
 * Represents a loaded enterprise skill definition.
 */
@JsonInclude(JsonInclude.Include.NON_NULL)
public class SkillDefinition {

    private String name;
    private String description;
    private String instructions;
    private String license;
    private String author;
    private String version;
    private String compatibility;
    @JsonProperty("allowed_tools")
    private String allowedTools;
    private Map<String, String> metadata = new HashMap<>();
    private Map<String, String> references = new HashMap<>();
    private Map<String, String> examples = new HashMap<>();
    private String path = "";

    public SkillDefinition() {
    }

    public SkillDefinition(String name, String description, String instructions, String license,
                           String author, String version, String compatibility, Map<String, String> metadata,
                           Map<String, String> references, Map<String, String> examples, String path) {
        this(name, description, instructions, license, author, version, compatibility, null, metadata, references, examples, path);
    }

    public SkillDefinition(String name, String description, String instructions, String license,
                           String author, String version, String compatibility, String allowedTools,
                           Map<String, String> metadata, Map<String, String> references, Map<String, String> examples, String path) {
        this.name = name;
        this.description = description;
        this.instructions = instructions;
        this.license = license;
        this.author = author;
        this.version = version;
        this.compatibility = compatibility;
        this.allowedTools = allowedTools;
        this.metadata = metadata != null ? metadata : new HashMap<>();
        this.references = references != null ? references : new HashMap<>();
        this.examples = examples != null ? examples : new HashMap<>();
        this.path = path != null ? path : "";
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getDescription() {
        return description;
    }

    public void setDescription(String description) {
        this.description = description;
    }

    public String getInstructions() {
        return instructions;
    }

    public void setInstructions(String instructions) {
        this.instructions = instructions;
    }

    public String getLicense() {
        return license;
    }

    public void setLicense(String license) {
        this.license = license;
    }

    public String getAuthor() {
        return author;
    }

    public void setAuthor(String author) {
        this.author = author;
    }

    public String getVersion() {
        return version;
    }

    public void setVersion(String version) {
        this.version = version;
    }

    public String getCompatibility() {
        return compatibility;
    }

    public void setCompatibility(String compatibility) {
        this.compatibility = compatibility;
    }

    public String getAllowedTools() {
        return allowedTools;
    }

    public void setAllowedTools(String allowedTools) {
        this.allowedTools = allowedTools;
    }

    public String getReferenceContent(String refName) {
        return references != null ? references.get(refName) : null;
    }

    public String getExampleContent(String exName) {
        return examples != null ? examples.get(exName) : null;
    }

    public Map<String, String> getMetadata() {
        return metadata;
    }

    public void setMetadata(Map<String, String> metadata) {
        this.metadata = metadata;
    }

    public Map<String, String> getReferences() {
        return references;
    }

    public void setReferences(Map<String, String> references) {
        this.references = references;
    }

    public Map<String, String> getExamples() {
        return examples;
    }

    public void setExamples(Map<String, String> examples) {
        this.examples = examples;
    }

    public String getPath() {
        return path;
    }

    public void setPath(String path) {
        this.path = path;
    }

    /**
     * Serializes skill definition to dictionary representation matching Python/Go clients.
     */
    public Map<String, Object> toMap() {
        List<String> refKeys = references.keySet().stream().sorted().collect(Collectors.toList());
        List<String> exKeys = examples.keySet().stream().sorted().collect(Collectors.toList());

        Map<String, Object> map = new HashMap<>();
        map.put("name", name);
        map.put("description", description);
        map.put("instructions", instructions);
        map.put("license", license);
        map.put("author", author);
        map.put("version", version);
        map.put("compatibility", compatibility);
        map.put("allowed_tools", allowedTools);
        map.put("metadata", metadata != null ? metadata : Collections.emptyMap());
        map.put("references", refKeys);
        map.put("examples", exKeys);
        map.put("path", path);
        return map;
    }

    @Override
    public boolean equals(Object o) {
        if (this == o) return true;
        if (o == null || getClass() != o.getClass()) return false;
        SkillDefinition that = (SkillDefinition) o;
        return Objects.equals(name, that.name) && Objects.equals(path, that.path);
    }

    @Override
    public int hashCode() {
        return Objects.hash(name, path);
    }
}
