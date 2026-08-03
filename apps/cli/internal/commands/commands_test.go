package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExecute_Version(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	code := Execute([]string{"version"}, out, errOut)
	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "skm version")
}

func TestExecute_Help(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	code := Execute([]string{"help"}, out, errOut)
	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "SKM - Enterprise Standalone Skills CLI Client")
}

func TestExecute_InitAndValidate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skm-cmd-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	// Run init
	code := Execute([]string{"init", "new-test-skill", "-d", tmpDir}, out, errOut)
	assert.Equal(t, 0, code, "stderr: %s", errOut.String())
	assert.Contains(t, out.String(), "Successfully initialized new skill")

	skillDir := filepath.Join(tmpDir, "new-test-skill")

	// Run validate on the initialized skill
	valOut := &bytes.Buffer{}
	valErr := &bytes.Buffer{}
	code = Execute([]string{"validate", skillDir}, valOut, valErr)
	assert.Equal(t, 0, code, "validation stderr: %s, stdout: %s", valErr.String(), valOut.String())
	assert.Contains(t, valOut.String(), "[PASS] new-test-skill")
}

func TestExecute_AddAndList(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skm-add-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create a dummy skill to add
	srcDir := filepath.Join(tmpDir, "src-skill")
	_ = os.MkdirAll(filepath.Join(srcDir, "references"), 0755)
	_ = os.MkdirAll(filepath.Join(srcDir, "examples"), 0755)
	_ = os.WriteFile(filepath.Join(srcDir, "references", "ref.md"), []byte("ref"), 0644)
	_ = os.WriteFile(filepath.Join(srcDir, "examples", "ex.md"), []byte("ex"), 0644)
	content := `---
name: src-skill
description: Source skill
license: Apache-2.0
author: Tester
version: 1.0
---
Security Checkpoint CWE-10. HTTP 429 Retry. [Link](file:///doc.md)
`
	_ = os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte(content), 0644)

	destDir := filepath.Join(tmpDir, "dest-skills")

	addOut := &bytes.Buffer{}
	addErr := &bytes.Buffer{}
	code := Execute([]string{"add", "file://" + srcDir, "-d", destDir}, addOut, addErr)
	assert.Equal(t, 0, code, "stderr: %s", addErr.String())
	assert.Contains(t, addOut.String(), "[+] Added:       src-skill")

	// Test list
	listOut := &bytes.Buffer{}
	listErr := &bytes.Buffer{}
	code = Execute([]string{"list", "-d", destDir}, listOut, listErr)
	assert.Equal(t, 0, code, "list stderr: %s", listErr.String())
	assert.Contains(t, listOut.String(), "src-skill")
}

func TestExecute_Compile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "skm-compile-test-*")
	assert.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	skillDir := filepath.Join(tmpDir, "skills", "compile-skill")
	_ = os.MkdirAll(filepath.Join(skillDir, "references"), 0755)
	_ = os.MkdirAll(filepath.Join(skillDir, "examples"), 0755)
	_ = os.WriteFile(filepath.Join(skillDir, "references", "ref.md"), []byte("ref"), 0644)
	_ = os.WriteFile(filepath.Join(skillDir, "examples", "ex.md"), []byte("ex"), 0644)
	content := `---
name: compile-skill
description: Skill for compile test
license: Apache-2.0
---
Body content
`
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)

	outFile := filepath.Join(tmpDir, "manifest.json")
	compOut := &bytes.Buffer{}
	compErr := &bytes.Buffer{}

	code := Execute([]string{"compile", "-d", tmpDir, "-o", outFile}, compOut, compErr)
	assert.Equal(t, 0, code, "compile stderr: %s", compErr.String())
	assert.Contains(t, compOut.String(), "Successfully compiled")

	// Verify compiled manifest exists and contains skill
	manifestBytes, errRead := os.ReadFile(outFile)
	assert.NoError(t, errRead)
	assert.Contains(t, string(manifestBytes), "compile-skill")
}

func TestExecute_Search(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "skills", "searchable-skill")
	_ = os.MkdirAll(filepath.Join(skillDir, "references"), 0755)
	_ = os.MkdirAll(filepath.Join(skillDir, "examples"), 0755)
	_ = os.WriteFile(filepath.Join(skillDir, "references", "ref.md"), []byte("ref"), 0644)
	_ = os.WriteFile(filepath.Join(skillDir, "examples", "ex.md"), []byte("ex"), 0644)

	content := `---
name: searchable-skill
description: Searchable skill description
license: Apache-2.0
---
Instruction with KeywordAlpha.
`
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)

	// Plain text search
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	code := Execute([]string{"search", "KeywordAlpha", "-d", tmpDir}, out, errOut)
	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "searchable-skill")
}

func TestExecute_ValidateJSON(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "json-skill")
	_ = os.MkdirAll(filepath.Join(skillDir, "references"), 0755)
	_ = os.MkdirAll(filepath.Join(skillDir, "examples"), 0755)
	_ = os.WriteFile(filepath.Join(skillDir, "references", "ref.md"), []byte("ref"), 0644)
	_ = os.WriteFile(filepath.Join(skillDir, "examples", "ex.md"), []byte("ex"), 0644)

	content := `---
name: json-skill
description: Valid JSON skill description
license: Apache-2.0
author: Tester
version: 1.0.0
---
Security Checkpoint CWE-10. HTTP 429 Retry. [Link](file:///doc.md)
`
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	code := Execute([]string{"validate", "-r", "--json", tmpDir}, out, errOut)
	assert.Equal(t, 0, code, "stderr: %s, stdout: %s", errOut.String(), out.String())
	assert.Contains(t, out.String(), "\"total_skills\"")
}

func TestExecute_UnknownCommand(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	code := Execute([]string{"bogus-command"}, out, errOut)
	assert.Equal(t, 1, code)
	assert.Contains(t, errOut.String(), "Unknown command: bogus-command")
}

func TestExecute_AddFlags(t *testing.T) {
	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src-skill")
	_ = os.MkdirAll(srcDir, 0755)
	_ = os.WriteFile(filepath.Join(srcDir, "SKILL.md"), []byte("---\nname: src-skill\n---\nBody"), 0644)

	destDir := filepath.Join(tmpDir, "target")

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	code := Execute([]string{"add", "--dir=" + destDir, "--filter=src-skill", "--force", "file://" + srcDir}, out, errOut)
	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "src-skill")

	// Test unknown flag error
	errBuf := &bytes.Buffer{}
	code = Execute([]string{"add", "--unknown-flag"}, out, errBuf)
	assert.Equal(t, 1, code)
	assert.Contains(t, errBuf.String(), "Unknown flag: --unknown-flag")
}

func TestExecute_InitErrors(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	// Missing skill name
	code := Execute([]string{"init"}, out, errOut)
	assert.Equal(t, 1, code)
	assert.Contains(t, errOut.String(), "Error: skill name required")

	// Existing directory
	tmpDir := t.TempDir()
	existing := filepath.Join(tmpDir, "existing-skill")
	_ = os.MkdirAll(existing, 0755)

	errOut2 := &bytes.Buffer{}
	code = Execute([]string{"init", "existing-skill", "-d", tmpDir}, out, errOut2)
	assert.Equal(t, 1, code)
	assert.Contains(t, errOut2.String(), "already exists")
}

func TestExecute_ListFlags(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "skills", "list-skill")
	_ = os.MkdirAll(skillDir, 0755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: list-skill\ndescription: desc\n---\nBody"), 0644)

	// List text
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	code := Execute([]string{"list", "-d", tmpDir}, out, errOut)
	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "list-skill")

	// List JSON
	jsonOut := &bytes.Buffer{}
	jsonErr := &bytes.Buffer{}
	code = Execute([]string{"list", "-d", tmpDir, "--json"}, jsonOut, jsonErr)
	assert.Equal(t, 0, code)
	assert.Contains(t, jsonOut.String(), "\"name\": \"list-skill\"")
}

func TestExecute_ValidateSingleDir(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "valid-single-skill")
	_ = os.MkdirAll(filepath.Join(skillDir, "references"), 0755)
	_ = os.MkdirAll(filepath.Join(skillDir, "examples"), 0755)
	_ = os.WriteFile(filepath.Join(skillDir, "references", "ref.md"), []byte("ref"), 0644)
	_ = os.WriteFile(filepath.Join(skillDir, "examples", "ex.md"), []byte("ex"), 0644)

	content := `---
name: valid-single-skill
description: Single skill description
license: Apache-2.0
author: Tester
version: 1.0.0
---
Security Checkpoint CWE-10. HTTP 429 Retry. [Link](file:///doc.md)
`
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)

	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	code := Execute([]string{"validate", skillDir}, out, errOut)
	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "[PASS] valid-single-skill")
}

func TestExecute_NoArgs(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	code := Execute([]string{}, out, errOut)
	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "SKM - Enterprise Standalone Skills CLI Client")
}

func TestExecute_SearchNoArgs(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	code := Execute([]string{"search"}, out, errOut)
	assert.Equal(t, 1, code)
	assert.Contains(t, errOut.String(), "search query required")
}

func TestExecute_CompileFlags(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "skills", "comp-flag-skill")
	_ = os.MkdirAll(filepath.Join(skillDir, "references"), 0755)
	_ = os.MkdirAll(filepath.Join(skillDir, "examples"), 0755)
	_ = os.WriteFile(filepath.Join(skillDir, "references", "ref.md"), []byte("ref"), 0644)
	_ = os.WriteFile(filepath.Join(skillDir, "examples", "ex.md"), []byte("ex"), 0644)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: comp-flag-skill\n---\nBody"), 0644)

	outFile := filepath.Join(tmpDir, "out_manifest.json")
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	code := Execute([]string{"compile", "-d=" + tmpDir, "--output=" + outFile}, out, errOut)
	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "Successfully compiled")
}

func TestExecute_AddNoURIs(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	code := Execute([]string{"add"}, out, errOut)
	assert.Equal(t, 1, code)
	assert.Contains(t, errOut.String(), "missing skill URI")
}

func TestExecute_ValidateInvalidFlag(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	code := Execute([]string{"validate", "--invalid-flag"}, out, errOut)
	assert.Equal(t, 1, code)
	assert.Contains(t, errOut.String(), "Unknown flag: --invalid-flag")
}

func TestExecute_ListAndSearchEqualsFlags(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "skills", "eq-skill")
	_ = os.MkdirAll(skillDir, 0755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: eq-skill\ndescription: desc\n---\nBody with MatchWord"), 0644)

	// List -d=
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	code := Execute([]string{"list", "-d=" + tmpDir}, out, errOut)
	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "eq-skill")

	// Search -d=
	out2 := &bytes.Buffer{}
	errOut2 := &bytes.Buffer{}
	code = Execute([]string{"search", "MatchWord", "-d=" + tmpDir}, out2, errOut2)
	assert.Equal(t, 0, code)
	assert.Contains(t, out2.String(), "eq-skill")
}

func TestExecute_Completion(t *testing.T) {
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	// Bash
	code := Execute([]string{"completion", "bash"}, out, errOut)
	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "_skm_completions")

	// Zsh
	out2 := &bytes.Buffer{}
	code = Execute([]string{"completion", "zsh"}, out2, errOut)
	assert.Equal(t, 0, code)
	assert.Contains(t, out2.String(), "#compdef skm")

	// Fish
	out3 := &bytes.Buffer{}
	code = Execute([]string{"completion", "fish"}, out3, errOut)
	assert.Equal(t, 0, code)
	assert.Contains(t, out3.String(), "complete -c skm")

	// Unsupported
	errBuf := &bytes.Buffer{}
	code = Execute([]string{"completion", "invalid-shell"}, out, errBuf)
	assert.Equal(t, 1, code)
	assert.Contains(t, errBuf.String(), "Unsupported shell")
}

func TestExecute_InitFlags(t *testing.T) {
	tmpDir := t.TempDir()
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}

	code := Execute([]string{
		"init", "custom-flag-skill",
		"-d=" + tmpDir,
		"--description=Custom Description",
		"--license=MIT",
		"--author=Jane Doe",
		"--version=2.0.0",
	}, out, errOut)

	assert.Equal(t, 0, code)
	assert.Contains(t, out.String(), "Successfully initialized")

	skillMD, err := os.ReadFile(filepath.Join(tmpDir, "custom-flag-skill", "SKILL.md"))
	assert.NoError(t, err)
	assert.Contains(t, string(skillMD), "description: Custom Description")
	assert.Contains(t, string(skillMD), "license: MIT")
	assert.Contains(t, string(skillMD), "author: Jane Doe")
	assert.Contains(t, string(skillMD), "version: 2.0.0")
}

func TestExecute_Verify(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "verify-skill")
	_ = os.MkdirAll(skillDir, 0755)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: verify-skill\n---\nBody"), 0644)

	// Create .manifest.lock
	lockContent := `{
  "version": "1.0.0",
  "skills": {
    "verify-skill": {
      "skill_name": "verify-skill",
      "uri": "file://` + filepath.ToSlash(skillDir) + `",
      "sha256": "4a7c8e9b01d2e3f4a5b6c7d8e9f01a2b3c4d5e6f7a8b9c0d1e2f3a4b5c6d7e8f"
    }
  }
}`
	_ = os.WriteFile(filepath.Join(tmpDir, ".manifest.lock"), []byte(lockContent), 0644)

	// Verify text
	out := &bytes.Buffer{}
	errOut := &bytes.Buffer{}
	code := Execute([]string{"verify", "-d", tmpDir}, out, errOut)
	assert.Equal(t, 1, code) // Status is modified because sha256 doesn't match dummy
	assert.Contains(t, out.String(), "SKM Skill Integrity Verification Report")

	// Verify JSON
	out2 := &bytes.Buffer{}
	errOut2 := &bytes.Buffer{}
	code = Execute([]string{"verify", "--dir=" + tmpDir, "--json"}, out2, errOut2)
	assert.Equal(t, 1, code)
	assert.Contains(t, out2.String(), "\"target_dir\"")

	// Verify missing lockfile
	emptyDir := t.TempDir()
	out3 := &bytes.Buffer{}
	errOut3 := &bytes.Buffer{}
	code = Execute([]string{"verify", "-d=" + emptyDir}, out3, errOut3)
	assert.Equal(t, 1, code)
	assert.Contains(t, errOut3.String(), "Error verifying skills integrity")
}

