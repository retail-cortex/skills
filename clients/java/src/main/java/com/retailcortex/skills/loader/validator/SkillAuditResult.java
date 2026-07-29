package com.retailcortex.skills.loader.validator;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.ArrayList;
import java.util.List;

/**
 * Audit result evaluation for a single enterprise skill definition directory.
 */
public class SkillAuditResult {

    @JsonProperty("skill_name")
    private String skillName;
    @JsonProperty("directory_path")
    private String directoryPath;
    @JsonProperty("frontmatter_valid")
    private boolean frontmatterValid;
    @JsonProperty("l3_tree_valid")
    private boolean l3TreeValid;
    @JsonProperty("cwe_security_valid")
    private boolean cweSecurityValid;
    @JsonProperty("rate_limit_429_valid")
    private boolean rateLimit429Valid;
    @JsonProperty("clickable_links_valid")
    private boolean clickableLinksValid;
    private List<String> errors = new ArrayList<>();

    public SkillAuditResult() {
    }

    public SkillAuditResult(String skillName, String directoryPath) {
        this.skillName = skillName;
        this.directoryPath = directoryPath;
    }

    @JsonProperty("passed")
    public boolean isPassed() {
        return frontmatterValid
                && l3TreeValid
                && cweSecurityValid
                && rateLimit429Valid
                && clickableLinksValid
                && errors.isEmpty();
    }

    public String getSkillName() {
        return skillName;
    }

    public void setSkillName(String skillName) {
        this.skillName = skillName;
    }

    public String getDirectoryPath() {
        return directoryPath;
    }

    public void setDirectoryPath(String directoryPath) {
        this.directoryPath = directoryPath;
    }

    public boolean isFrontmatterValid() {
        return frontmatterValid;
    }

    public void setFrontmatterValid(boolean frontmatterValid) {
        this.frontmatterValid = frontmatterValid;
    }

    public boolean isL3TreeValid() {
        return l3TreeValid;
    }

    public void setL3TreeValid(boolean l3TreeValid) {
        this.l3TreeValid = l3TreeValid;
    }

    public boolean isCweSecurityValid() {
        return cweSecurityValid;
    }

    public void setCweSecurityValid(boolean cweSecurityValid) {
        this.cweSecurityValid = cweSecurityValid;
    }

    public boolean isRateLimit429Valid() {
        return rateLimit429Valid;
    }

    public void setRateLimit429Valid(boolean rateLimit429Valid) {
        this.rateLimit429Valid = rateLimit429Valid;
    }

    public boolean isClickableLinksValid() {
        return clickableLinksValid;
    }

    public void setClickableLinksValid(boolean clickableLinksValid) {
        this.clickableLinksValid = clickableLinksValid;
    }

    public List<String> getErrors() {
        return errors;
    }

    public void setErrors(List<String> errors) {
        this.errors = errors;
    }
}
