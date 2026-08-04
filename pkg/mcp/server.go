package mcp

import (
	"context"
	"encoding/json"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/retail-cortex/skills/pkg/data"
	"github.com/retail-cortex/skills/pkg/model"
	"github.com/retail-cortex/skills/pkg/service"
)

type MCPServer struct {
	server       *server.MCPServer
	appsService  *service.AppsService
	skillService *service.SkillsService
}

func NewMCPServer(appsSvc *service.AppsService, skillSvc *service.SkillsService) *MCPServer {
	s := server.NewMCPServer(
		"Skills Service MCP",
		"1.0.0",
	)

	mcpSvc := &MCPServer{
		server:       s,
		appsService:  appsSvc,
		skillService: skillSvc,
	}

	mcpSvc.registerTools()
	return mcpSvc
}

func (m *MCPServer) Server() *server.MCPServer {
	return m.server
}

func getArgString(req mcp.CallToolRequest, key string) string {
	if args, ok := req.Params.Arguments.(map[string]interface{}); ok {
		if val, ok := args[key].(string); ok {
			return val
		}
	}
	return ""
}

func (m *MCPServer) registerTools() {
	// 1. search_skills
	searchSkillsTool := mcp.NewTool("search_skills",
		mcp.WithDescription("Searches registered skills using Gemini semantic vector matching."),
		mcp.WithString("query", mcp.Description("Optional search query string")),
	)
	m.server.AddTool(searchSkillsTool, m.HandleSearchSkills)

	// 2. get_skill
	getSkillTool := mcp.NewTool("get_skill",
		mcp.WithDescription("Retrieves full details and compiled schema of a skill by ID or name."),
		mcp.WithString("skill_id_or_name", mcp.Required(), mcp.Description("Skill ID or unique skill name")),
	)
	m.server.AddTool(getSkillTool, m.HandleGetSkill)

	// 3. register_app
	registerAppTool := mcp.NewTool("register_app",
		mcp.WithDescription("Registers a new application and returns API key and verification link."),
		mcp.WithString("app_name", mcp.Required(), mcp.Description("Name of application")),
		mcp.WithString("email", mcp.Required(), mcp.Description("Contact email")),
	)
	m.server.AddTool(registerAppTool, m.HandleRegisterApp)

	// 4. verify_app
	verifyAppTool := mcp.NewTool("verify_app",
		mcp.WithDescription("Verifies and activates a registered application."),
		mcp.WithString("token", mcp.Required(), mcp.Description("Verification token")),
	)
	m.server.AddTool(verifyAppTool, m.HandleVerifyApp)

	// 5. register_skill
	registerSkillTool := mcp.NewTool("register_skill",
		mcp.WithDescription("Registers a new skill using an active application API key."),
		mcp.WithString("api_key", mcp.Required(), mcp.Description("Active application API Key")),
		mcp.WithString("name", mcp.Required(), mcp.Description("Skill name")),
		mcp.WithString("description", mcp.Required(), mcp.Description("Skill description")),
		mcp.WithString("instructions", mcp.Required(), mcp.Description("Skill instructions")),
		mcp.WithString("category", mcp.Description("Skill category")),
	)
	m.server.AddTool(registerSkillTool, m.HandleRegisterSkill)

	// 6. update_skill
	updateSkillTool := mcp.NewTool("update_skill",
		mcp.WithDescription("Updates an existing skill using an active application API key."),
		mcp.WithString("api_key", mcp.Required(), mcp.Description("Active application API Key")),
		mcp.WithString("skill_id", mcp.Required(), mcp.Description("Skill ID")),
		mcp.WithString("description", mcp.Description("Updated description")),
		mcp.WithString("instructions", mcp.Description("Updated instructions")),
	)
	m.server.AddTool(updateSkillTool, m.HandleUpdateSkill)

	// 7. delete_skill
	deleteSkillTool := mcp.NewTool("delete_skill",
		mcp.WithDescription("Deletes an existing skill using an active application API key."),
		mcp.WithString("api_key", mcp.Required(), mcp.Description("Active application API Key")),
		mcp.WithString("skill_id", mcp.Required(), mcp.Description("Skill ID")),
	)
	m.server.AddTool(deleteSkillTool, m.HandleDeleteSkill)
}

func (m *MCPServer) HandleSearchSkills(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query := getArgString(req, "query")
	db := data.GetDB()
	results, err := m.skillService.ListSkills(db, query)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	bytes, _ := json.Marshal(results)
	return mcp.NewToolResultText(string(bytes)), nil
}

func (m *MCPServer) HandleGetSkill(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	skillIDOrName := getArgString(req, "skill_id_or_name")
	db := data.GetDB()
	res, err := m.skillService.GetSkill(db, skillIDOrName)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	bytes, _ := json.Marshal(res)
	return mcp.NewToolResultText(string(bytes)), nil
}

func (m *MCPServer) HandleRegisterApp(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	appName := getArgString(req, "app_name")
	email := getArgString(req, "email")
	db := data.GetDB()
	res, err := m.appsService.RegisterApp(db, model.AppRegisterRequest{
		AppName: appName,
		Email:   email,
	}, "")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	bytes, _ := json.Marshal(res)
	return mcp.NewToolResultText(string(bytes)), nil
}

func (m *MCPServer) HandleVerifyApp(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	token := getArgString(req, "token")
	db := data.GetDB()
	res, err := m.appsService.VerifyApp(db, token)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	bytes, _ := json.Marshal(res)
	return mcp.NewToolResultText(string(bytes)), nil
}

func (m *MCPServer) HandleRegisterSkill(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiKey := getArgString(req, "api_key")
	name := getArgString(req, "name")
	description := getArgString(req, "description")
	instructions := getArgString(req, "instructions")
	category := getArgString(req, "category")

	db := data.GetDB()
	app, err := m.appsService.AuthenticateAPIKey(db, apiKey)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	cat := category
	if cat == "" {
		cat = "general"
	}

	res, err := m.skillService.CreateSkill(db, app.AppID, model.SkillCreateRequest{
		Name:         name,
		Description:  description,
		Instructions: instructions,
		Category:     &cat,
	})
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	bytes, _ := json.Marshal(res)
	return mcp.NewToolResultText(string(bytes)), nil
}

func (m *MCPServer) HandleUpdateSkill(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiKey := getArgString(req, "api_key")
	skillID := getArgString(req, "skill_id")

	db := data.GetDB()
	app, err := m.appsService.AuthenticateAPIKey(db, apiKey)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	updateReq := model.SkillUpdateRequest{}
	if desc := getArgString(req, "description"); desc != "" {
		updateReq.Description = &desc
	}
	if inst := getArgString(req, "instructions"); inst != "" {
		updateReq.Instructions = &inst
	}

	res, err := m.skillService.UpdateSkill(db, skillID, app.AppID, updateReq, false)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	bytes, _ := json.Marshal(res)
	return mcp.NewToolResultText(string(bytes)), nil
}

func (m *MCPServer) HandleDeleteSkill(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiKey := getArgString(req, "api_key")
	skillID := getArgString(req, "skill_id")

	db := data.GetDB()
	app, err := m.appsService.AuthenticateAPIKey(db, apiKey)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	res, err := m.skillService.DeleteSkill(db, skillID, app.AppID)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	bytes, _ := json.Marshal(res)
	return mcp.NewToolResultText(string(bytes)), nil
}
