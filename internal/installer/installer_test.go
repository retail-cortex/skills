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
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddSkills_LocalFileURI(t *testing.T) {
	// Setup source skill directory
	srcDir, err := os.MkdirTemp("", "src-skill-*")
	assert.NoError(t, err)
	defer os.RemoveAll(srcDir)

	skillDir := filepath.Join(srcDir, "my-skill")
	err = os.MkdirAll(filepath.Join(skillDir, "references"), 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(filepath.Join(skillDir, "examples"), 0755)
	assert.NoError(t, err)

	skillContent := `---
name: my-skill
description: Test skill
license: Apache-2.0
author: Tester
version: 1.0
---
Instruction text
`
	err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillContent), 0644)
	assert.NoError(t, err)

	// Destination directory
	destDir, err := os.MkdirTemp("", "dest-skills-*")
	assert.NoError(t, err)
	defer os.RemoveAll(destDir)

	// Run AddSkills with file:// URI
	fileURI := "file://" + skillDir
	results, err := AddSkills([]string{fileURI}, destDir, nil, false)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "my-skill", results[0].SkillName)
	assert.Equal(t, "added", results[0].Status)

	// Verify files were copied and .manifest.lock created
	copiedSKILL := filepath.Join(destDir, "my-skill", "SKILL.md")
	assert.True(t, isFile(copiedSKILL))
	assert.True(t, isFile(filepath.Join(destDir, ".manifest.lock")))

	// Re-adding without force should skip
	results2, err := AddSkills([]string{fileURI}, destDir, nil, false)
	assert.NoError(t, err)
	assert.Len(t, results2, 1)
	assert.Equal(t, "skipped", results2[0].Status)

	// Re-adding with force should overwrite
	results3, err := AddSkills([]string{fileURI}, destDir, nil, true)
	assert.NoError(t, err)
	assert.Len(t, results3, 1)
	assert.Equal(t, "overwritten", results3[0].Status)
}

func TestAddSkills_InvalidURI(t *testing.T) {
	destDir := t.TempDir()
	results, err := AddSkills([]string{"file:///nonexistent/path/skill"}, destDir, nil, false)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "failed", results[0].Status)
}

func TestAddSkills_Filter(t *testing.T) {
	srcDir := t.TempDir()
	skillDir := filepath.Join(srcDir, "filter-skill")
	_ = os.MkdirAll(skillDir, 0755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: filter-skill\n---\nBody"), 0644)

	destDir := t.TempDir()

	// Matching filter
	resMatch, err := AddSkills([]string{"file://" + skillDir}, destDir, []string{"filter-skill"}, true)
	assert.NoError(t, err)
	assert.Len(t, resMatch, 1)
	assert.Equal(t, "added", resMatch[0].Status)

	// Non-matching filter
	destDir2 := t.TempDir()
	resNoMatch, err := AddSkills([]string{"file://" + skillDir}, destDir2, []string{"other-skill"}, true)
	assert.NoError(t, err)
	assert.Len(t, resNoMatch, 1)
	assert.Equal(t, "failed", resNoMatch[0].Status)
}

func TestCopyDirectory_SymlinkBoundary(t *testing.T) {
	srcDir := t.TempDir()
	outOfBoundsDir := t.TempDir()

	_ = os.WriteFile(filepath.Join(outOfBoundsDir, "secret.txt"), []byte("secret"), 0644)
	_ = os.WriteFile(filepath.Join(srcDir, "valid.txt"), []byte("valid"), 0644)

	// Create a symlink pointing outside srcDir
	symlinkPath := filepath.Join(srcDir, "external_link")
	_ = os.Symlink(outOfBoundsDir, symlinkPath)

	destDir := t.TempDir()
	err := copyDirectory(srcDir, destDir)
	assert.NoError(t, err)

	// Valid file copied
	assert.True(t, isFile(filepath.Join(destDir, "valid.txt")))
	// Symlink pointing outside is NOT copied/traversed
	assert.False(t, isFile(filepath.Join(destDir, "external_link", "secret.txt")))
}

func TestAddSkills_AllSchemes(t *testing.T) {
	destDir := t.TempDir()

	uris := []string{
		"pkg://retailcortex_skills_python",
		"mod://github.com/retail-cortex/castor@v1.0.0",
		"maven://com.retailcortex.castor:skills-java:1.0.0",
		"github://retail-cortex/castor@main",
	}

	results, err := AddSkills(uris, destDir, []string{"python-core"}, true)
	assert.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestResolveFileSkills_Tilde(t *testing.T) {
	home, err := os.UserHomeDir()
	if err == nil {
		skillDir := filepath.Join(home, ".skills", "test-tilde-skill")
		_ = os.MkdirAll(skillDir, 0755)
		defer os.RemoveAll(skillDir)
		_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: test-tilde-skill\n---\nBody"), 0644)

		skills, err := ResolveFileSkills("~/.skills/test-tilde-skill", nil)
		assert.NoError(t, err)
		assert.Contains(t, skills, "test-tilde-skill")
	}
}

func TestAddSkills_DefaultDestDir(t *testing.T) {
	srcDir := t.TempDir()
	skillDir := filepath.Join(srcDir, "default-dir-skill")
	_ = os.MkdirAll(skillDir, 0755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: default-dir-skill\n---\nBody"), 0644)

	tmpWorkDir := t.TempDir()
	cwd, _ := os.Getwd()
	_ = os.Chdir(tmpWorkDir)
	defer func() { _ = os.Chdir(cwd) }()

	results, err := AddSkills([]string{"file://" + skillDir}, "", nil, true)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "added", results[0].Status)
	assert.True(t, isDir(".skills"))
}
