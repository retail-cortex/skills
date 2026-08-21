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

package castor_client

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
