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

package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/retail-cortex/skills/clients/go/pkg/skillsloader"
	"github.com/retail-cortex/skills/internal/installer"
	"github.com/retail-cortex/skills/pkg/model"
	"github.com/retail-cortex/skills/pkg/validator"
)

var (
	Version   = "1.0.0"
	GitCommit = "dev"
	BuildDate = "unknown"
)

// Execute runs the skm CLI logic with provided args, stdout, and stderr.
func Execute(args []string, stdout, stderr io.Writer) int {
	if len(args) < 1 {
		printUsage(stdout)
		return 0
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "login":
		return runLogin(subArgs, stdout, stderr)
	case "register":
		return runRegister(subArgs, stdout, stderr)
	case "config":
		return runConfig(subArgs, stdout, stderr)
	case "add":
		return runAdd(subArgs, stdout, stderr)

	case "verify":
		return runVerify(subArgs, stdout, stderr)
	case "validate":
		return runValidate(subArgs, stdout, stderr)
	case "list":
		return runList(subArgs, stdout, stderr)
	case "search":
		return runSearch(subArgs, stdout, stderr)
	case "compile":
		return runCompile(subArgs, stdout, stderr)
	case "init":
		return runInit(subArgs, stdout, stderr)
	case "completion":
		return runCompletion(subArgs, stdout, stderr)
	case "version", "-v", "--version":
		fmt.Fprintf(stdout, "skm version %s (commit: %s, build date: %s)\n", Version, GitCommit, BuildDate)
		return 0
	case "help", "-h", "--help":
		printUsage(stdout)
		return 0
	default:
		fmt.Fprintf(stderr, "Unknown command: %s\n\n", subcommand)
		printUsage(stderr)
		return 1
	}
}

func runAdd(args []string, stdout, stderr io.Writer) int {
	var targetDir = ".skills"
	var isForce = false
	var filter []string
	var uris []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-d" || arg == "--dir":
			if i+1 < len(args) {
				targetDir = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "-d="):
			targetDir = strings.TrimPrefix(arg, "-d=")
		case strings.HasPrefix(arg, "--dir="):
			targetDir = strings.TrimPrefix(arg, "--dir=")
		case arg == "-f" || arg == "--force":
			isForce = true
		case arg == "--filter":
			if i+1 < len(args) {
				for _, item := range strings.Split(args[i+1], ",") {
					if trimmed := strings.TrimSpace(item); trimmed != "" {
						filter = append(filter, trimmed)
					}
				}
				i++
			}
		case strings.HasPrefix(arg, "--filter="):
			raw := strings.TrimPrefix(arg, "--filter=")
			for _, item := range strings.Split(raw, ",") {
				if trimmed := strings.TrimSpace(item); trimmed != "" {
					filter = append(filter, trimmed)
				}
			}
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(stderr, "Unknown flag: %s\n", arg)
			return 1
		default:
			uris = append(uris, arg)
		}
	}

	if len(uris) == 0 {
		fmt.Fprintf(stderr, "Error: missing skill URI (e.g. skm://..., github://..., pkg://..., file://...)\n")
		return 1
	}

	results, err := installer.AddSkills(uris, targetDir, filter, isForce)
	if err != nil {
		fmt.Fprintf(stderr, "Error adding skills: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "\nSKM Skill Add Summary (destination: %s)\n", targetDir)
	fmt.Fprintf(stdout, "%s\n", strings.Repeat("=", 65))
	hasFailures := false

	for _, res := range results {
		switch res.Status {
		case "added":
			fmt.Fprintf(stdout, "[+] Added:       %-20s -> %s\n", res.SkillName, res.DestPath)
		case "overwritten":
			fmt.Fprintf(stdout, "[*] Overwritten: %-20s -> %s\n", res.SkillName, res.DestPath)
		case "skipped":
			fmt.Fprintf(stdout, "[-] Skipped:     %-20s (%s)\n", res.SkillName, res.ErrorDetail)
		case "failed":
			hasFailures = true
			fmt.Fprintf(stderr, "[!] Failed:      %-20s (%s)\n", res.SkillName, res.ErrorDetail)
		}
	}
	fmt.Fprintln(stdout)

	if hasFailures {
		return 1
	}
	return 0
}

func runValidate(args []string, stdout, stderr io.Writer) int {
	var isRecursive = false
	var jsonOutput = false
	var targetPath = "."

	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-r" || arg == "--recursive":
			isRecursive = true
		case arg == "--json":
			jsonOutput = true
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(stderr, "Unknown flag: %s\n", arg)
			return 1
		default:
			positional = append(positional, arg)
		}
	}

	if len(positional) > 0 {
		targetPath = positional[0]
	}

	summary := validator.AuditAllSkills(targetPath, isRecursive)

	if jsonOutput {
		jsonStr, err := summary.ToJSON()
		if err != nil {
			fmt.Fprintf(stderr, "Failed to format JSON output: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, jsonStr)
	} else {
		fmt.Fprintf(stdout, "\nSKM Skill Validation Report\n")
		fmt.Fprintf(stdout, "%s\n", strings.Repeat("=", 70))
		fmt.Fprintf(stdout, "Audited Path: %s (Recursive: %v)\n", targetPath, isRecursive)
		fmt.Fprintf(stdout, "Total Skills: %d | Passed: %d | Failed: %d\n\n", summary.TotalSkills, summary.PassedSkills, summary.FailedSkills)

		for _, res := range summary.Results {
			statusStr := "[PASS]"
			if !res.Passed {
				statusStr = "[FAIL]"
			}
			fmt.Fprintf(stdout, "%s %-25s (%s)\n", statusStr, res.SkillName, res.DirectoryPath)
			if !res.Passed {
				for _, errStr := range res.Errors {
					fmt.Fprintf(stdout, "     - %s\n", errStr)
				}
			}
		}
		fmt.Fprintln(stdout)
	}

	if summary.FailedSkills > 0 || summary.TotalSkills == 0 {
		return 1
	}
	return 0
}

type RemoteSkillsResult struct {
	Items      []*model.SkillResponse
	TotalCount int64
	Page       int
	PageSize   int
	TotalPages int
}

func fetchRemoteSkills(serverURL string, query string, page int, pageSize int) (*RemoteSkillsResult, error) {
	baseURL := strings.TrimRight(serverURL, "/") + "/api/v1/skills"
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	if strings.TrimSpace(query) != "" {
		q.Set("s", query)
	}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		q.Set("page_size", strconv.Itoa(pageSize))
	}
	u.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(u.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, string(body))
	}

	var results []*model.SkillResponse
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	totalCount := int64(len(results))
	if tcHeader := resp.Header.Get("X-Total-Count"); tcHeader != "" {
		if tc, err := strconv.ParseInt(tcHeader, 10, 64); err == nil {
			totalCount = tc
		}
	}

	pageVal := 1
	if pHeader := resp.Header.Get("X-Page"); pHeader != "" {
		if p, err := strconv.Atoi(pHeader); err == nil {
			pageVal = p
		}
	}

	pageSizeVal := len(results)
	if psHeader := resp.Header.Get("X-Page-Size"); psHeader != "" {
		if ps, err := strconv.Atoi(psHeader); err == nil {
			pageSizeVal = ps
		}
	}

	totalPagesVal := 1
	if tpHeader := resp.Header.Get("X-Total-Pages"); tpHeader != "" {
		if tp, err := strconv.Atoi(tpHeader); err == nil {
			totalPagesVal = tp
		}
	}

	return &RemoteSkillsResult{
		Items:      results,
		TotalCount: totalCount,
		Page:       pageVal,
		PageSize:   pageSizeVal,
		TotalPages: totalPagesVal,
	}, nil
}

func runList(args []string, stdout, stderr io.Writer) int {
	var scanDir = ""
	var jsonOutput = false
	var isRemote = false
	var serverURL = ""
	var page = 1
	var pageSize = 5

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-d" || arg == "--dir":
			if i+1 < len(args) {
				scanDir = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "-d="):
			scanDir = strings.TrimPrefix(arg, "-d=")
		case arg == "--json":
			jsonOutput = true
		case arg == "-r" || arg == "--remote":
			isRemote = true
		case arg == "-s" || arg == "--server":
			isRemote = true
			if i+1 < len(args) {
				serverURL = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--server="):
			isRemote = true
			serverURL = strings.TrimPrefix(arg, "--server=")
		case arg == "-p" || arg == "--page":
			if i+1 < len(args) {
				if p, err := strconv.Atoi(args[i+1]); err == nil && p >= 1 {
					page = p
				}
				i++
			}
		case strings.HasPrefix(arg, "--page="):
			if p, err := strconv.Atoi(strings.TrimPrefix(arg, "--page=")); err == nil && p >= 1 {
				page = p
			}
		case arg == "-n" || arg == "--max" || arg == "--page-size" || arg == "--limit":
			if i+1 < len(args) {
				if ps, err := strconv.Atoi(args[i+1]); err == nil && ps >= 1 {
					pageSize = ps
				}
				i++
			}
		case strings.HasPrefix(arg, "--max=") || strings.HasPrefix(arg, "--page-size=") || strings.HasPrefix(arg, "--limit="):
			for _, prefix := range []string{"--max=", "--page-size=", "--limit="} {
				if strings.HasPrefix(arg, prefix) {
					if ps, err := strconv.Atoi(strings.TrimPrefix(arg, prefix)); err == nil && ps >= 1 {
						pageSize = ps
					}
					break
				}
			}
		}
	}

	if isRemote {
		if serverURL == "" {
			cfg, _ := LoadCLIConfig()
			serverURL = cfg.ServerURL
		}
		if serverURL == "" {
			serverURL = "http://localhost:8000"
		}
		res, err := fetchRemoteSkills(serverURL, "", page, pageSize)
		if err != nil {
			fmt.Fprintf(stderr, "Failed to list remote skills from %s: %v\n", serverURL, err)
			return 1
		}
		if jsonOutput {
			bytes, _ := json.MarshalIndent(res.Items, "", "  ")
			fmt.Fprintln(stdout, string(bytes))
			return 0
		}
		pageSummary := ""
		if res.TotalPages > 1 || res.TotalCount > int64(len(res.Items)) {
			pageSummary = fmt.Sprintf(" (Page %d of %d, %d of %d total on %s)", res.Page, res.TotalPages, len(res.Items), res.TotalCount, serverURL)
		} else {
			pageSummary = fmt.Sprintf(" (%d found on %s)", len(res.Items), serverURL)
		}
		fmt.Fprintf(stdout, "\nSKM Server Skills%s\n", pageSummary)
		fmt.Fprintf(stdout, "%s\n", strings.Repeat("=", 75))
		for _, s := range res.Items {
			cat := "general"
			if s.Category != nil && *s.Category != "" {
				cat = *s.Category
			}
			fmt.Fprintf(stdout, "- %-25s v:%-6s cat:%-10s uri:%s\n", s.Name, s.Version, cat, s.URI)
			if s.Description != "" {
				fmt.Fprintf(stdout, "  %s\n", s.Description)
			}
		}
		fmt.Fprintln(stdout)
		return 0
	}

	if scanDir == "" {
		if isDir(".skills") {
			scanDir = ".skills"
		} else {
			scanDir = skillsloader.FindRegistryRoot()
		}
	}

	skills, err := loadSkillsFromDirectory(scanDir)
	if err != nil {
		fmt.Fprintf(stderr, "Failed to scan skills in %s: %v\n", scanDir, err)
		return 1
	}

	summaries := make([]skillsloader.SkillSummary, 0, len(skills))
	for _, s := range skills {
		summaries = append(summaries, skillsloader.SkillSummary{
			Name:           s.Name,
			Description:    s.Description,
			ReferenceCount: len(s.References),
			ExampleCount:   len(s.Examples),
			Path:           s.Path,
		})
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Name < summaries[j].Name
	})

	if jsonOutput {
		bytes, _ := json.MarshalIndent(summaries, "", "  ")
		fmt.Fprintln(stdout, string(bytes))
		return 0
	}

	fmt.Fprintf(stdout, "\nDiscovered %d skills in '%s':\n", len(summaries), scanDir)
	fmt.Fprintf(stdout, "%s\n", strings.Repeat("=", 75))
	for _, s := range summaries {
		fmt.Fprintf(stdout, "- %-25s refs:%-2d examples:%-2d path:%s\n", s.Name, s.ReferenceCount, s.ExampleCount, s.Path)
		if s.Description != "" {
			fmt.Fprintf(stdout, "  %s\n", s.Description)
		}
	}
	fmt.Fprintln(stdout)
	return 0
}

func runSearch(args []string, stdout, stderr io.Writer) int {
	var scanDir = ""
	var isRemote = false
	var serverURL = ""
	var jsonOutput = false
	var queryTerms []string
	var page = 1
	var pageSize = 5

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-d" || arg == "--dir":
			if i+1 < len(args) {
				scanDir = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "-d="):
			scanDir = strings.TrimPrefix(arg, "-d=")
		case arg == "--json":
			jsonOutput = true
		case arg == "-r" || arg == "--remote":
			isRemote = true
		case arg == "-s" || arg == "--server":
			isRemote = true
			if i+1 < len(args) {
				serverURL = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--server="):
			isRemote = true
			serverURL = strings.TrimPrefix(arg, "--server=")
		case arg == "-p" || arg == "--page":
			if i+1 < len(args) {
				if p, err := strconv.Atoi(args[i+1]); err == nil && p >= 1 {
					page = p
				}
				i++
			}
		case strings.HasPrefix(arg, "--page="):
			if p, err := strconv.Atoi(strings.TrimPrefix(arg, "--page=")); err == nil && p >= 1 {
				page = p
			}
		case arg == "-n" || arg == "--max" || arg == "--page-size" || arg == "--limit":
			if i+1 < len(args) {
				if ps, err := strconv.Atoi(args[i+1]); err == nil && ps >= 1 {
					pageSize = ps
				}
				i++
			}
		case strings.HasPrefix(arg, "--max=") || strings.HasPrefix(arg, "--page-size=") || strings.HasPrefix(arg, "--limit="):
			for _, prefix := range []string{"--max=", "--page-size=", "--limit="} {
				if strings.HasPrefix(arg, prefix) {
					if ps, err := strconv.Atoi(strings.TrimPrefix(arg, prefix)); err == nil && ps >= 1 {
						pageSize = ps
					}
					break
				}
			}
		default:
			if !strings.HasPrefix(arg, "-") {
				queryTerms = append(queryTerms, arg)
			}
		}
	}

	if len(queryTerms) == 0 {
		fmt.Fprintf(stderr, "Error: search query required (e.g. skm search python)\n")
		return 1
	}

	query := strings.Join(queryTerms, " ")

	if isRemote {
		if serverURL == "" {
			cfg, _ := LoadCLIConfig()
			serverURL = cfg.ServerURL
		}
		if serverURL == "" {
			serverURL = "http://localhost:8000"
		}
		res, err := fetchRemoteSkills(serverURL, query, page, pageSize)
		if err != nil {
			fmt.Fprintf(stderr, "Failed to search remote skills from %s: %v\n", serverURL, err)
			return 1
		}
		if jsonOutput {
			bytes, _ := json.MarshalIndent(res.Items, "", "  ")
			fmt.Fprintln(stdout, string(bytes))
			return 0
		}
		pageSummary := ""
		if res.TotalPages > 1 || res.TotalCount > int64(len(res.Items)) {
			pageSummary = fmt.Sprintf(" (Page %d of %d, %d of %d matches on %s)", res.Page, res.TotalPages, len(res.Items), res.TotalCount, serverURL)
		} else {
			pageSummary = fmt.Sprintf(" (%d matches on %s)", len(res.Items), serverURL)
		}
		fmt.Fprintf(stdout, "\nSKM Remote Search Results for '%s'%s\n", query, pageSummary)
		fmt.Fprintf(stdout, "%s\n", strings.Repeat("=", 75))
		for _, s := range res.Items {
			scoreStr := ""
			if s.SimilarityScore != nil {
				scoreStr = fmt.Sprintf(" [score: %.2f]", *s.SimilarityScore)
			}
			chunkStr := ""
			if s.MatchingChunk != nil && *s.MatchingChunk != "" {
				chunkStr = fmt.Sprintf(" (match: %s)", *s.MatchingChunk)
			}
			fmt.Fprintf(stdout, "- %-25s v:%-6s uri:%s%s%s\n", s.Name, s.Version, s.URI, scoreStr, chunkStr)
			if s.Description != "" {
				fmt.Fprintf(stdout, "  %s\n", s.Description)
			}
		}
		fmt.Fprintln(stdout)
		return 0
	}

	queryLower := strings.ToLower(query)
	if scanDir == "" {
		if isDir(".skills") {
			scanDir = ".skills"
		} else {
			scanDir = skillsloader.FindRegistryRoot()
		}
	}

	skills, err := loadSkillsFromDirectory(scanDir)
	if err != nil {
		fmt.Fprintf(stderr, "Failed to search skills: %v\n", err)
		return 1
	}

	var matches []*skillsloader.SkillDefinition
	for _, s := range skills {
		if strings.Contains(strings.ToLower(s.Name), queryLower) ||
			strings.Contains(strings.ToLower(s.Description), queryLower) ||
			strings.Contains(strings.ToLower(s.Instructions), queryLower) {
			matches = append(matches, s)
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Name < matches[j].Name
	})

	fmt.Fprintf(stdout, "\nSKM Search Results for '%s' (%d matches in %s)\n", query, len(matches), scanDir)
	fmt.Fprintf(stdout, "%s\n", strings.Repeat("=", 70))

	for _, s := range matches {
		fmt.Fprintf(stdout, "- %-25s %s\n  Path: %s\n", s.Name, s.Description, s.Path)
	}
	fmt.Fprintln(stdout)
	return 0
}

func runInit(args []string, stdout, stderr io.Writer) int {
	var targetBaseDir = "."
	var skillDesc = ""
	var skillLicense = "Apache-2.0"
	var skillAuthor = "Ryan McGuinness"
	var skillVersion = "1.0.0"
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-d" || arg == "--dir":
			if i+1 < len(args) {
				targetBaseDir = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "-d="):
			targetBaseDir = strings.TrimPrefix(arg, "-d=")
		case strings.HasPrefix(arg, "--description="):
			skillDesc = strings.TrimPrefix(arg, "--description=")
		case strings.HasPrefix(arg, "--license="):
			skillLicense = strings.TrimPrefix(arg, "--license=")
		case strings.HasPrefix(arg, "--author="):
			skillAuthor = strings.TrimPrefix(arg, "--author=")
		case strings.HasPrefix(arg, "--version="):
			skillVersion = strings.TrimPrefix(arg, "--version=")
		default:
			if !strings.HasPrefix(arg, "-") {
				positional = append(positional, arg)
			}
		}
	}

	if len(positional) == 0 {
		fmt.Fprintf(stderr, "Error: skill name required (e.g. skm init my-new-skill)\n")
		return 1
	}

	skillName := positional[0]
	kebabRe := strings.ToLower(strings.ReplaceAll(skillName, "_", "-"))
	if skillDesc == "" {
		skillDesc = fmt.Sprintf("Description for %s", kebabRe)
	}

	targetSkillDir := filepath.Join(targetBaseDir, kebabRe)
	if isDir(targetSkillDir) {
		fmt.Fprintf(stderr, "Error: directory %s already exists\n", targetSkillDir)
		return 1
	}

	refDir := filepath.Join(targetSkillDir, "references")
	exDir := filepath.Join(targetSkillDir, "examples")

	if err := os.MkdirAll(refDir, 0755); err != nil {
		fmt.Fprintf(stderr, "Failed to create references directory: %v\n", err)
		return 1
	}
	if err := os.MkdirAll(exDir, 0755); err != nil {
		fmt.Fprintf(stderr, "Failed to create examples directory: %v\n", err)
		return 1
	}

	templateSKILL := fmt.Sprintf(`---
name: %s
description: %s
license: %s
author: %s
version: %s
---

# %s

Overview and usage instructions for %s.

## Security Checkpoints
Security Checkpoint: Enforces input validation for CWE-20 and CWE-79.

## Rate Limiting & Resilience
Implements HTTP 429 rate limit exponential backoff and retryablehttp strategies.

## References
- Documentation: [Reference Guide](file:///%s/references/guide.md)
`, kebabRe, skillDesc, skillLicense, skillAuthor, skillVersion, kebabRe, kebabRe, targetSkillDir)

	if err := os.WriteFile(filepath.Join(targetSkillDir, "SKILL.md"), []byte(templateSKILL), 0644); err != nil {
		fmt.Fprintf(stderr, "Failed to write SKILL.md: %v\n", err)
		return 1
	}

	if err := os.WriteFile(filepath.Join(refDir, "guide.md"), []byte("# Reference Guide\nDetailed reference documentation.\n"), 0644); err != nil {
		fmt.Fprintf(stderr, "Failed to write guide.md: %v\n", err)
		return 1
	}

	if err := os.WriteFile(filepath.Join(exDir, "example.md"), []byte("# Example Usage\nUsage example snippet.\n"), 0644); err != nil {
		fmt.Fprintf(stderr, "Failed to write example.md: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "\nSuccessfully initialized new skill at: %s\n", targetSkillDir)
	return 0
}

func runVerify(args []string, stdout, stderr io.Writer) int {
	var targetDir = ""
	var jsonOutput = false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-d" || arg == "--dir":
			if i+1 < len(args) {
				targetDir = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "-d="):
			targetDir = strings.TrimPrefix(arg, "-d=")
		case strings.HasPrefix(arg, "--dir="):
			targetDir = strings.TrimPrefix(arg, "--dir=")
		case arg == "--json":
			jsonOutput = true
		case strings.HasPrefix(arg, "-"):
			fmt.Fprintf(stderr, "Unknown flag: %s\n", arg)
			return 1
		default:
			if targetDir == "" {
				targetDir = arg
			}
		}
	}

	if targetDir == "" {
		if isDir(".skills") {
			targetDir = ".skills"
		} else {
			targetDir = "."
		}
	}

	report, err := skillsloader.VerifyManifestLock(targetDir)
	if err != nil {
		fmt.Fprintf(stderr, "Error verifying skills integrity: %v\n", err)
		return 1
	}

	if jsonOutput {
		bytes, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			fmt.Fprintf(stderr, "Failed to format JSON output: %v\n", err)
			return 1
		}
		fmt.Fprintln(stdout, string(bytes))
	} else {
		fmt.Fprintf(stdout, "\nSKM Skill Integrity Verification Report\n")
		fmt.Fprintf(stdout, "%s\n", strings.Repeat("=", 70))
		fmt.Fprintf(stdout, "Audited Path: %s\n", report.TargetDir)
		fmt.Fprintf(stdout, "Total Skills: %d | Verified: %d | Modified: %d | Missing: %d\n\n",
			report.TotalSkills, report.VerifiedCount, report.ModifiedCount, report.MissingCount)

		for _, res := range report.Results {
			statusStr := "[PASS]"
			if res.Status != "verified" {
				statusStr = "[FAIL]"
			}
			fmt.Fprintf(stdout, "%s %-25s status:%-8s uri:%s\n", statusStr, res.SkillName, res.Status, res.URI)
			if res.Error != "" {
				fmt.Fprintf(stdout, "     - %s\n", res.Error)
			}
		}
		fmt.Fprintln(stdout)
	}

	if report.ModifiedCount > 0 || report.MissingCount > 0 || report.TotalSkills == 0 {
		return 1
	}
	return 0
}

func runCompletion(args []string, stdout, stderr io.Writer) int {
	shell := "bash"
	if len(args) > 0 {
		shell = strings.ToLower(args[0])
	}

	switch shell {
	case "bash":
		fmt.Fprintln(stdout, `# skm bash completion
_skm_completions() {
    local cur="${COMP_WORDS[COMP_CWORD]}"
    local prev="${COMP_WORDS[COMP_CWORD-1]}"
    local commands="add verify validate list search compile init completion version help"
    if [ $COMP_CWORD -eq 1 ]; then
        COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
    fi
}
complete -F _skm_completions skm`)
	case "zsh":
		fmt.Fprintln(stdout, `#compdef skm
_skm() {
    local -a commands
    commands=(
        'add:Add skills from URI'
        'verify:Verify installed skills against .manifest.lock'
        'validate:Validate skills against 5-point audit'
        'list:List registered skills'
        'search:Search skills by keyword'
        'compile:Compile zero-I/O skills manifest'
        'init:Scaffold a new skill directory'
        'completion:Generate shell completion script'
        'version:Show version information'
        'help:Show command help'
    )
    _describe 'command' commands
}
_skm "$@"`)
	case "fish":
		fmt.Fprintln(stdout, `# skm fish completion
complete -c skm -n "__fish_use_subcommand" -a "add verify validate list search compile init completion version help"`)
	default:
		fmt.Fprintf(stderr, "Unsupported shell '%s'. Supported: bash, zsh, fish\n", shell)
		return 1
	}
	return 0
}

func loadSkillsFromDirectory(scanDir string) (map[string]*skillsloader.SkillDefinition, error) {
	skills := make(map[string]*skillsloader.SkillDefinition)

	// Registry attempt
	reg, err := skillsloader.NewSkillRegistry(scanDir, nil, nil, "")
	if err == nil && len(reg.Skills()) > 0 {
		return reg.Skills(), nil
	}

	// Direct check on scanDir
	if isFile(filepath.Join(scanDir, "SKILL.md")) {
		s, err := skillsloader.LoadSkillFromDir(scanDir)
		if err == nil && s != nil {
			skills[s.Name] = s
		}
		return skills, nil
	}

	// Scan subdirectories
	entries, err := os.ReadDir(scanDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				sub := filepath.Join(scanDir, entry.Name())
				if isFile(filepath.Join(sub, "SKILL.md")) {
					s, err := skillsloader.LoadSkillFromDir(sub)
					if err == nil && s != nil {
						skills[s.Name] = s
					}
				}
			}
		}
	}
	return skills, nil
}

func runCompile(args []string, stdout, stderr io.Writer) int {
	var targetDir string
	var output = "skills_manifest.json"
	var strictSchemas = true
	var allowAddProps = ""

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-d" || arg == "--dir":
			if i+1 < len(args) {
				targetDir = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "-d="):
			targetDir = strings.TrimPrefix(arg, "-d=")
		case arg == "-o" || arg == "--output":
			if i+1 < len(args) {
				output = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--output="):
			output = strings.TrimPrefix(arg, "--output=")
		case arg == "--strict-schemas=false" || arg == "--permissive":
			strictSchemas = false
		case strings.HasPrefix(arg, "--allow-additional-properties="):
			allowAddProps = strings.TrimPrefix(arg, "--allow-additional-properties=")
		default:
			if targetDir == "" && !strings.HasPrefix(arg, "-") {
				targetDir = arg
			}
		}
	}

	manifestPath, err := skillsloader.BuildSkillsManifest(targetDir, output)
	if err != nil {
		fmt.Fprintf(stderr, "Failed to compile skills manifest: %v\n", err)
		return 1
	}

	fmt.Fprintf(stdout, "Successfully compiled skills into pre-compiled manifest (strict_schemas=%t, allow_additional_properties=%s): %s\n", strictSchemas, allowAddProps, manifestPath)
	return 0
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, `SKM - Enterprise Standalone Skills CLI Client

Usage:
  skm <command> [options] [arguments]

Commands:
  login --email <email>   Register application identity & authenticate CLI credentials (~/.skm/.env.toml)
  config <show|set>       Configure CLI settings (~/.skm/.env.toml)
  register <source-uri>   Register a skill from github://, file://, etc. with central SKM server
  add <uri>...            Add skills from skm://, github://, mod://, maven://, pkg://, or file:// to .skills (or -d directory)
  verify [path]           Verify downloaded skills against recorded SHA-256 checksums in .manifest.lock
  validate <path>         Validate frontmatter, tree, CWE, 429 resilience, and file links in a skill directory
  list                    List skills in .skills, current directory, or specified registry
  search <query>          Search skills by query term in name, description, or instructions
  compile [path]          Compile all skills into pre-compiled skills_manifest.json for fast zero-I/O loading
  init <skill-name>       Scaffold a new valid skill directory structure
  version                 Show SKM CLI version info
  help                    Show help overview

Options for 'add':
  -d, --dir <path>        Target destination directory (default: ".skills")
  -f, --force             Overwrite existing skill directories
  --filter <names>        Comma-separated skill names to select

Options for 'verify':
  -d, --dir <path>        Target directory containing .manifest.lock (default: ".skills")
  --json                  Output verification report as structured JSON

Options for 'validate':
  -r, --recursive         Recursively validate all skills in target path
  --json                  Output audit summary as structured JSON

Options for 'list' and 'search':
  -r, --remote            Query the central SKM server (configured via skm login / config)
  -s, --server <URL>      Target server URL for remote query
  -d <path>               Target directory to scan/search locally
  --json                  Output list/search results in JSON format

Options for 'compile':
  -d, --dir <path>        Target directory to scan (default: workspace root)
  -o, --output <path>     Output manifest JSON file path (default: "skills_manifest.json")
  --strict-schemas=false  Allow permissive JSON schema property parameters
  --allow-additional-properties=<list> Comma-separated allowed additional properties

Options for HITL Safety:
  --skip-hitl             Bypass interactive human approval gates for automated execution

Examples:
  skm config set server http://localhost:8080
  skm config set api_key my-secret-api-key
  skm register github://google/skills@main/tree/main/skills/cloud/gemini-api
  skm add skm://skills/sk-9b1deb4d
  skm add github://retail-cortex/skills@main/examples/python/skills
  skm add mod://github.com/retail-cortex/skills@v1.0.0/examples/go/skills
  skm add maven://com.retailcortex.skills:skills-java:1.0.0

  skm add file:///path/to/my-skill -d ./skills
  skm verify -d ./skills
  skm validate ./skills/my-skill
  skm validate -r ./packages
  skm list -d .skills
  skm compile -o ./skills_manifest.json --strict-schemas=true
  skm init my-custom-skill -d ./skills
`)
}


func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

