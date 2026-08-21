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

package installer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/retail-cortex/castor/clients/go/pkg/skillsloader"
)

// AddResult holds details of a skill copy/install operation.
type AddResult struct {
	SkillName   string `json:"skill_name"`
	SourcePath  string `json:"source_path"`
	DestPath    string `json:"dest_path"`
	Status      string `json:"status"` // "added", "overwritten", "skipped", "failed"
	ErrorDetail string `json:"error_detail,omitempty"`
}

// AddSkills resolves URI specifications and copies matching skills to destDir.
func AddSkills(uris []string, destDir string, filter []string, force bool) ([]AddResult, error) {
	if destDir == "" {
		destDir = ".skills"
	}

	absDest, err := filepath.Abs(destDir)
	if err != nil {
		absDest = destDir
	}

	if err := os.MkdirAll(absDest, 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory %s: %w", absDest, err)
	}

	var results []AddResult

	for _, uri := range uris {
		scheme, target, ref, subpath := skillsloader.ParseSkillRootURI(uri)
		var skills map[string]*skillsloader.SkillDefinition
		var loadErr error

		switch scheme {
		case "github":
			var roots []string
			if subpath != "" {
				roots = []string{subpath}
			}
			skills, loadErr = skillsloader.LoadSkillsFromGitHub(target, ref, roots, filter, os.Getenv("GITHUB_TOKEN"), "")
		case "maven", "mvn":
			var roots []string
			if subpath != "" {
				roots = []string{subpath}
			}
			skills, loadErr = skillsloader.LoadSkillsFromMaven(target, ref, roots, filter)
		case "mod", "go":
			var roots []string
			if subpath != "" {
				roots = []string{subpath}
			}
			skills, loadErr = skillsloader.LoadSkillsFromGoModule(target, ref, roots, filter)
		case "castor", "castors", "cstr", "cstrs", "skm", "skms":
			serverURL := os.Getenv("CASTOR_SERVER_URL")
			if serverURL == "" {
				serverURL = os.Getenv("CSTR_SERVER_URL")
			}
			if serverURL == "" {
				serverURL = os.Getenv("SKM_SERVER_URL")
			}
			apiKey := os.Getenv("CASTOR_API_KEY")
			if apiKey == "" {
				apiKey = os.Getenv("CSTR_API_KEY")
			}
			if apiKey == "" {
				apiKey = os.Getenv("SKM_API_KEY")
			}
			if serverURL == "" || apiKey == "" {
				if home, err := os.UserHomeDir(); err == nil {
					for _, dir := range []string{".castor", ".cstr", ".skm"} {
						cfgPath := filepath.Join(home, dir, ".env.toml")
						if content, err := os.ReadFile(cfgPath); err == nil {
							for _, line := range strings.Split(string(content), "\n") {
								parts := strings.SplitN(line, "=", 2)
								if len(parts) == 2 {
									k := strings.TrimSpace(parts[0])
									v := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
									if serverURL == "" && (k == "CASTOR_SERVER_URL" || k == "CSTR_SERVER_URL" || k == "SKM_SERVER_URL" || k == "server_url" || k == "SERVER_URL") {
										serverURL = v
									}
									if apiKey == "" && (k == "CASTOR_API_KEY" || k == "CSTR_API_KEY" || k == "SKM_API_KEY" || k == "api_key" || k == "API_KEY") {
										apiKey = v
									}
								}
							}
						}
					}
				}
			}
			skills, loadErr = skillsloader.LoadSkillsFromCastorRegistry(target, filter, serverURL, apiKey)
		case "file":
			skills, loadErr = ResolveFileSkills(target, filter)
		default:
			skills, loadErr = ResolveFileSkills(uri, filter)
		}


		if loadErr != nil {
			results = append(results, AddResult{
				SkillName:   uri,
				Status:      "failed",
				ErrorDetail: fmt.Sprintf("failed to resolve skill URI %s: %v", uri, loadErr),
			})
			continue
		}

		if len(skills) == 0 {
			results = append(results, AddResult{
				SkillName:   uri,
				Status:      "failed",
				ErrorDetail: fmt.Sprintf("no valid skills found at %s", uri),
			})
			continue
		}

		for skillName, skillDef := range skills {
			targetPath := filepath.Join(absDest, skillName)
			exists := isDir(targetPath)

			if exists && !force {
				results = append(results, AddResult{
					SkillName:   skillName,
					SourcePath:  skillDef.Path,
					DestPath:    targetPath,
					Status:      "skipped",
					ErrorDetail: "destination directory exists (use --force to overwrite)",
				})
				continue
			}

			if exists && force {
				if err := os.RemoveAll(targetPath); err != nil {
					results = append(results, AddResult{
						SkillName:   skillName,
						SourcePath:  skillDef.Path,
						DestPath:    targetPath,
						Status:      "failed",
						ErrorDetail: fmt.Sprintf("failed to clean existing dir: %v", err),
					})
					continue
				}
			}

			if err := copyDirectory(skillDef.Path, targetPath); err != nil {
				results = append(results, AddResult{
					SkillName:   skillName,
					SourcePath:  skillDef.Path,
					DestPath:    targetPath,
					Status:      "failed",
					ErrorDetail: fmt.Sprintf("failed to copy files: %v", err),
				})
				continue
			}

			// Update .manifest.lock with skill_name, uri, and sha256 checksum
			if lockErr := skillsloader.UpdateManifestLock(absDest, skillName, uri, ""); lockErr != nil {
				// Non-fatal warning if manifest lock update encounters issue
				_ = lockErr
			}

			status := "added"
			if exists && force {
				status = "overwritten"
			}

			results = append(results, AddResult{
				SkillName:  skillName,
				SourcePath: skillDef.Path,
				DestPath:   targetPath,
				Status:     status,
			})
		}
	}

	return results, nil
}

func ResolveFileSkills(target string, filter []string) (map[string]*skillsloader.SkillDefinition, error) {
	// Expand path
	path := target
	if strings.HasPrefix(path, "file://") {
		path = strings.TrimPrefix(path, "file://")
	}
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, path[2:])
		}
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}

	if isFile(filepath.Join(absPath, "SKILL.md")) {
		skill, err := skillsloader.LoadSkillFromDir(absPath)
		if err != nil {
			return nil, err
		}
		if len(filter) > 0 {
			match := false
			for _, f := range filter {
				if f == skill.Name {
					match = true
					break
				}
			}
			if !match {
				return make(map[string]*skillsloader.SkillDefinition), nil
			}
		}
		return map[string]*skillsloader.SkillDefinition{skill.Name: skill}, nil
	}

	// Try loading all skills under the directory
	return skillsloader.LoadAllSkills(absPath, filter)
}

func copyDirectory(src, dst string) error {
	absSrc, err := filepath.Abs(src)
	if err != nil {
		absSrc = src
	}
	realSrc, err := filepath.EvalSymlinks(absSrc)
	if err != nil {
		realSrc = absSrc
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Security Checkpoint (CWE-59 / CWE-22): Verify path resolves within source directory boundary
		realPath, errEval := filepath.EvalSymlinks(path)
		if errEval == nil {
			relToSrc, relErr := filepath.Rel(realSrc, realPath)
			if relErr != nil || strings.HasPrefix(relToSrc, "..") {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		// Copy file content
		return copyFileContents(path, targetPath, info.Mode())
	})
}

func copyFileContents(srcFile, dstFile string, perm os.FileMode) error {
	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dstFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
