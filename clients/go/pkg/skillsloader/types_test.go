package skillsloader

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSkillDefinitionToMap(t *testing.T) {
	skill := &SkillDefinition{
		Name:          "go-test-skill",
		Description:   "Go client test skill",
		Instructions:  "Run hermetic tests.",
		License:       "Apache-2.0",
		Author:        "Retail Cortex",
		Version:       "1.0.0",
		Compatibility: "ADK-v1",
		Metadata: map[string]string{
			"custom_key": "custom_val",
		},
		References: map[string]string{
			"ref1.md": "Reference 1",
		},
		Examples: map[string]string{
			"ex1.go": "Example 1",
		},
		Path: "/path/to/skill",
	}

	m := skill.ToMap()
	assert.Equal(t, "go-test-skill", m["name"])
	assert.Equal(t, "Go client test skill", m["description"])
	assert.Equal(t, "Run hermetic tests.", m["instructions"])
	assert.Equal(t, "Apache-2.0", m["license"])
	assert.Equal(t, "Retail Cortex", m["author"])
	assert.Equal(t, "1.0.0", m["version"])
	assert.Equal(t, "ADK-v1", m["compatibility"])
	assert.Equal(t, []string{"ref1.md"}, m["references"])
	assert.Equal(t, []string{"ex1.go"}, m["examples"])
	assert.Equal(t, "/path/to/skill", m["path"])
}
