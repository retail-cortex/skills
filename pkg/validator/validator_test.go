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

package validator_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/retail-cortex/castor/pkg/validator"
	"github.com/stretchr/testify/assert"
)

func TestParseFrontmatter(t *testing.T) {
	content := `---
name: test-skill
description: A test skill definition
license: Apache-2.0
author: Ryan McGuinness
version: 1.0.0
---
# Test Skill
This is instructions.
`
	fm, body := validator.ParseFrontmatter(content)
	assert.Equal(t, "test-skill", fm["name"])
	assert.Equal(t, "A test skill definition", fm["description"])
	assert.Equal(t, "Apache-2.0", fm["license"])
	assert.Contains(t, body, "# Test Skill")
}

func TestSkillFrontmatter_Validate(t *testing.T) {
	validFM := validator.SkillFrontmatter{
		Name:        "my-awesome-skill",
		Description: "Detailed description",
		License:     "Apache-2.0",
		Author:      "Tester",
		Version:     "1.0",
	}
	assert.NoError(t, validFM.Validate())

	invalidNameFM := validFM
	invalidNameFM.Name = "Invalid_Name"
	assert.Error(t, invalidNameFM.Validate())

	emptyDescFM := validFM
	emptyDescFM.Description = ""
	assert.Error(t, emptyDescFM.Validate())
}

func TestAuditSkillDirectory_ValidSkill(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skill-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	skillDir := filepath.Join(tmpDir, "valid-skill")
	err = os.MkdirAll(filepath.Join(skillDir, "references"), 0755)
	assert.NoError(t, err)
	err = os.MkdirAll(filepath.Join(skillDir, "examples"), 0755)
	assert.NoError(t, err)

	skillMDContent := `---
name: valid-skill
description: Valid skill for unit test
license: Apache-2.0
author: Ryan McGuinness
version: 1.0.0
---
# Valid Skill

Security Checkpoint: Ensures CWE-20 validation.
HTTP 429 Rate Limit exponential backoff is implemented.
See [docs](file:///path/to/doc.md) for details.
`
	err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(skillMDContent), 0644)
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(skillDir, "references", "ref.md"), []byte("[reference](file:///ref.md)"), 0644)
	assert.NoError(t, err)

	err = os.WriteFile(filepath.Join(skillDir, "examples", "ex.md"), []byte("Example code"), 0644)
	assert.NoError(t, err)

	result := validator.AuditSkillDirectory(skillDir)
	assert.True(t, result.Passed, "Audit result errors: %v", result.Errors)
	assert.True(t, result.FrontmatterValid)
	assert.True(t, result.L3TreeValid)
	assert.True(t, result.CWESecurityValid)
	assert.True(t, result.RateLimit429Valid)
	assert.True(t, result.ClickableLinksValid)
	assert.Empty(t, result.Errors)
}

func TestAuditSkillDirectory_MissingSKILLMD(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skill-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	result := validator.AuditSkillDirectory(tmpDir)
	assert.False(t, result.Passed)
	assert.Contains(t, result.Errors[0], "Missing SKILL.md file")
}

func TestAuditAllSkills_Recursive(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skills-root-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	skillDir1 := filepath.Join(tmpDir, "skill-one")
	skillDir2 := filepath.Join(tmpDir, "sub", "skill-two")

	for _, sDir := range []string{skillDir1, skillDir2} {
		_ = os.MkdirAll(filepath.Join(sDir, "references"), 0755)
		_ = os.MkdirAll(filepath.Join(sDir, "examples"), 0755)
		_ = os.WriteFile(filepath.Join(sDir, "references", "ref.md"), []byte("ref"), 0644)
		_ = os.WriteFile(filepath.Join(sDir, "examples", "ex.md"), []byte("ex"), 0644)

		sName := filepath.Base(sDir)
		content := `---
name: ` + sName + `
description: Skill description
license: Apache-2.0
author: Tester
version: 1.0
---
Security Checkpoint CWE-79.
HTTP 429 Retryablehttp.
Link: [doc](file:///doc.md)
`
		_ = os.WriteFile(filepath.Join(sDir, "SKILL.md"), []byte(content), 0644)
	}

	summary := validator.AuditAllSkills(tmpDir, true)
	assert.GreaterOrEqual(t, summary.TotalSkills, 2)
	assert.Equal(t, summary.TotalSkills, summary.PassedSkills)
	assert.Equal(t, 0, summary.FailedSkills)
}

func TestAuditSummary_ToJSON(t *testing.T) {
	summary := validator.AuditSummary{
		TotalSkills:  2,
		PassedSkills: 1,
		FailedSkills: 1,
		Results: []validator.SkillAuditResult{
			{SkillName: "s1", Passed: true},
			{SkillName: "s2", Passed: false, Errors: []string{"error"}},
		},
	}
	jsonStr, err := summary.ToJSON()
	assert.NoError(t, err)
	assert.Contains(t, jsonStr, "s1")
	assert.Contains(t, jsonStr, "s2")
}

func TestAuditSkillDirectory_Failures(t *testing.T) {
	tmpDir := t.TempDir()

	badFMDir := filepath.Join(tmpDir, "bad-fm")
	_ = os.MkdirAll(filepath.Join(badFMDir, "references"), 0755)
	_ = os.MkdirAll(filepath.Join(badFMDir, "examples"), 0755)
	_ = os.WriteFile(filepath.Join(badFMDir, "references", "ref.md"), []byte("ref"), 0644)
	_ = os.WriteFile(filepath.Join(badFMDir, "examples", "ex.md"), []byte("ex"), 0644)

	badContent := `---
name: BAD_NAME
description: Short
---
Security Checkpoint CWE-10. 429 Retry. [Link](file:///doc.md)
`
	_ = os.WriteFile(filepath.Join(badFMDir, "SKILL.md"), []byte(badContent), 0644)

	res := validator.AuditSkillDirectory(badFMDir)
	assert.False(t, res.Passed)
	assert.False(t, res.FrontmatterValid)

	t.Setenv("TEST_SRCDIR", "")
	noL3Dir := filepath.Join(tmpDir, "no-l3")
	_ = os.MkdirAll(noL3Dir, 0755)
	validContent := `---
name: no-l3
description: Valid description for no l3 skill test
license: Apache-2.0
---
Security Checkpoint CWE-10. 429 Retry. [Link](file:///doc.md)
`
	_ = os.WriteFile(filepath.Join(noL3Dir, "SKILL.md"), []byte(validContent), 0644)

	res2 := validator.AuditSkillDirectory(noL3Dir)
	assert.False(t, res2.Passed)
	assert.False(t, res2.L3TreeValid)
}
