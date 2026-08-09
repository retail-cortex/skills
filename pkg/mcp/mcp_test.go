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

package mcp_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/retail-cortex/skills/pkg/data"
	mcpServer "github.com/retail-cortex/skills/pkg/mcp"
	"github.com/retail-cortex/skills/pkg/model"
	"github.com/retail-cortex/skills/pkg/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMCPServerHandlers(t *testing.T) {
	data.ResetEngine()
	_, err := data.InitDB(filepath.Join(t.TempDir(), "test.db"))
	require.NoError(t, err)

	appsSvc := service.NewAppsService()
	skillsSvc := service.NewSkillsService()

	s := mcpServer.NewMCPServer(appsSvc, skillsSvc)
	require.NotNil(t, s)
	assert.NotNil(t, s.Server())

	ctx := context.Background()

	// 1. Test register_app
	regReq := mcp.CallToolRequest{}
	regReq.Params.Arguments = map[string]interface{}{
		"app_name": "mcp-app",
		"email":    "mcp@example.com",
	}
	regRes, err := s.HandleRegisterApp(ctx, regReq)
	require.NoError(t, err)
	assert.False(t, regRes.IsError)
	assert.NotEmpty(t, regRes.Content)

	var appRegResp model.AppRegisterResponse
	textObj, ok := regRes.Content[0].(mcp.TextContent)
	require.True(t, ok)
	err = json.Unmarshal([]byte(textObj.Text), &appRegResp)
	require.NoError(t, err)
	assert.NotEmpty(t, appRegResp.APIKey)

	// 2. Test verify_app
	verReq := mcp.CallToolRequest{}
	verReq.Params.Arguments = map[string]interface{}{
		"token": appRegResp.VerificationToken,
	}
	verRes, err := s.HandleVerifyApp(ctx, verReq)
	require.NoError(t, err)
	assert.False(t, verRes.IsError)

	// 3. Test register_skill
	regSkillReq := mcp.CallToolRequest{}
	regSkillReq.Params.Arguments = map[string]interface{}{
		"api_key":      appRegResp.APIKey,
		"name":         "mcp-skill",
		"description":  "MCP description",
		"instructions": "MCP instructions",
		"category":     "mcp-cat",
	}
	regSkillRes, err := s.HandleRegisterSkill(ctx, regSkillReq)
	require.NoError(t, err)
	assert.False(t, regSkillRes.IsError)

	var skillResp model.SkillResponse
	textObj2, ok := regSkillRes.Content[0].(mcp.TextContent)
	require.True(t, ok)
	err = json.Unmarshal([]byte(textObj2.Text), &skillResp)
	require.NoError(t, err)
	assert.Equal(t, "mcp-skill", skillResp.Name)

	// 4. Test search_skills
	searchReq := mcp.CallToolRequest{}
	searchReq.Params.Arguments = map[string]interface{}{
		"query": "mcp-skill",
	}
	searchRes, err := s.HandleSearchSkills(ctx, searchReq)
	require.NoError(t, err)
	assert.False(t, searchRes.IsError)

	// 5. Test get_skill
	getReq := mcp.CallToolRequest{}
	getReq.Params.Arguments = map[string]interface{}{
		"skill_id_or_name": skillResp.ID,
	}
	getRes, err := s.HandleGetSkill(ctx, getReq)
	require.NoError(t, err)
	assert.False(t, getRes.IsError)

	// 6. Test update_skill
	updReq := mcp.CallToolRequest{}
	updReq.Params.Arguments = map[string]interface{}{
		"api_key":     appRegResp.APIKey,
		"skill_id":    skillResp.ID,
		"description": "Updated MCP desc",
	}
	updRes, err := s.HandleUpdateSkill(ctx, updReq)
	require.NoError(t, err)
	assert.False(t, updRes.IsError)

	// 7. Test delete_skill
	delReq := mcp.CallToolRequest{}
	delReq.Params.Arguments = map[string]interface{}{
		"api_key":  appRegResp.APIKey,
		"skill_id": skillResp.ID,
	}
	delRes, err := s.HandleDeleteSkill(ctx, delReq)
	require.NoError(t, err)
	assert.False(t, delRes.IsError)

	// 8. Test error paths
	badKeyReq := mcp.CallToolRequest{}
	badKeyReq.Params.Arguments = map[string]interface{}{
		"api_key":  "invalid-key",
		"skill_id": skillResp.ID,
	}
	badKeyRes, err := s.HandleDeleteSkill(ctx, badKeyReq)
	require.NoError(t, err)
	assert.True(t, badKeyRes.IsError)
}
