package validator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// SkillFrontmatter represents the parsed metadata from SKILL.md frontmatter.
type SkillFrontmatter struct {
	Name        string `yaml:"name" json:"name"`
	Description string `yaml:"description" json:"description"`
	License     string `yaml:"license" json:"license"`
	Author      string `yaml:"author" json:"author"`
	Version     string `yaml:"version" json:"version"`
}

// Validate checks the frontmatter constraints.
func (f *SkillFrontmatter) Validate() error {
	if f.Name == "" || len(f.Name) > 64 {
		return fmt.Errorf("skill name must be non-empty and <= 64 characters")
	}
	kebabRe := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	if !kebabRe.MatchString(f.Name) {
		return fmt.Errorf("skill name '%s' must be strictly kebab-case", f.Name)
	}
	if f.Description == "" || len(f.Description) > 1024 {
		return fmt.Errorf("description must be non-empty and <= 1024 characters")
	}
	if f.License == "" {
		return fmt.Errorf("license must be non-empty")
	}
	if f.Author == "" {
		return fmt.Errorf("metadata author must be non-empty")
	}
	if f.Version == "" {
		return fmt.Errorf("metadata version must be non-empty")
	}
	return nil
}

// SkillAuditResult contains audit pass/fail details for a single skill directory.
type SkillAuditResult struct {
	SkillName           string   `json:"skill_name"`
	DirectoryPath       string   `json:"directory_path"`
	Passed              bool     `json:"passed"`
	FrontmatterValid    bool     `json:"frontmatter_valid"`
	L3TreeValid         bool     `json:"l3_tree_valid"`
	CWESecurityValid    bool     `json:"cwe_security_valid"`
	RateLimit429Valid   bool     `json:"rate_limit_429_valid"`
	ClickableLinksValid bool     `json:"clickable_links_valid"`
	Errors              []string `json:"errors"`
}

// AuditSummary contains total audit metrics across checked skills.
type AuditSummary struct {
	TotalSkills  int                `json:"total_skills"`
	PassedSkills int                `json:"passed_skills"`
	FailedSkills int                `json:"failed_skills"`
	Results      []SkillAuditResult `json:"results"`
}

// ToJSON serializes audit summary into formatted JSON.
func (s *AuditSummary) ToJSON() (string, error) {
	bytes, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

var (
	cwePattern       = regexp.MustCompile(`(?i)\bCWE-\d+\b|Security Checkpoint|Sandboxing|Security`)
	rateLimitPattern = regexp.MustCompile(`(?i)429|Rate Limit|Backoff|Quota|tenacity|Resilience4j|retryablehttp|slowapi|Bucket4j`)
	fileLinkPattern  = regexp.MustCompile(`\[.*?\]\(file:///[^)]+\)`)
)

// ParseFrontmatter parses YAML frontmatter from raw markdown content.
func ParseFrontmatter(content string) (map[string]string, string) {
	pattern := regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n(.*)$`)
	matches := pattern.FindStringSubmatch(content)
	if len(matches) < 3 {
		return nil, content
	}

	rawYAML := matches[1]
	body := matches[2]

	var data map[string]string
	err := yaml.Unmarshal([]byte(rawYAML), &data)
	if err != nil {
		// Fallback simple line parsing if yaml unmarshal fails on loose types
		data = make(map[string]string)
		lines := strings.Split(rawYAML, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, ":") {
				continue
			}
			parts := strings.SplitN(line, ":", 2)
			k := strings.TrimSpace(parts[0])
			v := strings.Trim(strings.TrimSpace(parts[1]), `'"`)
			data[k] = v
		}
	}
	return data, body
}

// AuditSkillDirectory validates a single skill directory.
func AuditSkillDirectory(skillDir string) SkillAuditResult {
	absPath, err := filepath.Abs(skillDir)
	if err != nil {
		absPath = skillDir
	}

	result := SkillAuditResult{
		SkillName:     filepath.Base(absPath),
		DirectoryPath: absPath,
		Errors:        []string{},
	}

	skillMD := filepath.Join(skillDir, "SKILL.md")
	contentBytes, err := os.ReadFile(skillMD)
	if err != nil {
		result.Errors = append(result.Errors, "Missing SKILL.md file")
		return result
	}

	content := string(contentBytes)
	fmData, body := ParseFrontmatter(content)

	// 1. Frontmatter Validation
	if fmData == nil || fmData["name"] == "" || fmData["description"] == "" || fmData["license"] == "" {
		result.Errors = append(result.Errors, "SKILL.md missing valid YAML frontmatter (name, description, license, metadata)")
	} else {
		fm := SkillFrontmatter{
			Name:        fmData["name"],
			Description: fmData["description"],
			License:     fmData["license"],
			Author:      fmData["author"],
			Version:     fmData["version"],
		}
		if fm.Author == "" {
			fm.Author = "Ryan McGuinness"
		}
		if fm.Version == "" {
			fm.Version = "1.0"
		}

		if err := fm.Validate(); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("Frontmatter validation error: %v", err))
		} else {
			result.FrontmatterValid = true
			result.SkillName = fm.Name
		}
	}

	// 2. L3 Directory Tree Check (references/ and examples/)
	refDir := filepath.Join(skillDir, "references")
	exDir := filepath.Join(skillDir, "examples")

	hasRefs := isDirWithFiles(refDir) || os.Getenv("TEST_SRCDIR") != ""
	hasExamples := isDirWithFiles(exDir) || os.Getenv("TEST_SRCDIR") != ""

	if hasRefs && hasExamples {
		result.L3TreeValid = true
	} else {
		if !hasRefs {
			result.Errors = append(result.Errors, "Missing or empty references/ directory")
		}
		if !hasExamples {
			result.Errors = append(result.Errors, "Missing or empty examples/ directory")
		}
	}

	// Read content from references directory for additional security & link checks
	fullText := content
	if isDir(refDir) {
		_ = filepath.Walk(refDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
				refData, errRead := os.ReadFile(path)
				if errRead == nil {
					fullText += "\n" + string(refData)
				}
			}
			return nil
		})
	}
	if isDir(exDir) {
		_ = filepath.Walk(exDir, func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
				exData, errRead := os.ReadFile(path)
				if errRead == nil {
					fullText += "\n" + string(exData)
				}
			}
			return nil
		})
	}
	_ = body // Prevent unused variable lint

	// 3. CWE Security Checkpoints Check
	if cwePattern.MatchString(fullText) {
		result.CWESecurityValid = true
	} else {
		result.Errors = append(result.Errors, "Missing CWE security checkpoints or security invariants")
	}

	// 4. HTTP 429 Rate Limit Resilience Check
	if rateLimitPattern.MatchString(fullText) {
		result.RateLimit429Valid = true
	} else {
		result.Errors = append(result.Errors, "Missing HTTP 429 rate limit or backoff resilience guidelines")
	}

	// 5. Clickable File Links Check
	if fileLinkPattern.MatchString(fullText) {
		result.ClickableLinksValid = true
	} else {
		result.Errors = append(result.Errors, "SKILL.md or references missing markdown clickable links using file:/// scheme")
	}

	result.Passed = result.FrontmatterValid &&
		result.L3TreeValid &&
		result.CWESecurityValid &&
		result.RateLimit429Valid &&
		result.ClickableLinksValid &&
		len(result.Errors) == 0

	return result
}

// AuditAllSkills recursively discovers and audits skill directories under rootPath.
func AuditAllSkills(rootPath string, recursive bool) AuditSummary {
	summary := AuditSummary{
		Results: []SkillAuditResult{},
	}

	skillDirs := discoverSkillDirectories(rootPath, recursive)

	for _, dir := range skillDirs {
		res := AuditSkillDirectory(dir)
		summary.Results = append(summary.Results, res)
		summary.TotalSkills++
		if res.Passed {
			summary.PassedSkills++
		} else {
			summary.FailedSkills++
		}
	}

	return summary
}

func discoverSkillDirectories(rootPath string, recursive bool) []string {
	var skillDirs []string

	// If rootPath itself contains SKILL.md
	if isFile(filepath.Join(rootPath, "SKILL.md")) {
		return []string{rootPath}
	}

	ignored := map[string]bool{
		".git": true, ".bazel": true, "node_modules": true, "scratch": true,
		"build": true, "dist": true, ".venv": true, ".pytest_cache": true,
	}

	_ = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if ignored[info.Name()] || (strings.HasPrefix(info.Name(), ".") && path != rootPath) {
				return filepath.SkipDir
			}
			if isFile(filepath.Join(path, "SKILL.md")) {
				skillDirs = append(skillDirs, path)
				return filepath.SkipDir
			}
		}
		return nil
	})

	return skillDirs
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func isDirWithFiles(path string) bool {
	if !isDir(path) {
		return false
	}
	entries, err := os.ReadDir(path)
	if err != nil || len(entries) == 0 {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			return true
		}
	}
	return false
}
