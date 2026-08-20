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

package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/retail-cortex/skills/pkg/data"
	"github.com/retail-cortex/skills/pkg/model"
	"github.com/retail-cortex/skills/pkg/service"
)

type ServerHandlers struct {
	appsService   *service.AppsService
	skillsService *service.SkillsService
}

func NewServerHandlers(appsSvc *service.AppsService, skillsSvc *service.SkillsService) *ServerHandlers {
	return &ServerHandlers{
		appsService:   appsSvc,
		skillsService: skillsSvc,
	}
}

func (h *ServerHandlers) HealthCheck(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "castor-registry",
			"version": "0.1.0",
			"ports": gin.H{
				"rest": cfg.Port,
			},
		})
	}
}

func (h *ServerHandlers) RegisterApp(c *gin.Context) {
	var req model.AppRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

	db := data.GetDB()
	res, err := h.appsService.RegisterApp(db, req, baseURL)
	if err != nil {
		if errors.Is(err, data.ErrAppAlreadyRegistered) || errors.Is(err, data.ErrInvalidRegistrationEmail) {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}
		if errors.Is(err, data.ErrFreemailDomainProhibited) {
			c.JSON(http.StatusForbidden, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *ServerHandlers) VerifyApp(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Missing verification token parameter."})
		return
	}

	db := data.GetDB()
	res, err := h.appsService.VerifyApp(db, token)
	if err != nil {
		if errors.Is(err, data.ErrInvalidVerificationToken) {
			c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *ServerHandlers) authenticateContext(c *gin.Context, requiredRole model.AppRole) (*model.AuthContext, bool) {
	apiKey := c.GetHeader("X-API-Key")
	db := data.GetDB()
	authCtx, err := h.appsService.AuthenticateContext(db, apiKey)
	if err != nil {
		if errors.Is(err, data.ErrMissingAPIKey) || errors.Is(err, data.ErrInvalidAPIKey) {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": err.Error()})
			return nil, false
		}
		if errors.Is(err, data.ErrAppNotVerified) {
			c.JSON(http.StatusForbidden, gin.H{"detail": err.Error()})
			return nil, false
		}
		c.JSON(http.StatusUnauthorized, gin.H{"detail": err.Error()})
		return nil, false
	}

	if requiredRole != "" && !authCtx.Role.HasPermission(requiredRole) {
		c.JSON(http.StatusForbidden, gin.H{"detail": "insufficient role permission for this operation"})
		return nil, false
	}

	return authCtx, true
}

func (h *ServerHandlers) authenticateCurrentApp(c *gin.Context) (*model.RegisteredApp, bool) {
	authCtx, ok := h.authenticateContext(c, model.RoleViewer)
	if !ok {
		return nil, false
	}
	return authCtx.App, true
}

// ----------------------------------------------------------------------------
// RBAC Collaborator Endpoints
// ----------------------------------------------------------------------------

func (h *ServerHandlers) ListMembers(c *gin.Context) {
	authCtx, ok := h.authenticateContext(c, model.RoleViewer)
	if !ok {
		return
	}

	db := data.GetDB()
	members, err := h.appsService.ListMembers(db, authCtx.App.AppID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, members)
}

func (h *ServerHandlers) InviteMember(c *gin.Context) {
	authCtx, ok := h.authenticateContext(c, model.RoleOwner)
	if !ok {
		return
	}

	var req model.MemberInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

	db := data.GetDB()
	res, err := h.appsService.InviteMember(db, authCtx.App.AppID, authCtx.MemberEmail, req.Email, req.Role, baseURL)
	if err != nil {
		if errors.Is(err, data.ErrMemberAlreadyExists) || errors.Is(err, data.ErrInvalidRegistrationEmail) {
			c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *ServerHandlers) AcceptInvitation(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		var req model.MemberAcceptRequest
		if err := c.ShouldBindJSON(&req); err == nil {
			token = req.Token
		}
	}

	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"detail": "Missing invitation token."})
		return
	}

	db := data.GetDB()
	member, err := h.appsService.AcceptInvitation(db, token)
	if err != nil {
		if errors.Is(err, data.ErrInvalidInvitationToken) {
			c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invitation accepted successfully. You are now an active collaborator.",
		"member":  member,
	})
}

func (h *ServerHandlers) UpdateMemberRole(c *gin.Context) {
	authCtx, ok := h.authenticateContext(c, model.RoleOwner)
	if !ok {
		return
	}

	memberID := c.Param("member_id")
	var req model.MemberUpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	db := data.GetDB()
	updated, err := h.appsService.UpdateMemberRole(db, authCtx.App.AppID, memberID, req.Role)
	if err != nil {
		if errors.Is(err, data.ErrMemberNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *ServerHandlers) RemoveMember(c *gin.Context) {
	authCtx, ok := h.authenticateContext(c, model.RoleOwner)
	if !ok {
		return
	}

	memberID := c.Param("member_id")
	db := data.GetDB()
	err := h.appsService.RemoveMember(db, authCtx.App.AppID, memberID)
	if err != nil {
		if errors.Is(err, data.ErrMemberNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Collaborator removed successfully."})
}

// ----------------------------------------------------------------------------
// Scoped API Key Endpoints
// ----------------------------------------------------------------------------

func (h *ServerHandlers) ListAPIKeys(c *gin.Context) {
	authCtx, ok := h.authenticateContext(c, model.RoleOwner)
	if !ok {
		return
	}

	db := data.GetDB()
	keys, err := h.appsService.ListAPIKeys(db, authCtx.App.AppID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, keys)
}

func (h *ServerHandlers) CreateScopedAPIKey(c *gin.Context) {
	authCtx, ok := h.authenticateContext(c, model.RoleEditor)
	if !ok {
		return
	}

	var req model.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	// Non-owners cannot create keys with higher permissions than their own
	if authCtx.Role != model.RoleOwner && req.Role == model.RoleOwner {
		c.JSON(http.StatusForbidden, gin.H{"detail": "Cannot provision API key with higher role than caller"})
		return
	}

	db := data.GetDB()
	res, err := h.appsService.CreateScopedAPIKey(db, authCtx.App.AppID, authCtx.MemberEmail, req.Name, req.Role, req.ExpiresInDays)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *ServerHandlers) RevokeAPIKey(c *gin.Context) {
	authCtx, ok := h.authenticateContext(c, model.RoleOwner)
	if !ok {
		return
	}

	keyID := c.Param("key_id")
	db := data.GetDB()
	err := h.appsService.RevokeAPIKey(db, authCtx.App.AppID, keyID)
	if err != nil {
		if errors.Is(err, data.ErrKeyNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key revoked successfully."})
}

// ----------------------------------------------------------------------------
// Skills Endpoints
// ----------------------------------------------------------------------------

func (h *ServerHandlers) ListSkills(c *gin.Context) {
	query := c.Query("s")

	page := 1
	if pStr := c.Query("page"); pStr != "" {
		if p, err := strconv.Atoi(pStr); err == nil && p >= 1 {
			page = p
		}
	}

	pageSize := 5
	for _, k := range []string{"page_size", "limit", "max", "max_results", "size"} {
		if psStr := c.Query(k); psStr != "" {
			if ps, err := strconv.Atoi(psStr); err == nil && ps >= 1 {
				pageSize = ps
				break
			}
		}
	}

	pagination := model.PaginationParams{
		Page:     page,
		PageSize: pageSize,
	}

	db := data.GetDB()
	res, err := h.skillsService.ListSkills(db, query, pagination)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}

	// Set standard HTTP pagination headers
	c.Header("X-Total-Count", strconv.FormatInt(res.TotalCount, 10))
	c.Header("X-Page", strconv.Itoa(res.Page))
	c.Header("X-Page-Size", strconv.Itoa(res.PageSize))
	c.Header("X-Total-Pages", strconv.Itoa(res.TotalPages))

	// If format=envelope or envelope=true, return full pagination JSON envelope
	if c.Query("envelope") == "true" || c.Query("format") == "envelope" {
		c.JSON(http.StatusOK, res)
		return
	}

	// Default JSON response is the paged slice of skills
	c.JSON(http.StatusOK, res.Items)
}

func (h *ServerHandlers) GetSkill(c *gin.Context) {
	skillIDOrName := strings.TrimPrefix(c.Param("skill_id_or_name"), "/")
	db := data.GetDB()
	res, err := h.skillsService.GetSkill(db, skillIDOrName)
	if err != nil {
		if errors.Is(err, data.ErrSkillNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ServerHandlers) RegisterSkill(c *gin.Context) {
	authCtx, ok := h.authenticateContext(c, model.RoleEditor)
	if !ok {
		return
	}

	var req model.SkillCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	db := data.GetDB()
	res, err := h.skillsService.CreateSkill(db, authCtx.App.AppID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *ServerHandlers) ReplaceSkill(c *gin.Context) {
	authCtx, ok := h.authenticateContext(c, model.RoleEditor)
	if !ok {
		return
	}

	skillID := strings.TrimPrefix(c.Param("skill_id"), "/")
	var req model.SkillUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	db := data.GetDB()
	res, err := h.skillsService.UpdateSkill(db, skillID, authCtx.App.AppID, req, true)
	if err != nil {
		if errors.Is(err, data.ErrSkillNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
			return
		}
		if errors.Is(err, data.ErrUnauthorizedSkillAccess) {
			c.JSON(http.StatusForbidden, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ServerHandlers) UpdateSkill(c *gin.Context) {
	authCtx, ok := h.authenticateContext(c, model.RoleEditor)
	if !ok {
		return
	}

	skillID := strings.TrimPrefix(c.Param("skill_id"), "/")
	var req model.SkillUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	db := data.GetDB()
	res, err := h.skillsService.UpdateSkill(db, skillID, authCtx.App.AppID, req, false)
	if err != nil {
		if errors.Is(err, data.ErrSkillNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
			return
		}
		if errors.Is(err, data.ErrUnauthorizedSkillAccess) {
			c.JSON(http.StatusForbidden, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ServerHandlers) DeleteSkill(c *gin.Context) {
	authCtx, ok := h.authenticateContext(c, model.RoleEditor)
	if !ok {
		return
	}

	skillID := strings.TrimPrefix(c.Param("skill_id"), "/")
	db := data.GetDB()
	res, err := h.skillsService.DeleteSkill(db, skillID, authCtx.App.AppID)
	if err != nil {
		if errors.Is(err, data.ErrSkillNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"detail": err.Error()})
			return
		}
		if errors.Is(err, data.ErrUnauthorizedSkillAccess) {
			c.JSON(http.StatusForbidden, gin.H{"detail": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}
