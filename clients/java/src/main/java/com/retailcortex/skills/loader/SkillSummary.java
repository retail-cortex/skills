package com.retailcortex.skills.loader;

import com.fasterxml.jackson.annotation.JsonProperty;

/**
 * High-level summary of a registered skill.
 */
public class SkillSummary {

    private String name;
    private String description;
    @JsonProperty("reference_count")
    private int referenceCount;
    @JsonProperty("example_count")
    private int exampleCount;
    private String path;

    public SkillSummary() {
    }

    public SkillSummary(String name, String description, int referenceCount, int exampleCount, String path) {
        this.name = name;
        this.description = description;
        this.referenceCount = referenceCount;
        this.exampleCount = exampleCount;
        this.path = path;
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

    public int getReferenceCount() {
        return referenceCount;
    }

    public void setReferenceCount(int referenceCount) {
        this.referenceCount = referenceCount;
    }

    public int getExampleCount() {
        return exampleCount;
    }

    public void setExampleCount(int exampleCount) {
        this.exampleCount = exampleCount;
    }

    public String getPath() {
        return path;
    }

    public void setPath(String path) {
        this.path = path;
    }
}
