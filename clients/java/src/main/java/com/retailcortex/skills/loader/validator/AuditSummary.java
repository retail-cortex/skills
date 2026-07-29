package com.retailcortex.skills.loader.validator;

import com.fasterxml.jackson.annotation.JsonProperty;

import java.util.ArrayList;
import java.util.List;

/**
 * Summary metrics of 5-point SDLC skill audit execution.
 */
public class AuditSummary {

    @JsonProperty("total_skills")
    private int totalSkills;
    @JsonProperty("passed_skills")
    private int passedSkills;
    @JsonProperty("failed_skills")
    private int failedSkills;
    private List<SkillAuditResult> results = new ArrayList<>();

    public AuditSummary() {
    }

    public int getTotalSkills() {
        return totalSkills;
    }

    public void setTotalSkills(int totalSkills) {
        this.totalSkills = totalSkills;
    }

    public int getPassedSkills() {
        return passedSkills;
    }

    public void setPassedSkills(int passedSkills) {
        this.passedSkills = passedSkills;
    }

    public int getFailedSkills() {
        return failedSkills;
    }

    public void setFailedSkills(int failedSkills) {
        this.failedSkills = failedSkills;
    }

    public List<SkillAuditResult> getResults() {
        return results;
    }

    public void setResults(List<SkillAuditResult> results) {
        this.results = results;
    }
}
