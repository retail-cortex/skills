package installer

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/retail-cortex/skills/clients/go/pkg/skillsloader"
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
		case "pkg":
			skills, loadErr = skillsloader.LoadSkillsFromPackage(target, filter)
		case "file":
			skills, loadErr = resolveFileSkills(target, filter)
		default:
			skills, loadErr = resolveFileSkills(uri, filter)
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

func resolveFileSkills(target string, filter []string) (map[string]*skillsloader.SkillDefinition, error) {
	// Expand path
	path := target
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
