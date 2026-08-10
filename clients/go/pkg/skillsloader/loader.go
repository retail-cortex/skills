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

package skillsloader

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

// FindRegistryRoot discovers the root workspace directory containing enterprise skill packages.
func FindRegistryRoot() string {
	if ws := os.Getenv("BUILD_WORKSPACE_DIRECTORY"); ws != "" {
		if isDir(filepath.Join(ws, "packages")) || isDir(filepath.Join(ws, "skills")) || isDir(filepath.Join(ws, "examples")) {
			return ws
		}
	}

	if testSrcDir := os.Getenv("TEST_SRCDIR"); testSrcDir != "" {
		wsName := os.Getenv("TEST_WORKSPACE")
		candidates := []string{}
		if wsName != "" {
			candidates = append(candidates, filepath.Join(testSrcDir, wsName))
		}
		candidates = append(candidates,
			filepath.Join(testSrcDir, "_main"),
			filepath.Join(testSrcDir, "skill_builder"),
			testSrcDir,
		)

		for _, cand := range candidates {
			if isDir(filepath.Join(cand, "packages")) || isDir(filepath.Join(cand, "skills")) || isDir(filepath.Join(cand, "examples")) {
				return cand
			}
		}

		// Search for any SKILL.md under TEST_SRCDIR
		var foundRoot string
		_ = filepath.Walk(testSrcDir, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			if info.Name() == "SKILL.md" {
				dir := filepath.Dir(path)
				for dir != "." && dir != "/" {
					if isDir(filepath.Join(dir, "packages")) || isDir(filepath.Join(dir, "skills")) || isDir(filepath.Join(dir, "examples")) {
						foundRoot = dir
						return fmt.Errorf("found")
					}
					parent := filepath.Dir(dir)
					if parent == dir {
						break
					}
					dir = parent
				}
			}
			return nil
		})
		if foundRoot != "" {
			return foundRoot
		}
	}

	// Parent walk fallback from current directory
	cwd, err := os.Getwd()
	if err == nil {
		curr := cwd
		for {
			if isDir(filepath.Join(curr, "packages")) || isDir(filepath.Join(curr, "skills")) || isDir(filepath.Join(curr, "examples")) {
				return curr
			}
			parent := filepath.Dir(curr)
			if parent == curr {
				break
			}
			curr = parent
		}
	}

	return cwd
}

// GetLoaderSkillsDir returns the persistent workspace directory for cached downloaded skill trees.
func GetLoaderSkillsDir() string {
	loaderDir := filepath.Join(FindRegistryRoot(), ".loader_skills")
	_ = os.MkdirAll(loaderDir, 0755)
	return loaderDir
}

// ParseFrontmatter parses YAML frontmatter block from SKILL.md content.
func ParseFrontmatter(content string) (map[string]string, string) {
	re := regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n(.*)$`)
	matches := re.FindStringSubmatch(content)
	if len(matches) < 3 {
		return map[string]string{}, content
	}

	yamlText := matches[1]
	body := matches[2]

	data := make(map[string]string)
	var rawMap map[string]any
	if err := yaml.Unmarshal([]byte(yamlText), &rawMap); err == nil {
		for k, v := range rawMap {
			data[k] = fmt.Sprintf("%v", v)
		}
	} else {
		// Line-based fallback
		lines := strings.Split(yamlText, "\n")
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

// ParseDotenvFile parses key-value environment variables from a .env or dotenv configuration file.
func ParseDotenvFile(envPath string) map[string]string {
	envVars := make(map[string]string)
	content, err := os.ReadFile(envPath)
	if err != nil {
		return envVars
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		k := strings.TrimSpace(parts[0])
		v := strings.Trim(strings.TrimSpace(parts[1]), `'"`)
		envVars[k] = v
	}
	return envVars
}

// ParseSkillRootURI parses a qualified skill root URI into (scheme, target, ref, subpath).
func ParseSkillRootURI(uri string) (scheme, target, ref, subpath string) {
	clean := strings.TrimSpace(uri)
	if strings.HasPrefix(clean, "skm://") || strings.HasPrefix(clean, "skms://") {
		prefix := "skm://"
		if strings.HasPrefix(clean, "skms://") {
			prefix = "skms://"
		}
		raw := clean[len(prefix):]
		raw = strings.TrimPrefix(raw, "skills/")
		return "skm", raw, "", ""
	}
	if strings.HasPrefix(clean, "file://") {
		return "file", clean[len("file://"):], "", ""
	}
	if strings.HasPrefix(clean, "pkg://") || strings.HasPrefix(clean, "package://") {
		prefix := "pkg://"
		if strings.HasPrefix(clean, "package://") {
			prefix = "package://"
		}
		return "pkg", clean[len(prefix):], "", ""
	}


	if strings.HasPrefix(clean, "mod://") || strings.HasPrefix(clean, "go://") {
		prefix := "mod://"
		if strings.HasPrefix(clean, "go://") {
			prefix = "go://"
		}
		raw := clean[len(prefix):]
		var target, ref, subpath string
		if strings.Contains(raw, "@") {
			parts := strings.SplitN(raw, "@", 2)
			target = parts[0]
			refSub := parts[1]
			if strings.Contains(refSub, "/") {
				subParts := strings.SplitN(refSub, "/", 2)
				ref = subParts[0]
				subpath = subParts[1]
			} else {
				ref = refSub
			}
		} else if strings.Contains(raw, "/") {
			parts := strings.Split(raw, "/")
			if len(parts) >= 3 {
				target = strings.Join(parts[:3], "/")
				if len(parts) > 3 {
					subpath = strings.Join(parts[3:], "/")
				}
			} else {
				target = raw
			}
		} else {
			target = raw
		}
		return "mod", target, ref, subpath
	}

	if strings.HasPrefix(clean, "maven://") || strings.HasPrefix(clean, "mvn://") {
		prefix := "maven://"
		if strings.HasPrefix(clean, "mvn://") {
			prefix = "mvn://"
		}
		raw := clean[len(prefix):]
		if idx := strings.Index(raw, "/"); idx != -1 {
			target = raw[:idx]
			subpath = raw[idx+1:]
		} else {
			target = raw
		}
		if strings.Contains(target, ":") {
			parts := strings.Split(target, ":")
			if len(parts) >= 3 {
				ref = parts[2]
			}
		}
		return "maven", target, ref, subpath
	}

	if strings.HasPrefix(clean, "github://") || strings.HasPrefix(clean, "https://github.com/") {

		prefix := "github://"
		if strings.HasPrefix(clean, "https://github.com/") {
			prefix = "https://github.com/"
		}
		target = strings.TrimSuffix(clean[len(prefix):], "/")

		if strings.Contains(target, "/tree/") {
			parts := strings.SplitN(target, "/tree/", 2)
			target = parts[0]
			treeBits := strings.SplitN(parts[1], "/", 2)
			ref = treeBits[0]
			if len(treeBits) > 1 {
				subpath = treeBits[1]
			}
		} else if strings.Contains(target, ":") {
			part1, part2 := rsplitOnce(target, ":")
			if !strings.Contains(part2, "/") {
				ref = part2
				target = part1
			} else {
				parts := strings.SplitN(target, ":", 2)
				target = parts[0]
				refSub := parts[1]
				if strings.Contains(refSub, "/") {
					parts2 := strings.SplitN(refSub, "/", 2)
					ref = parts2[0]
					subpath = parts2[1]
				} else {
					ref = refSub
				}
			}
		} else if strings.Contains(target, "@") {
			parts := strings.SplitN(target, "@", 2)
			target = parts[0]
			refSub := parts[1]
			if strings.Contains(refSub, "/") {
				parts2 := strings.SplitN(refSub, "/", 2)
				ref = parts2[0]
				subpath = parts2[1]
			} else {
				ref = refSub
			}
		}

		if subpath == "" {
			parts := strings.Split(target, "/")
			if len(parts) > 2 {
				target = parts[0] + "/" + parts[1]
				subpath = strings.Join(parts[2:], "/")
			}
		}

		if ref == "" {
			ref = "main"
		}
		return "github", target, ref, subpath
	}

	return "file", clean, "", ""
}

// readSkillFileContent reads file content, preserving text as-is and encoding binary files as Data URIs.
func readSkillFileContent(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	// If valid UTF-8 and does not contain raw null byte, return as plain text
	if utf8.Valid(data) && !bytes.Contains(data, []byte{0}) {
		return string(data), nil
	}
	// Fallback to Data URI with MIME detection for binary formats
	mimeType := http.DetectContentType(data)
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".pdf":
		mimeType = "application/pdf"
	case ".svg":
		mimeType = "image/svg+xml"
	case ".png":
		mimeType = "image/png"
	case ".jpg", ".jpeg":
		mimeType = "image/jpeg"
	case ".webp":
		mimeType = "image/webp"
	case ".gif":
		mimeType = "image/gif"
	case ".ico":
		mimeType = "image/x-icon"
	case ".wasm":
		mimeType = "application/wasm"
	case ".wav":
		mimeType = "audio/wav"
	case ".mp3":
		mimeType = "audio/mpeg"
	case ".zip":
		mimeType = "application/zip"
	case ".db", ".sqlite":
		mimeType = "application/x-sqlite3"
	case ".pb", ".bin":
		mimeType = "application/octet-stream"
	}
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data)), nil
}

// LoadSkillFromDir loads a single skill definition from its directory, enforcing boundary checks.
func LoadSkillFromDir(skillDir string) (*SkillDefinition, error) {
	skillMD := filepath.Join(skillDir, "SKILL.md")
	content, err := os.ReadFile(skillMD)
	if err != nil {
		return nil, fmt.Errorf("SKILL.md not found in %s: %w", skillDir, err)
	}

	fmData, body := ParseFrontmatter(string(content))
	name := fmData["name"]
	if name == "" {
		name = filepath.Base(skillDir)
	}

	desc := fmData["description"]
	if desc == "" {
		desc = fmt.Sprintf("Enterprise skill for %s", name)
	}

	allowedTools := fmData["allowed-tools"]
	if allowedTools == "" {
		allowedTools = fmData["allowed_tools"]
	}

	metaDict := make(map[string]string)
	knownKeys := map[string]bool{
		"name":              true,
		"description":       true,
		"license":           true,
		"author":            true,
		"version":           true,
		"compatibility":     true,
		"allowed-tools":     true,
		"allowed_tools":     true,
		"category":          true,
		"tags":              true,
		"trigger_phrases":   true,
		"execution_hints":   true,
		"authors":           true,
		"tool_requirements": true,
		"metadata":          true,
		"source_uri":        true,
		"src":               true,
		"source":            true,
	}

	for k, v := range fmData {
		if !knownKeys[k] && strings.TrimSpace(v) != "" {
			metaDict[k] = v
		}
	}

	author := fmData["author"]
	if author == "" {
		author = metaDict["author"]
	} else if metaDict["author"] == "" {
		metaDict["author"] = author
	}

	version := fmData["version"]
	if version == "" {
		version = metaDict["version"]
	} else if metaDict["version"] == "" {
		metaDict["version"] = version
	}

	// Parse structured YAML frontmatter for complex fields
	var structFm struct {
		Authors          []AuthorDetails   `yaml:"authors"`
		ToolRequirements []ToolRequirement `yaml:"tool_requirements"`
		Category         string            `yaml:"category"`
		Tags             []string          `yaml:"tags"`
		TriggerPhrases   []string          `yaml:"trigger_phrases"`
		ExecutionHints   *struct {
			PreferredModel        string            `yaml:"preferred_model"`
			RequiresHumanApproval bool              `yaml:"requires_human_approval"`
			EnvironmentVariables  []string          `yaml:"environment_variables"`
			TimeoutSeconds        int               `yaml:"timeout_seconds"`
			CustomHints           map[string]string `yaml:"custom_hints"`
		} `yaml:"execution_hints"`
	}

	re := regexp.MustCompile(`(?s)^---\s*\n(.*?)\n---\s*\n`)
	if m := re.FindStringSubmatch(string(content)); len(m) > 1 {
		_ = yaml.Unmarshal([]byte(m[1]), &structFm)
	}

	var execHints *ExecutionHints
	if structFm.ExecutionHints != nil {
		execHints = &ExecutionHints{
			PreferredModel:        structFm.ExecutionHints.PreferredModel,
			RequiresHumanApproval: structFm.ExecutionHints.RequiresHumanApproval,
			EnvironmentVariables:  structFm.ExecutionHints.EnvironmentVariables,
			TimeoutSeconds:        structFm.ExecutionHints.TimeoutSeconds,
			CustomHints:           structFm.ExecutionHints.CustomHints,
		}
	}

	// Compute relative path against registry root
	registryRoot := FindRegistryRoot()
	relPath := skillDir
	if r, err := filepath.Rel(registryRoot, skillDir); err == nil && !strings.HasPrefix(r, "..") {
		relPath = r
	}

	sourceURI := fmData["source_uri"]
	if sourceURI == "" {
		sourceURI = fmData["src"]
	}
	if sourceURI == "" {
		sourceURI = fmData["source"]
	}
	if sourceURI == "" {
		sourceURI = fmt.Sprintf("file://%s", relPath)
	}

	references := make(map[string]string)
	refDir := filepath.Join(skillDir, "references")
	if isDir(refDir) {
		entries, _ := os.ReadDir(refDir)
		for _, entry := range entries {
			if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				refPath := filepath.Join(refDir, entry.Name())
				if isWithinBaseDir(skillDir, refPath) {
					if refContent, err := readSkillFileContent(refPath); err == nil {
						references[entry.Name()] = refContent
					}
				}
			}
		}
	}

	examples := make(map[string]string)
	exDir := filepath.Join(skillDir, "examples")
	if isDir(exDir) {
		entries, _ := os.ReadDir(exDir)
		for _, entry := range entries {
			if !entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
				exPath := filepath.Join(exDir, entry.Name())
				if isWithinBaseDir(skillDir, exPath) {
					if exContent, err := readSkillFileContent(exPath); err == nil {
						examples[entry.Name()] = exContent
					}
				}
			}
		}
	}

	payload := fmt.Sprintf("%s:%s:%s:%s", name, version, strings.TrimSpace(body), desc)
	h := sha256.Sum256([]byte(payload))
	sha256Hash := hex.EncodeToString(h[:])

	def := &SkillDefinition{
		Name:             name,
		Description:      desc,
		Instructions:     strings.TrimSpace(body),
		License:          fmData["license"],
		Author:           author,
		Authors:          structFm.Authors,
		Version:          version,
		Compatibility:    fmData["compatibility"],
		AllowedTools:     allowedTools,
		ToolRequirements: structFm.ToolRequirements,
		Category:         structFm.Category,
		Tags:             structFm.Tags,
		TriggerPhrases:   structFm.TriggerPhrases,
		ExecutionHints:   execHints,
		Metadata:         metaDict,
		References:       references,
		Examples:         examples,
		Path:             relPath,
		SourceURI:        sourceURI,
		SHA256Hash:       sha256Hash,
	}

	return def, nil
}

// LoadAllSkills scans and loads skill directories in the workspace packages and local folders.
func LoadAllSkills(skillsRoot string, skillFilter []string) (map[string]*SkillDefinition, error) {
	root := skillsRoot
	if root == "" {
		root = FindRegistryRoot()
	}

	filterSet := make(map[string]bool)
	for _, f := range skillFilter {
		filterSet[f] = true
	}
	hasFilter := len(filterSet) > 0

	loaded := make(map[string]*SkillDefinition)

	// Check if root itself is a skill dir
	if isFile(filepath.Join(root, "SKILL.md")) {
		single, err := LoadSkillFromDir(root)
		if err == nil && single != nil {
			if !hasFilter || filterSet[single.Name] {
				return map[string]*SkillDefinition{single.Name: single}, nil
			}
		}
	}

	// 1. Scan workspace packages and examples for skills
	searchDirs := []string{
		filepath.Join(root, "packages"),
		filepath.Join(root, "examples"),
	}
	if filepath.Base(root) == "packages" || filepath.Base(root) == "examples" {
		searchDirs = []string{root}
	}
	for _, parentDir := range searchDirs {
		if isDir(parentDir) {
			pattern := filepath.Join(parentDir, "skills-*", "src", "*", "skills", "*")
			matches, _ := filepath.Glob(pattern)
			sort.Strings(matches)
			for _, skillDir := range matches {
				if isDir(skillDir) && !strings.HasPrefix(filepath.Base(skillDir), ".") {
					skillName := filepath.Base(skillDir)
					if hasFilter && !filterSet[skillName] {
						continue
					}
					skillDef, err := LoadSkillFromDir(skillDir)
					if err == nil && skillDef != nil {
						if !hasFilter || filterSet[skillDef.Name] {
							loaded[skillDef.Name] = skillDef
						}
					}
				}
			}
		}
	}

	// 2. Fallback scan for standalone skills directory (and subcategory folders)
	skillsDirs := []string{
		filepath.Join(root, "skills"),
		filepath.Join(root, "examples", "skills"),
	}
	if filepath.Base(root) == "skills" {
		skillsDirs = []string{root}
	}

	for _, sDir := range skillsDirs {
		if isDir(sDir) {
			_ = filepath.Walk(sDir, func(path string, info os.FileInfo, err error) error {
				if err == nil && !info.IsDir() && info.Name() == "SKILL.md" {
					dir := filepath.Dir(path)
					dirName := filepath.Base(dir)
					if _, exists := loaded[dirName]; !exists {
						if !hasFilter || filterSet[dirName] {
							if skillDef, err := LoadSkillFromDir(dir); err == nil && skillDef != nil {
								loaded[skillDef.Name] = skillDef
							}
						}
					}
				}
				return nil
			})
		}
	}

	// 3. Scan standard cross-client .agents/skills directories (project-level & user-level)
	var agentDirs []string
	agentDirs = append(agentDirs, filepath.Join(root, ".agents", "skills"))
	if homeDir, err := os.UserHomeDir(); err == nil && homeDir != "" {
		agentDirs = append(agentDirs, filepath.Join(homeDir, ".agents", "skills"))
	}

	for _, agDir := range agentDirs {
		if isDir(agDir) {
			entries, _ := os.ReadDir(agDir)
			for _, entry := range entries {
				if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
					if _, exists := loaded[entry.Name()]; exists {
						continue
					}
					if hasFilter && !filterSet[entry.Name()] {
						continue
					}
					skillDir := filepath.Join(agDir, entry.Name())
					skillDef, err := LoadSkillFromDir(skillDir)
					if err == nil && skillDef != nil {
						if !hasFilter || filterSet[skillDef.Name] {
							loaded[skillDef.Name] = skillDef
						}
					}
				}
			}
		}
	}

	return loaded, nil
}

// LoadSkillsFromPackage loads enterprise skills from a specified package name.
func LoadSkillsFromPackage(packageName string, skillFilter []string) (map[string]*SkillDefinition, error) {
	cleanPkg := strings.TrimSpace(packageName)
	if cleanPkg == "" {
		return map[string]*SkillDefinition{}, nil
	}

	root := FindRegistryRoot()
	searchDirs := []string{
		filepath.Join(root, "packages"),
		filepath.Join(root, "examples"),
	}
	if filepath.Base(root) == "packages" || filepath.Base(root) == "examples" {
		searchDirs = []string{root}
	}

	loaded := make(map[string]*SkillDefinition)
	for _, parentDir := range searchDirs {
		if isDir(parentDir) {
			var pats []string
			p1, _ := filepath.Glob(filepath.Join(parentDir, "skills-*"))
			p2, _ := filepath.Glob(filepath.Join(parentDir, "*", "skills"))
			p3, _ := filepath.Glob(filepath.Join(parentDir, "*"))
			pats = append(pats, p1...)
			pats = append(pats, p2...)
			pats = append(pats, p3...)
			for _, p := range pats {
				srcDir := filepath.Join(p, "src")
				if isDir(srcDir) {
					entries, _ := os.ReadDir(srcDir)
					for _, entry := range entries {
						if entry.IsDir() && entry.Name() == cleanPkg {
							pkgDir := filepath.Join(srcDir, entry.Name())
							skillsSub := filepath.Join(pkgDir, "skills")
							target := pkgDir
							if isDir(skillsSub) {
								target = skillsSub
							}
							skills, err := LoadAllSkills(target, skillFilter)
							if err == nil {
								for k, v := range skills {
									loaded[k] = v
								}
							}
						}
					}
				}
			}
		}
	}


	return loaded, nil
}

// LoadSkillsFromGoModule resolves a Go module URI and loads contained skill definitions.
func LoadSkillsFromGoModule(target, ref string, roots, filter []string) (map[string]*SkillDefinition, error) {
	cleanMod := strings.TrimSpace(target)
	if cleanMod == "" {
		return map[string]*SkillDefinition{}, nil
	}

	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		if home, err := os.UserHomeDir(); err == nil {
			gopath = filepath.Join(home, "go")
		}
	}

	var modDir string
	if ref != "" && gopath != "" {
		cand := filepath.Join(gopath, "pkg", "mod", fmt.Sprintf("%s@%s", strings.ToLower(cleanMod), ref))
		if isDir(cand) {
			modDir = cand
		}
	}

	if modDir == "" && ref != "" {
		if _, err := exec.LookPath("go"); err == nil {
			modSpec := fmt.Sprintf("%s@%s", cleanMod, ref)
			cmd := exec.Command("go", "mod", "download", "-json", modSpec)
			var out bytes.Buffer
			cmd.Stdout = &out
			if err := cmd.Run(); err == nil {
				var result struct {
					Dir string `json:"Dir"`
				}
				if err := json.Unmarshal(out.Bytes(), &result); err == nil && result.Dir != "" && isDir(result.Dir) {
					modDir = result.Dir
				}
			}
		}
	}

	if modDir == "" && gopath != "" {
		modParent := filepath.Join(gopath, "pkg", "mod", filepath.Dir(cleanMod))
		baseName := filepath.Base(cleanMod)
		if isDir(modParent) {
			entries, _ := os.ReadDir(modParent)
			for _, entry := range entries {
				if entry.IsDir() && strings.HasPrefix(entry.Name(), baseName+"@") {
					modDir = filepath.Join(modParent, entry.Name())
					break
				}
			}
		}
	}

	if modDir != "" {
		candidateDir := modDir
		if len(roots) > 0 && roots[0] != "" {
			sub := filepath.Join(modDir, roots[0])
			if isDir(sub) {
				candidateDir = sub
			}
		}
		return LoadAllSkills(candidateDir, filter)
	}

	root := FindRegistryRoot()
	artName := filepath.Base(cleanMod)
	cleanArtName := strings.TrimPrefix(artName, "skills-")
	candidates := []string{
		filepath.Join(root, "examples", cleanArtName, "skills"),
		filepath.Join(root, "examples", "skills-"+cleanArtName),
		filepath.Join(root, "examples", cleanArtName),
		filepath.Join(root, "packages", "skills-"+cleanArtName),
		filepath.Join(root, "packages", cleanArtName),
		filepath.Join(root, "clients", "go"),
	}


	for _, cand := range candidates {
		if isDir(cand) {
			skills, err := LoadAllSkills(cand, filter)
			if err == nil && len(skills) > 0 {
				return skills, nil
			}
			matches, _ := filepath.Glob(filepath.Join(cand, "src", "*", "skills", "*"))
			if len(matches) > 0 {
				found := make(map[string]*SkillDefinition)
				for _, m := range matches {
					if isDir(m) {
						if s, err := LoadSkillFromDir(m); err == nil && s != nil {
							found[s.Name] = s
						}
					}
				}
				if len(found) > 0 {
					return found, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("go module %s (ref: %s) not found in GOPATH module cache or workspace", cleanMod, ref)
}

// LoadSkillsFromMaven locates/downloads a Java Maven artifact and loads contained skill definitions.
func LoadSkillsFromMaven(target string, ref string, roots []string, filter []string) (map[string]*SkillDefinition, error) {
	parts := strings.Split(target, ":")
	var groupId, artifactId, version string
	if len(parts) >= 3 {
		groupId = parts[0]
		artifactId = parts[1]
		version = parts[2]
	} else if len(parts) == 2 {
		groupId = parts[0]
		artifactId = parts[1]
		version = ref
	} else {
		artifactId = target
		version = ref
	}

	homeDir, _ := os.UserHomeDir()
	var jarPath string
	if groupId != "" && artifactId != "" && version != "" {
		groupPath := strings.ReplaceAll(groupId, ".", "/")
		cand := filepath.Join(homeDir, ".m2", "repository", groupPath, artifactId, version, fmt.Sprintf("%s-%s.jar", artifactId, version))
		if isFile(cand) {
			jarPath = cand
		}
	}

	if jarPath == "" && groupId != "" && artifactId != "" && version != "" {
		if _, err := exec.LookPath("mvn"); err == nil {
			artifactSpec := fmt.Sprintf("%s:%s:%s", groupId, artifactId, version)
			cmd := exec.Command("mvn", "dependency:get", fmt.Sprintf("-Dartifact=%s", artifactSpec))
			_ = cmd.Run()

			groupPath := strings.ReplaceAll(groupId, ".", "/")
			cand := filepath.Join(homeDir, ".m2", "repository", groupPath, artifactId, version, fmt.Sprintf("%s-%s.jar", artifactId, version))
			if isFile(cand) {
				jarPath = cand
			}
		}
	}

	if jarPath != "" {
		cacheKey := fmt.Sprintf("%s-%s-%s", groupId, artifactId, version)
		extractedDir := filepath.Join(GetLoaderSkillsDir(), "maven", cacheKey)
		if !isDir(extractedDir) {
			if err := unzipFile(jarPath, extractedDir); err != nil {
				return nil, fmt.Errorf("failed to extract maven jar %s: %w", jarPath, err)
			}
		}

		if len(roots) > 0 && roots[0] != "" {
			sub := filepath.Join(extractedDir, roots[0])
			if isDir(sub) {
				extractedDir = sub
			}
		}
		return LoadAllSkills(extractedDir, filter)
	}

	root := FindRegistryRoot()
	cleanArtId := strings.TrimPrefix(artifactId, "skills-")
	candidates := []string{
		filepath.Join(root, "examples", cleanArtId, "skills"),
		filepath.Join(root, "examples", "skills-"+cleanArtId),
		filepath.Join(root, "examples", cleanArtId),
		filepath.Join(root, "packages", "skills-"+cleanArtId),
		filepath.Join(root, "packages", cleanArtId),
		filepath.Join(root, "clients", "java"),
	}

	for _, cand := range candidates {
		if isDir(cand) {
			skills, err := LoadAllSkills(cand, filter)
			if err == nil && len(skills) > 0 {
				return skills, nil
			}
			matches, _ := filepath.Glob(filepath.Join(cand, "src", "*", "skills", "*"))
			if len(matches) > 0 {
				found := make(map[string]*SkillDefinition)
				for _, m := range matches {
					if isDir(m) {
						if s, err := LoadSkillFromDir(m); err == nil && s != nil {
							found[s.Name] = s
						}
					}
				}
				if len(found) > 0 {
					return found, nil
				}
			}
		}
	}

	return nil, fmt.Errorf("maven artifact %s (version: %s) not found in local m2 cache or workspace", target, version)
}

func unzipFile(zipPath, dst string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dst, f.Name)
		if !isWithinBaseDir(dst, fpath) {
			continue
		}

		if f.FileInfo().IsDir() {
			_ = os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		rc.Close()
		outFile.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// LoadSkillsFromGitHub loads skills from a remote GitHub repository.

func LoadSkillsFromGitHub(repo string, ref string, roots []string, filter []string, token string, dotenvPath string) (map[string]*SkillDefinition, error) {
	envFile := dotenvPath
	if envFile == "" {
		cwd, _ := os.Getwd()
		envFile = filepath.Join(cwd, ".env")
	}
	dotenvVars := ParseDotenvFile(envFile)

	cleanRepo := strings.TrimSpace(repo)
	if strings.HasPrefix(cleanRepo, "https://github.com/") {
		cleanRepo = strings.TrimSuffix(cleanRepo[len("https://github.com/"):], "/")
	}
	if strings.HasSuffix(cleanRepo, ".git") {
		cleanRepo = cleanRepo[:len(cleanRepo)-4]
	}

	var urlRoots []string
	if strings.Contains(cleanRepo, "/tree/") {
		parts := strings.SplitN(cleanRepo, "/tree/", 2)
		cleanRepo = strings.TrimSuffix(parts[0], "/")
		treeParts := strings.SplitN(parts[1], "/", 2)
		parsedRef := treeParts[0]
		if ref == "" {
			ref = parsedRef
		}
		if len(treeParts) > 1 && strings.TrimSpace(treeParts[1]) != "" {
			urlRoots = []string{strings.TrimSpace(treeParts[1])}
		}
	}

	gitToken := token
	if gitToken == "" {
		gitToken = os.Getenv("GITHUB_TOKEN")
	}
	if gitToken == "" {
		gitToken = os.Getenv("GH_TOKEN")
	}
	if gitToken == "" {
		gitToken = dotenvVars["GITHUB_TOKEN"]
	}
	if gitToken == "" {
		gitToken = dotenvVars["GH_TOKEN"]
	}

	gitRef := ref
	if gitRef == "" {
		gitRef = os.Getenv("GITHUB_REF")
	}
	if gitRef == "" {
		gitRef = dotenvVars["GITHUB_REF"]
	}
	if gitRef == "" {
		gitRef = "main"
	}

	var rootPaths []string
	if len(roots) > 0 {
		rootPaths = roots
	} else if len(urlRoots) > 0 {
		rootPaths = urlRoots
	} else if envRoots := os.Getenv("SKILLS_ROOTS"); envRoots != "" {
		for _, r := range strings.Split(envRoots, ",") {
			if strings.TrimSpace(r) != "" {
				rootPaths = append(rootPaths, strings.TrimSpace(r))
			}
		}
	} else if dotRoots := dotenvVars["SKILLS_ROOTS"]; dotRoots != "" {
		for _, r := range strings.Split(dotRoots, ",") {
			if strings.TrimSpace(r) != "" {
				rootPaths = append(rootPaths, strings.TrimSpace(r))
			}
		}
	} else {
		rootPaths = []string{"skills", "."}
	}

	var selectedSkills []string
	if len(filter) > 0 {
		selectedSkills = filter
	} else if envFilt := os.Getenv("SKILLS_FILTER"); envFilt != "" {
		for _, s := range strings.Split(envFilt, ",") {
			if strings.TrimSpace(s) != "" {
				selectedSkills = append(selectedSkills, strings.TrimSpace(s))
			}
		}
	} else if dotFilt := dotenvVars["SKILLS_FILTER"]; dotFilt != "" {
		for _, s := range strings.Split(dotFilt, ",") {
			if strings.TrimSpace(s) != "" {
				selectedSkills = append(selectedSkills, strings.TrimSpace(s))
			}
		}
	}

	repoSlug := strings.ReplaceAll(cleanRepo, "/", "_")
	loaderBase := GetLoaderSkillsDir()
	persistentRepoDir := filepath.Join(loaderBase, "github", repoSlug, gitRef)
	_ = os.MkdirAll(persistentRepoDir, 0755)

	tmpDir, err := os.MkdirTemp("", "skills-loader-gh-*")
	if err == nil {
		defer os.RemoveAll(tmpDir)
	}

	cloned := false
	repoDir := filepath.Join(tmpDir, "repo")

	cloneURL := fmt.Sprintf("https://github.com/%s.git", cleanRepo)
	if gitToken != "" {
		cloneURL = fmt.Sprintf("https://x-access-token:%s@github.com/%s.git", gitToken, cleanRepo)
	}

	cmdClone := exec.Command("git", "clone", "--depth", "1", "--branch", gitRef, cloneURL, repoDir)
	if err := cmdClone.Run(); err == nil {
		cloned = true
	} else {
		_ = exec.Command("git", "init", repoDir).Run()
		_ = exec.Command("git", "-C", repoDir, "remote", "add", "origin", cloneURL).Run()
		if errFetch := exec.Command("git", "-C", repoDir, "fetch", "--depth", "1", "origin", gitRef).Run(); errFetch == nil {
			if errCo := exec.Command("git", "-C", repoDir, "checkout", "FETCH_HEAD").Run(); errCo == nil {
				cloned = true
			}
		}
	}

	repoTargetDir := tmpDir
	if cloned {
		repoTargetDir = repoDir
	} else {
		archiveURL := fmt.Sprintf("https://api.github.com/repos/%s/zipball/%s", cleanRepo, gitRef)
		req, _ := http.NewRequest("GET", archiveURL, nil)
		req.Header.Set("User-Agent", "skills-loader-go/1.0.0")
		if gitToken != "" {
			req.Header.Set("Authorization", "token "+gitToken)
		}
		resp, httpErr := http.DefaultClient.Do(req)
		if httpErr == nil && resp.StatusCode == http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if zipReader, zipErr := zip.NewReader(bytes.NewReader(bodyBytes), int64(len(bodyBytes))); zipErr == nil {
				for _, zipFile := range zipReader.File {
					outPath := filepath.Join(tmpDir, zipFile.Name)
					if zipFile.FileInfo().IsDir() {
						_ = os.MkdirAll(outPath, zipFile.Mode())
						continue
					}
					_ = os.MkdirAll(filepath.Dir(outPath), 0755)
					outFile, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zipFile.Mode())
					if err == nil {
						rc, err := zipFile.Open()
						if err == nil {
							_, _ = io.Copy(outFile, rc)
							rc.Close()
						}
						outFile.Close()
					}
				}

				entries, _ := os.ReadDir(tmpDir)
				for _, entry := range entries {
					if entry.IsDir() && entry.Name() != "repo" {
						repoTargetDir = filepath.Join(tmpDir, entry.Name())
						break
					}
				}
			}
		} else {
			if isDirHasFiles(persistentRepoDir) {
				repoTargetDir = persistentRepoDir
			} else {
				return nil, fmt.Errorf("failed to load GitHub repo '%s' at ref '%s'", cleanRepo, gitRef)
			}
		}
	}

	// Mirror downloaded tree into persistent directory
	if repoTargetDir != persistentRepoDir && isDir(repoTargetDir) {
		_ = copyDirContents(repoTargetDir, persistentRepoDir)
		repoTargetDir = persistentRepoDir
	}

	loadedSkills := make(map[string]*SkillDefinition)
	filterSet := make(map[string]bool)
	for _, s := range selectedSkills {
		filterSet[s] = true
	}
	hasFilter := len(filterSet) > 0

	for _, rootRel := range rootPaths {
		candidateDir := repoTargetDir
		if rootRel != "." {
			candidateDir = filepath.Join(repoTargetDir, rootRel)
		}
		sourceURIBase := fmt.Sprintf("github://%s@%s", cleanRepo, gitRef)
		if rootRel != "." && rootRel != "skills" {
			sourceURIBase = fmt.Sprintf("github://%s@%s/%s", cleanRepo, gitRef, rootRel)
		}
		if isDir(candidateDir) {
			if singleSkill, err := LoadSkillFromDir(candidateDir); err == nil && singleSkill != nil {
				if !hasFilter || filterSet[singleSkill.Name] {
					singleSkill.SourceURI = sourceURIBase
					loadedSkills[singleSkill.Name] = singleSkill
				}
			} else {
				skills, err := LoadAllSkills(candidateDir, selectedSkills)
				if err == nil {
					for k, v := range skills {
						if v.SourceURI == "" || strings.HasPrefix(v.SourceURI, "file://") {
							if rootRel == "." || rootRel == "skills" {
								v.SourceURI = fmt.Sprintf("%s/%s", sourceURIBase, v.Name)
							} else {
								v.SourceURI = sourceURIBase
							}
						}
						loadedSkills[k] = v
					}
				}
			}
		}
	}

	return loadedSkills, nil
}

// LoadSkillsFromRoots loads enterprise skills across multiple qualified root URIs.
func LoadSkillsFromRoots(roots []string, filter []string, token string, dotenvPath string) (map[string]*SkillDefinition, error) {
	envFile := dotenvPath
	if envFile == "" {
		cwd, _ := os.Getwd()
		envFile = filepath.Join(cwd, ".env")
	}
	dotenvVars := ParseDotenvFile(envFile)

	rootURIs := roots
	if len(rootURIs) == 0 {
		if envRoots := os.Getenv("SKILLS_ROOTS"); envRoots != "" {
			for _, r := range strings.Split(envRoots, ",") {
				if strings.TrimSpace(r) != "" {
					rootURIs = append(rootURIs, strings.TrimSpace(r))
				}
			}
		} else if dotRoots := dotenvVars["SKILLS_ROOTS"]; dotRoots != "" {
			for _, r := range strings.Split(dotRoots, ",") {
				if strings.TrimSpace(r) != "" {
					rootURIs = append(rootURIs, strings.TrimSpace(r))
				}
			}
		} else {
			rootURIs = []string{"file://."}
		}
	}

	var selectedSkills []string
	if len(filter) > 0 {
		selectedSkills = filter
	} else if envFilt := os.Getenv("SKILLS_FILTER"); envFilt != "" {
		for _, s := range strings.Split(envFilt, ",") {
			if strings.TrimSpace(s) != "" {
				selectedSkills = append(selectedSkills, strings.TrimSpace(s))
			}
		}
	} else if dotFilt := dotenvVars["SKILLS_FILTER"]; dotFilt != "" {
		for _, s := range strings.Split(dotFilt, ",") {
			if strings.TrimSpace(s) != "" {
				selectedSkills = append(selectedSkills, strings.TrimSpace(s))
			}
		}
	}

	loaded := make(map[string]*SkillDefinition)
	filterSet := make(map[string]bool)
	for _, s := range selectedSkills {
		filterSet[s] = true
	}
	hasFilter := len(filterSet) > 0

	for _, uri := range rootURIs {
		scheme, target, ref, subpath := ParseSkillRootURI(uri)
		switch scheme {
		case "file":
			p := target
			if !filepath.IsAbs(p) {
				baseRoot := FindRegistryRoot()
				pCand := filepath.Join(baseRoot, target)
				if isDir(pCand) || isFile(pCand) {
					p = pCand
				} else {
					absP, err := filepath.Abs(target)
					if err == nil {
						p = absP
					}
				}
			}

			if isDir(p) {
				if single, err := LoadSkillFromDir(p); err == nil && single != nil {
					if !hasFilter || filterSet[single.Name] {
						loaded[single.Name] = single
					}
				} else {
					skills, err := LoadAllSkills(p, selectedSkills)
					if err == nil {
						for k, v := range skills {
							loaded[k] = v
						}
					}
				}
			}
		case "pkg":
			skills, err := LoadSkillsFromPackage(target, selectedSkills)
			if err == nil {
				for k, v := range skills {
					loaded[k] = v
				}
			}
		case "mod", "go":
			var modRoots []string
			if subpath != "" {
				modRoots = []string{subpath}
			}
			skills, err := LoadSkillsFromGoModule(target, ref, modRoots, selectedSkills)
			if err == nil {
				for k, v := range skills {
					loaded[k] = v
				}
			}
		case "maven", "mvn":
			var mavenRoots []string
			if subpath != "" {
				mavenRoots = []string{subpath}
			}
			skills, err := LoadSkillsFromMaven(target, ref, mavenRoots, selectedSkills)
			if err == nil {
				for k, v := range skills {
					loaded[k] = v
				}
			}
		case "github":
			var ghRoots []string
			if subpath != "" {
				ghRoots = []string{subpath}
			}
			skills, err := LoadSkillsFromGitHub(target, ref, ghRoots, selectedSkills, token, dotenvPath)
			if err == nil {
				for k, v := range skills {
					loaded[k] = v
				}
			}
		}
	}

	return loaded, nil
}

// BuildSkillsManifest builds a pre-compiled JSON manifest of all skills for fast loading.
func BuildSkillsManifest(skillsRoot string, outputPath string) (string, error) {
	skills, err := LoadAllSkills(skillsRoot, nil)
	if err != nil {
		return "", err
	}

	outFile := outputPath
	if outFile == "" {
		outFile = filepath.Join(GetLoaderSkillsDir(), "skills_manifest.json")
	}

	manifestDir, _ := filepath.Abs(filepath.Dir(outFile))
	_ = os.MkdirAll(manifestDir, 0755)

	registryRoot := FindRegistryRoot()

	manifestData := make(map[string]map[string]any)
	for name, s := range skills {
		skillAbs := s.Path
		if !filepath.IsAbs(skillAbs) {
			skillAbs = filepath.Join(registryRoot, skillAbs)
		}

		relPath := s.Path
		if r, err := filepath.Rel(manifestDir, skillAbs); err == nil {
			relPath = r
		}

		srcURI := s.SourceURI
		if srcURI == "" || strings.HasPrefix(srcURI, "file://") {
			srcURI = fmt.Sprintf("file://%s", relPath)
		}

		m := s.ToMap()
		m["path"] = relPath
		m["source_uri"] = srcURI
		m["references"] = s.References
		m["examples"] = s.Examples
		manifestData[name] = m
	}

	data, err := json.MarshalIndent(manifestData, "", "  ")
	if err != nil {
		return "", err
	}

	err = os.WriteFile(outFile, data, 0644)
	if err != nil {
		return "", err
	}
	return outFile, nil
}

// LoadSkillsFromManifest loads skill definitions directly from a pre-compiled JSON manifest file.
func LoadSkillsFromManifest(manifestPath string) (map[string]*SkillDefinition, error) {
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		return map[string]*SkillDefinition{}, err
	}

	var rawMap map[string]struct {
		Name          string            `json:"name"`
		Description   string            `json:"description"`
		Instructions  string            `json:"instructions"`
		License       string            `json:"license"`
		Author        string            `json:"author"`
		Version       string            `json:"version"`
		Compatibility string            `json:"compatibility"`
		Metadata      map[string]string `json:"metadata"`
		References    map[string]string `json:"references"`
		Examples      map[string]string `json:"examples"`
		Path          string            `json:"path"`
		SourceURI     string            `json:"source_uri"`
		SHA256Hash    string            `json:"sha256_hash"`
	}

	if err := json.Unmarshal(content, &rawMap); err != nil {
		return map[string]*SkillDefinition{}, err
	}

	loaded := make(map[string]*SkillDefinition)
	for name, data := range rawMap {
		loaded[name] = &SkillDefinition{
			Name:          data.Name,
			Description:   data.Description,
			Instructions:  data.Instructions,
			License:       data.License,
			Author:        data.Author,
			Version:       data.Version,
			Compatibility: data.Compatibility,
			Metadata:      data.Metadata,
			References:    data.References,
			Examples:      data.Examples,
			Path:          data.Path,
			SourceURI:     data.SourceURI,
			SHA256Hash:    data.SHA256Hash,
		}
	}
	return loaded, nil
}

// Registry defines the contract for querying, searching, and evaluating loaded skills.
type Registry interface {
	Skills() map[string]*SkillDefinition
	Get(name string) *SkillDefinition
	ListSkills() []SkillSummary
	Search(query string) []*SkillDefinition
	GetDomainSkills(domain string) []*SkillDefinition
	SuggestSkills(prompt string, maxSkills int, serverURL string) []*SkillDefinition
}

// SkillRegistry provides high-performance skill discovery and querying for Go agent applications.
type SkillRegistry struct {
	root   string
	skills map[string]*SkillDefinition
}

// NewSkillRegistry creates a SkillRegistry from a skills root directory or qualified roots.
func NewSkillRegistry(skillsRoot string, roots []string, filter []string, dotenvPath string) (*SkillRegistry, error) {
	root := skillsRoot
	if root == "" {
		root = FindRegistryRoot()
	}

	var rootList []string
	if len(roots) > 0 {
		rootList = roots
	} else if skillsRoot != "" {
		rootList = []string{skillsRoot}
	}

	loaded, err := LoadSkillsFromRoots(rootList, filter, "", dotenvPath)
	if err != nil {
		return nil, err
	}

	return &SkillRegistry{
		root:   root,
		skills: loaded,
	}, nil
}

// NewSkillRegistryFromRoots creates a SkillRegistry from explicit root URIs.
func NewSkillRegistryFromRoots(roots []string, filter []string, token string, dotenvPath string) (*SkillRegistry, error) {
	loaded, err := LoadSkillsFromRoots(roots, filter, token, dotenvPath)
	if err != nil {
		return nil, err
	}
	cwd, _ := os.Getwd()
	return &SkillRegistry{
		root:   cwd,
		skills: loaded,
	}, nil
}

// NewSkillRegistryFromGitHub creates a SkillRegistry populated from a remote GitHub repository.
func NewSkillRegistryFromGitHub(repo string, ref string, roots []string, filter []string, token string, dotenvPath string) (*SkillRegistry, error) {
	loaded, err := LoadSkillsFromGitHub(repo, ref, roots, filter, token, dotenvPath)
	if err != nil {
		return nil, err
	}
	cwd, _ := os.Getwd()
	return &SkillRegistry{
		root:   cwd,
		skills: loaded,
	}, nil
}

// Skills returns map of loaded skill definitions.
func (r *SkillRegistry) Skills() map[string]*SkillDefinition {
	return r.skills
}

// Get retrieves a skill definition by name.
func (r *SkillRegistry) Get(name string) *SkillDefinition {
	return r.skills[name]
}

// ListSkills returns summarized metadata for all registered skills.
func (r *SkillRegistry) ListSkills() []SkillSummary {
	summaries := make([]SkillSummary, 0, len(r.skills))
	for _, s := range r.skills {
		summaries = append(summaries, SkillSummary{
			Name:           s.Name,
			Description:    s.Description,
			ReferenceCount: len(s.References),
			ExampleCount:   len(s.Examples),
			Path:           s.Path,
			Category:       s.Category,
			Tags:           s.Tags,
			TriggerPhrases: s.TriggerPhrases,
		})
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Name < summaries[j].Name
	})
	return summaries
}

// Search searches skills by keyword matching in name, description, instructions, or references.
func (r *SkillRegistry) Search(query string) []*SkillDefinition {
	q := strings.ToLower(query)
	results := make([]*SkillDefinition, 0)
	for _, s := range r.skills {
		match := strings.Contains(strings.ToLower(s.Name), q) ||
			strings.Contains(strings.ToLower(s.Description), q) ||
			strings.Contains(strings.ToLower(s.Instructions), q)

		if !match {
			for _, refText := range s.References {
				if strings.Contains(strings.ToLower(refText), q) {
					match = true
					break
				}
			}
		}

		if match {
			results = append(results, s)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})
	return results
}

// GetDomainSkills retrieves skills matching a specific language or domain.
func (r *SkillRegistry) GetDomainSkills(domain string) []*SkillDefinition {
	domainNorm := strings.ToLower(domain)
	results := make([]*SkillDefinition, 0)
	for _, s := range r.skills {
		if strings.Contains(strings.ToLower(s.Name), domainNorm) ||
			strings.Contains(strings.ToLower(s.Description), domainNorm) {
			results = append(results, s)
		}
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Name < results[j].Name
	})
	return results
}

// SuggestSkills dynamically suggests the top-k most relevant skills for an agent prompt using remote vector search with local fallback.
func (r *SkillRegistry) SuggestSkills(prompt string, maxSkills int, serverURL string) []*SkillDefinition {
	if strings.TrimSpace(prompt) == "" {
		all := make([]*SkillDefinition, 0, len(r.skills))
		for _, s := range r.skills {
			all = append(all, s)
			if len(all) >= maxSkills {
				break
			}
		}
		return all
	}

	boundedMax := maxSkills
	if boundedMax <= 0 {
		boundedMax = 3
	}
	if boundedMax > 25 {
		boundedMax = 25
	}

	targetServer := serverURL
	if targetServer == "" {
		targetServer = os.Getenv("SKILLS_SERVER_URL")
		if targetServer == "" {
			targetServer = os.Getenv("SKM_SERVER_URL")
		}
	}

	if targetServer != "" {
		endpoint := fmt.Sprintf("%s/api/v1/skills?s=%s&page_size=%d", strings.TrimRight(targetServer, "/"), url.QueryEscape(prompt), boundedMax)
		client := &http.Client{Timeout: 3 * time.Second}
		if resp, err := client.Get(endpoint); err == nil && resp.StatusCode == http.StatusOK {
			defer resp.Body.Close()
			var items []struct {
				Name        string   `json:"name"`
				Description string   `json:"description"`
				Version     string   `json:"version"`
				URI         string   `json:"uri"`
				Tags        []string `json:"tags"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&items); err == nil && len(items) > 0 {
				remoteMatches := make([]*SkillDefinition, 0, len(items))
				for _, item := range items {
					if localSkill, exists := r.skills[item.Name]; exists {
						remoteMatches = append(remoteMatches, localSkill)
					} else {
						remoteMatches = append(remoteMatches, &SkillDefinition{
							Name:        item.Name,
							Description: item.Description,
							Version:     item.Version,
							Path:        item.URI,
							Tags:        item.Tags,
						})
					}
				}
				if len(remoteMatches) > 0 {
					if len(remoteMatches) > boundedMax {
						return remoteMatches[:boundedMax]
					}
					return remoteMatches
				}
			}
		}
	}

	// Local fallback: search matching keywords
	localMatches := r.Search(prompt)
	if len(localMatches) > 0 {
		if len(localMatches) > boundedMax {
			return localMatches[:boundedMax]
		}
		return localMatches
	}

	// Domain fallback
	words := strings.Fields(strings.ToLower(prompt))
	domainMatches := make([]*SkillDefinition, 0)
	seen := make(map[string]bool)
	for _, w := range words {
		for _, s := range r.GetDomainSkills(w) {
			if !seen[s.Name] {
				seen[s.Name] = true
				domainMatches = append(domainMatches, s)
			}
		}
	}
	if len(domainMatches) > 0 {
		if len(domainMatches) > boundedMax {
			return domainMatches[:boundedMax]
		}
		return domainMatches
	}

	// Return first available skills up to boundedMax
	fallback := make([]*SkillDefinition, 0, boundedMax)
	for _, s := range r.skills {
		fallback = append(fallback, s)
		if len(fallback) >= boundedMax {
			break
		}
	}
	return fallback
}

// --- Helper Functions ---

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func isDirHasFiles(path string) bool {
	if !isDir(path) {
		return false
	}
	entries, err := os.ReadDir(path)
	return err == nil && len(entries) > 0
}

func isWithinBaseDir(baseDir, targetPath string) bool {
	rel, err := filepath.Rel(baseDir, targetPath)
	if err != nil {
		return false
	}
	return !strings.HasPrefix(rel, "..")
}

func rsplitOnce(s, sep string) (string, string) {
	idx := strings.LastIndex(s, sep)
	if idx == -1 {
		return s, ""
	}
	return s[:idx], s[idx+len(sep):]
}

func copyDirContents(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(targetPath, data, info.Mode())
	})
}

// ManifestLockEntry holds recorded checksum and source URI for an installed skill.
type ManifestLockEntry struct {
	SkillName string `json:"skill_name"`
	URI       string `json:"uri"`
	SHA256    string `json:"sha256"`
}

// ManifestLock represents the .manifest.lock file contents.
type ManifestLock struct {
	Version string                       `json:"version"`
	Skills  map[string]ManifestLockEntry `json:"skills"`
}

// VerificationResult details the checksum verification status for a skill.
type VerificationResult struct {
	SkillName string `json:"skill_name"`
	URI       string `json:"uri"`
	Status    string `json:"status"` // "verified", "modified", "missing"
	Expected  string `json:"expected_sha256,omitempty"`
	Actual    string `json:"actual_sha256,omitempty"`
	Error     string `json:"error,omitempty"`
}

// VerificationReport encapsulates overall integrity audit results.
type VerificationReport struct {
	TargetDir     string               `json:"target_dir"`
	TotalSkills   int                  `json:"total_skills"`
	VerifiedCount int                  `json:"verified_count"`
	ModifiedCount int                  `json:"modified_count"`
	MissingCount  int                  `json:"missing_count"`
	Results       []VerificationResult `json:"results"`
}

// CalculateSkillChecksum computes a deterministic SHA256 checksum of a skill directory's contents.
func CalculateSkillChecksum(skillDir string) (string, error) {
	absDir, err := filepath.Abs(skillDir)
	if err != nil {
		absDir = skillDir
	}

	if !isDir(absDir) {
		return "", fmt.Errorf("skill directory does not exist: %s", absDir)
	}

	hasher := sha256.New()
	var files []string

	err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			name := info.Name()
			if name == ".DS_Store" || name == ".manifest.lock" {
				return nil
			}
			rel, err := filepath.Rel(absDir, path)
			if err == nil {
				files = append(files, rel)
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	sort.Strings(files)
	for _, rel := range files {
		normRel := filepath.ToSlash(rel)
		hasher.Write([]byte(normRel))
		content, err := os.ReadFile(filepath.Join(absDir, rel))
		if err != nil {
			return "", fmt.Errorf("failed reading file %s for checksum: %w", rel, err)
		}
		hasher.Write(content)
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// ReadManifestLock loads the .manifest.lock file from a target destination directory.
func ReadManifestLock(destDir string) (*ManifestLock, error) {
	lockFile := filepath.Join(destDir, ".manifest.lock")
	content, err := os.ReadFile(lockFile)
	if err != nil {
		return nil, fmt.Errorf(".manifest.lock file not found in %s: %w", destDir, err)
	}

	var lock ManifestLock
	if err := json.Unmarshal(content, &lock); err != nil {
		return nil, fmt.Errorf("failed to parse .manifest.lock: %w", err)
	}
	if lock.Skills == nil {
		lock.Skills = make(map[string]ManifestLockEntry)
	}
	return &lock, nil
}

// WriteManifestLock writes the .manifest.lock file to a target destination directory.
func WriteManifestLock(destDir string, lock *ManifestLock) error {
	if lock == nil {
		return fmt.Errorf("lock cannot be nil")
	}
	if lock.Version == "" {
		lock.Version = "1.0.0"
	}
	if lock.Skills == nil {
		lock.Skills = make(map[string]ManifestLockEntry)
	}

	lockFile := filepath.Join(destDir, ".manifest.lock")
	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal .manifest.lock: %w", err)
	}

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create target dir %s: %w", destDir, err)
	}

	return os.WriteFile(lockFile, data, 0644)
}

// UpdateManifestLock updates or adds a skill entry in the destination directory's .manifest.lock file.
func UpdateManifestLock(destDir string, skillName string, uri string, checksum string) error {
	lock, err := ReadManifestLock(destDir)
	if err != nil {
		lock = &ManifestLock{
			Version: "1.0.0",
			Skills:  make(map[string]ManifestLockEntry),
		}
	}

	if checksum == "" {
		skillDir := filepath.Join(destDir, skillName)
		cs, err := CalculateSkillChecksum(skillDir)
		if err != nil {
			return err
		}
		checksum = cs
	}

	lock.Skills[skillName] = ManifestLockEntry{
		SkillName: skillName,
		URI:       uri,
		SHA256:    checksum,
	}

	return WriteManifestLock(destDir, lock)
}

// VerifyManifestLock validates that skills present in .manifest.lock match their original recorded checksums.
func VerifyManifestLock(destDir string) (*VerificationReport, error) {
	lock, err := ReadManifestLock(destDir)
	if err != nil {
		return nil, err
	}

	report := &VerificationReport{
		TargetDir:   destDir,
		TotalSkills: len(lock.Skills),
		Results:     make([]VerificationResult, 0, len(lock.Skills)),
	}

	var names []string
	for k := range lock.Skills {
		names = append(names, k)
	}
	sort.Strings(names)

	for _, name := range names {
		entry := lock.Skills[name]
		skillDir := filepath.Join(destDir, name)

		if !isDir(skillDir) {
			report.MissingCount++
			report.Results = append(report.Results, VerificationResult{
				SkillName: name,
				URI:       entry.URI,
				Status:    "missing",
				Expected:  entry.SHA256,
				Error:     "skill directory missing",
			})
			continue
		}

		currentCS, err := CalculateSkillChecksum(skillDir)
		if err != nil {
			report.ModifiedCount++
			report.Results = append(report.Results, VerificationResult{
				SkillName: name,
				URI:       entry.URI,
				Status:    "modified",
				Expected:  entry.SHA256,
				Error:     fmt.Sprintf("failed to compute checksum: %v", err),
			})
			continue
		}

		if currentCS == entry.SHA256 {
			report.VerifiedCount++
			report.Results = append(report.Results, VerificationResult{
				SkillName: name,
				URI:       entry.URI,
				Status:    "verified",
				Expected:  entry.SHA256,
				Actual:    currentCS,
			})
		} else {
			report.ModifiedCount++
			report.Results = append(report.Results, VerificationResult{
				SkillName: name,
				URI:       entry.URI,
				Status:    "modified",
				Expected:  entry.SHA256,
				Actual:    currentCS,
				Error:     "checksum mismatch (skill files modified)",
			})
		}
	}

	return report, nil
}

// LoadSkillsFromSKMServer fetches a skill definition from a central SKM server.
func LoadSkillsFromSKMServer(target string, filter []string, serverURL string, apiKey string) (map[string]*SkillDefinition, error) {
	if serverURL == "" {
		serverURL = "http://localhost:8080"
	}
	endpoint := fmt.Sprintf("%s/api/v1/skills/%s", strings.TrimRight(serverURL, "/"), target)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request for %s: %w", endpoint, err)
	}
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch skill from server %s: %w", endpoint, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server error (%d): %s", resp.StatusCode, string(body))
	}

	var data struct {
		ID           string            `json:"id"`
		Name         string            `json:"name"`
		URI          string            `json:"uri"`
		SourceURI    string            `json:"source_uri"`
		Description  string            `json:"description"`
		Instructions string            `json:"instructions"`
		License      *string           `json:"license"`
		Author       *string           `json:"author"`
		Version      string            `json:"version"`
		Category     *string           `json:"category"`
		Tags         []string          `json:"tags"`
		Metadata     map[string]string `json:"metadata"`
		References   map[string]string `json:"references"`
		Examples     map[string]string `json:"examples"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse skill response: %w", err)
	}

	lic := ""
	if data.License != nil {
		lic = *data.License
	}
	auth := ""
	if data.Author != nil {
		auth = *data.Author
	}

	s := &SkillDefinition{
		Name:         data.Name,
		Description:  data.Description,
		Instructions: data.Instructions,
		License:      lic,
		Author:       auth,
		Version:      data.Version,
		Metadata:     data.Metadata,
		References:   data.References,
		Examples:     data.Examples,
	}

	tmpDir, err := os.MkdirTemp("", "skm-server-skill-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp dir: %w", err)
	}
	s.Path = tmpDir

	skillMDContent := fmt.Sprintf("---\nname: %s\ndescription: %s\nlicense: %s\nauthor: %s\nversion: %s\n---\n\n# %s\n\n%s\n",
		s.Name, s.Description, s.License, s.Author, s.Version, s.Name, s.Instructions)
	if strings.HasPrefix(strings.TrimSpace(s.Instructions), "---") || strings.HasPrefix(strings.TrimSpace(s.Instructions), "# ") {
		skillMDContent = s.Instructions
	}
	_ = os.WriteFile(filepath.Join(tmpDir, "SKILL.md"), []byte(skillMDContent), 0644)

	writeFilePayload := func(destPath string, content string) error {
		var data []byte
		if strings.HasPrefix(content, "data:") && strings.Contains(content, ";base64,") {
			parts := strings.SplitN(content, ";base64,", 2)
			if decoded, err := base64.StdEncoding.DecodeString(parts[1]); err == nil {
				data = decoded
			} else {
				data = []byte(content)
			}
		} else {
			data = []byte(content)
		}
		return os.WriteFile(destPath, data, 0644)
	}

	if len(s.References) > 0 {
		refDir := filepath.Join(tmpDir, "references")
		_ = os.MkdirAll(refDir, 0755)
		for name, content := range s.References {
			_ = writeFilePayload(filepath.Join(refDir, name), content)
		}
	}

	if len(s.Examples) > 0 {
		exDir := filepath.Join(tmpDir, "examples")
		_ = os.MkdirAll(exDir, 0755)
		for name, content := range s.Examples {
			_ = writeFilePayload(filepath.Join(exDir, name), content)
		}
	}

	return map[string]*SkillDefinition{s.Name: s}, nil
}


