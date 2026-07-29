package skillsloader

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindRegistryRoot(t *testing.T) {
	root := FindRegistryRoot()
	assert.NotEmpty(t, root)
	assert.True(t, isDir(filepath.Join(root, "packages")) || isDir(filepath.Join(root, "skills")))
}

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		expectedFm   map[string]string
		expectedBody string
	}{
		{
			name: "valid frontmatter",
			content: "---\nname: test-skill\ndescription: A test description\n---\n# Instructions\nFollow these rules.",
			expectedFm: map[string]string{
				"name":        "test-skill",
				"description": "A test description",
			},
			expectedBody: "# Instructions\nFollow these rules.",
		},
		{
			name:         "empty frontmatter",
			content:      "# No Frontmatter\nJust body content.",
			expectedFm:   map[string]string{},
			expectedBody: "# No Frontmatter\nJust body content.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, body := ParseFrontmatter(tt.content)
			assert.Equal(t, tt.expectedFm, fm)
			assert.Equal(t, tt.expectedBody, body)
		})
	}
}

func TestParseDotenvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envPath := filepath.Join(tmpDir, ".env")
	content := "GITHUB_TOKEN=ghp_secret_123\nGITHUB_REF=v2.1.0\nSKILLS_ROOTS=skills,custom_skills\n# Comment\n"
	err := os.WriteFile(envPath, []byte(content), 0644)
	require.NoError(t, err)

	parsed := ParseDotenvFile(envPath)
	assert.Equal(t, "ghp_secret_123", parsed["GITHUB_TOKEN"])
	assert.Equal(t, "v2.1.0", parsed["GITHUB_REF"])
	assert.Equal(t, "skills,custom_skills", parsed["SKILLS_ROOTS"])

	// Test non-existent file
	bogusParsed := ParseDotenvFile(filepath.Join(tmpDir, "nonexistent.env"))
	assert.Empty(t, bogusParsed)
}

func TestParseSkillRootURI(t *testing.T) {
	tests := []struct {
		name       string
		uri        string
		reqScheme  string
		reqTarget  string
		reqRef     string
		reqSubpath string
	}{
		{
			name:       "file URI",
			uri:        "file://skills/custom",
			reqScheme:  "file",
			reqTarget:  "skills/custom",
			reqRef:     "",
			reqSubpath: "",
		},
		{
			name:       "package URI",
			uri:        "pkg://retailcortex_skills_python",
			reqScheme:  "pkg",
			reqTarget:  "retailcortex_skills_python",
			reqRef:     "",
			reqSubpath: "",
		},
		{
			name:       "maven URI",
			uri:        "maven://com.retailcortex.skills:skills-java:1.0.0",
			reqScheme:  "maven",
			reqTarget:  "com.retailcortex.skills:skills-java:1.0.0",
			reqRef:     "1.0.0",
			reqSubpath: "",
		},
		{
			name:       "mvn URI with subpath",
			uri:        "mvn://com.retailcortex.skills:skills-java:1.0.0/java-enterprise",
			reqScheme:  "maven",
			reqTarget:  "com.retailcortex.skills:skills-java:1.0.0",
			reqRef:     "1.0.0",
			reqSubpath: "java-enterprise",
		},
		{
			name:       "mod URI with version and subpath",
			uri:        "mod://github.com/retail-cortex/skills@v1.0.0/packages/skills-go",
			reqScheme:  "mod",
			reqTarget:  "github.com/retail-cortex/skills",
			reqRef:     "v1.0.0",
			reqSubpath: "packages/skills-go",
		},
		{
			name:       "github URI with trailing ref",
			uri:        "github://google/skills/skills/cloud/gemini-api:main",
			reqScheme:  "github",
			reqTarget:  "google/skills",
			reqRef:     "main",
			reqSubpath: "skills/cloud/gemini-api",
		},
		{
			name:       "github URI with repo:version",
			uri:        "github://owner/repo:v2.5.0",
			reqScheme:  "github",
			reqTarget:  "owner/repo",
			reqRef:     "v2.5.0",
			reqSubpath: "",
		},
		{
			name:       "github URI with @ref/subpath",
			uri:        "github://owner/repo@v1.2.0/skills/cloud",
			reqScheme:  "github",
			reqTarget:  "owner/repo",
			reqRef:     "v1.2.0",
			reqSubpath: "skills/cloud",
		},
		{
			name:       "github URI with web tree path",
			uri:        "github://google/skills/tree/main/skills/cloud/gemini-api",
			reqScheme:  "github",
			reqTarget:  "google/skills",
			reqRef:     "main",
			reqSubpath: "skills/cloud/gemini-api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, target, r, sub := ParseSkillRootURI(tt.uri)
			assert.Equal(t, tt.reqScheme, s)
			assert.Equal(t, tt.reqTarget, target)
			assert.Equal(t, tt.reqRef, r)
			assert.Equal(t, tt.reqSubpath, sub)
		})
	}
}

func TestLoadAllSkills(t *testing.T) {
	skills, err := LoadAllSkills("", nil)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(skills), 20)
	assert.Contains(t, skills, "python-core")
	assert.Contains(t, skills, "go-lang")
}

func TestSkillRegistryQueries(t *testing.T) {
	registry, err := NewSkillRegistry("", nil, nil, "")
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(registry.Skills()), 20)

	pythonSkill := registry.Get("python-core")
	require.NotNil(t, pythonSkill)
	assert.Equal(t, "python-core", pythonSkill.Name)
	assert.NotEmpty(t, pythonSkill.Instructions)

	// Test Search
	goMatches := registry.Search("Go")
	assert.NotEmpty(t, goMatches)

	// Test Domain Filtering
	domainMatches := registry.GetDomainSkills("fastapi")
	assert.NotEmpty(t, domainMatches)

	// Test ListSkills Summary
	summaries := registry.ListSkills()
	assert.GreaterOrEqual(t, len(summaries), 20)
}

func TestFrontmatterMetadataMapping(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "test-meta-skill")
	err := os.MkdirAll(skillDir, 0755)
	require.NoError(t, err)

	refDir := filepath.Join(skillDir, "references")
	err = os.MkdirAll(refDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(refDir, "ref1.md"), []byte("Reference 1 Content"), 0644)
	require.NoError(t, err)

	content := "---\nname: test-meta-skill\ndescription: Meta test\nlicense: MIT\nauthor: Jane Doe\nversion: 2.0.0\ncompatibility: ADK-v2\nallowed-tools: Bash(git:*) Read\ncustom_field: custom_val\n---\n# Instructions"
	err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0644)
	require.NoError(t, err)

	skill, err := LoadSkillFromDir(skillDir)
	require.NoError(t, err)
	require.NotNil(t, skill)

	assert.Equal(t, "MIT", skill.License)
	assert.Equal(t, "Jane Doe", skill.Author)
	assert.Equal(t, "2.0.0", skill.Version)
	assert.Equal(t, "ADK-v2", skill.Compatibility)
	assert.Equal(t, "Bash(git:*) Read", skill.AllowedTools)
	assert.Equal(t, "custom_val", skill.Metadata["custom_field"])
	assert.Equal(t, "Jane Doe", skill.Metadata["author"])
	assert.Equal(t, "Reference 1 Content", skill.GetReferenceContent("ref1.md"))
	assert.Equal(t, "", skill.GetReferenceContent("nonexistent.md"))
}

func TestAgentsSkillsDirectoryScanning(t *testing.T) {
	tmpDir := t.TempDir()
	agDir := filepath.Join(tmpDir, ".agents", "skills", "ag-go-skill")
	err := os.MkdirAll(agDir, 0755)
	require.NoError(t, err)

	content := "---\nname: ag-go-skill\ndescription: Cross client agent skill for Go\n---\n# Instructions"
	err = os.WriteFile(filepath.Join(agDir, "SKILL.md"), []byte(content), 0644)
	require.NoError(t, err)

	skills, err := LoadAllSkills(tmpDir, nil)
	require.NoError(t, err)
	assert.Contains(t, skills, "ag-go-skill")
	assert.Equal(t, "Cross client agent skill for Go", skills["ag-go-skill"].Description)
}

func TestBuildAndLoadManifest(t *testing.T) {
	tmpDir := t.TempDir()
	outJSON := filepath.Join(tmpDir, "skills_manifest.json")

	genFile, err := BuildSkillsManifest("", outJSON)
	require.NoError(t, err)
	assert.FileExists(t, genFile)

	loaded, err := LoadSkillsFromManifest(genFile)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(loaded), 20)
	assert.Contains(t, loaded, "python-core")
}

func TestSkillRegistryFromRootsFactory(t *testing.T) {
	regFile, err := SkillRegistryFromRoots(
		[]string{"file://."},
		[]string{"python-core"},
		"", "",
	)
	require.NoError(t, err)
	assert.Contains(t, regFile.Skills(), "python-core")
}

func SkillRegistryFromRoots(roots []string, filter []string, token string, dotenvPath string) (*SkillRegistry, error) {
	return NewSkillRegistryFromRoots(roots, filter, token, dotenvPath)
}

func TestSkillDefinitionToMap_Extra(t *testing.T) {
	def := &SkillDefinition{
		Name:          "test-skill",
		Description:   "Description",
		Instructions:  "Instructions",
		License:       "Apache-2.0",
		Author:        "Author",
		Version:       "1.0",
		Compatibility: "ADK-v1",
		AllowedTools:  "Read",
		Path:          "/path/to/skill",
		Metadata:      map[string]string{"foo": "bar"},
		References:    map[string]string{"ref.md": "ref content"},
		Examples:      map[string]string{"ex.py": "ex content"},
	}
	m := def.ToMap()
	assert.Equal(t, "test-skill", m["name"])
	assert.Equal(t, "Description", m["description"])
	assert.Equal(t, "Instructions", m["instructions"])
	assert.Equal(t, "Apache-2.0", m["license"])
}

func TestLoadSkillsFromPackage(t *testing.T) {
	skills, err := LoadSkillsFromPackage("retailcortex_skills_python", []string{"python-core"})
	require.NoError(t, err)
	assert.Contains(t, skills, "python-core")

	emptySkills, err := LoadSkillsFromPackage("", nil)
	require.NoError(t, err)
	assert.Empty(t, emptySkills)
}

func TestLoadSkillsFromGoModule_LocalCache(t *testing.T) {
	skills, err := LoadSkillsFromGoModule("github.com/retail-cortex/skills", "v1.0.0", nil, []string{"python-core"})
	if err == nil && len(skills) > 0 {
		assert.Contains(t, skills, "python-core")
	}

	empty, err := LoadSkillsFromGoModule("", "", nil, nil)
	require.NoError(t, err)
	assert.Empty(t, empty)
}

func TestLoadSkillsFromMaven_LocalCache(t *testing.T) {
	tmpDir := t.TempDir()
	m2Dir := filepath.Join(tmpDir, ".m2", "repository", "com", "test", "my-artifact", "1.0.0")
	_ = os.MkdirAll(m2Dir, 0755)
	jarFile := filepath.Join(m2Dir, "my-artifact-1.0.0.jar")

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	f, _ := zw.Create("skills/maven-skill/SKILL.md")
	_, _ = f.Write([]byte("---\nname: maven-skill\n---\nBody"))
	_ = zw.Close()
	_ = os.WriteFile(jarFile, buf.Bytes(), 0644)

	t.Setenv("HOME", tmpDir)

	skills, err := LoadSkillsFromMaven("com.test:my-artifact:1.0.0", "", nil, nil)
	assert.NoError(t, err)
	assert.Contains(t, skills, "maven-skill")
}

func TestLoadSkillsFromManifest_Errors(t *testing.T) {
	tmpDir := t.TempDir()
	badJSON := filepath.Join(tmpDir, "bad.json")
	_ = os.WriteFile(badJSON, []byte("{invalid json"), 0644)

	_, err := LoadSkillsFromManifest(badJSON)
	assert.Error(t, err)

	_, err = LoadSkillsFromManifest(filepath.Join(tmpDir, "nonexistent.json"))
	assert.Error(t, err)
}

func TestLoadSkillFromDir_EdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	// Dir without SKILL.md
	nonSkillDir := filepath.Join(tmpDir, "no-skill")
	_ = os.MkdirAll(nonSkillDir, 0755)
	s, err := LoadSkillFromDir(nonSkillDir)
	assert.Error(t, err)
	assert.Nil(t, s)

	// Skill with examples directory
	exSkillDir := filepath.Join(tmpDir, "ex-skill")
	_ = os.MkdirAll(filepath.Join(exSkillDir, "examples"), 0755)
	_ = os.WriteFile(filepath.Join(exSkillDir, "examples", "code.py"), []byte("print('hello')"), 0644)
	_ = os.WriteFile(filepath.Join(exSkillDir, "SKILL.md"), []byte("---\nname: ex-skill\n---\nBody"), 0644)

	skill, err := LoadSkillFromDir(exSkillDir)
	require.NoError(t, err)
	assert.Equal(t, "print('hello')", skill.GetExampleContent("code.py"))
	assert.Equal(t, "", skill.GetExampleContent("nonexistent.py"))
}

func TestLoadSkillsFromGoModule_GOPATH(t *testing.T) {
	tmpDir := t.TempDir()
	modDir := filepath.Join(tmpDir, "pkg", "mod", "github.com", "mock-owner", "mock-mod@v1.0.0")
	skillDir := filepath.Join(modDir, "skills", "mod-skill")
	_ = os.MkdirAll(filepath.Join(skillDir, "references"), 0755)
	_ = os.MkdirAll(filepath.Join(skillDir, "examples"), 0755)
	_ = os.WriteFile(filepath.Join(skillDir, "references", "ref.md"), []byte("ref"), 0644)
	_ = os.WriteFile(filepath.Join(skillDir, "examples", "ex.md"), []byte("ex"), 0644)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: mod-skill\n---\nBody"), 0644)

	t.Setenv("GOPATH", tmpDir)

	skills, err := LoadSkillsFromGoModule("github.com/mock-owner/mock-mod", "v1.0.0", nil, nil)
	assert.NoError(t, err)
	assert.Contains(t, skills, "mod-skill")
}

func TestLoadSkillsFromGitHub_Cached(t *testing.T) {
	loaderBase := GetLoaderSkillsDir()
	ghDir := filepath.Join(loaderBase, "github", "mock-owner_mock-repo", "main")
	_ = os.MkdirAll(ghDir, 0755)
	defer os.RemoveAll(filepath.Join(loaderBase, "github", "mock-owner_mock-repo"))

	_ = os.WriteFile(filepath.Join(ghDir, "dummy_file.txt"), []byte("data"), 0644)

	skillDir := filepath.Join(ghDir, "skills", "gh-skill")
	_ = os.MkdirAll(filepath.Join(skillDir, "references"), 0755)
	_ = os.MkdirAll(filepath.Join(skillDir, "examples"), 0755)
	_ = os.WriteFile(filepath.Join(skillDir, "references", "ref.md"), []byte("ref"), 0644)
	_ = os.WriteFile(filepath.Join(skillDir, "examples", "ex.md"), []byte("ex"), 0644)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: gh-skill\n---\nBody"), 0644)

	skills, err := LoadSkillsFromGitHub("mock-owner/mock-repo", "main", nil, nil, "", "")
	assert.NoError(t, err)
	assert.Contains(t, skills, "gh-skill")
}

func TestLoadSkillsFromRoots_MultipleSchemes(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "root-skill")
	_ = os.MkdirAll(filepath.Join(skillDir, "references"), 0755)
	_ = os.MkdirAll(filepath.Join(skillDir, "examples"), 0755)
	_ = os.WriteFile(filepath.Join(skillDir, "references", "ref.md"), []byte("ref"), 0644)
	_ = os.WriteFile(filepath.Join(skillDir, "examples", "ex.md"), []byte("ex"), 0644)
	_ = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte("---\nname: root-skill\n---\nBody"), 0644)

	roots := []string{
		"file://" + skillDir,
		"pkg://retailcortex_skills_python",
		"mod://github.com/mock-owner/mock-mod@v1.0.0",
		"maven://com.test:my-artifact:1.0.0",
		"github://mock-owner/mock-repo@main",
	}

	skills, err := LoadSkillsFromRoots(roots, nil, "", "")
	assert.NoError(t, err)
	assert.Contains(t, skills, "root-skill")
}
