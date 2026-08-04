package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/retail-cortex/skills/pkg/data"
	"github.com/retail-cortex/skills/pkg/model"
	"github.com/retail-cortex/skills/pkg/service"
)

type ServerHandlers struct {
	appsService  *service.AppsService
	skillsService *service.SkillsService
}

func NewServerHandlers(appsSvc *service.AppsService, skillsSvc *service.SkillsService) *ServerHandlers {
	return &ServerHandlers{
		appsService:  appsSvc,
		skillsService: skillsSvc,
	}
}

func (h *ServerHandlers) HealthCheck(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "skills-service",
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

func (h *ServerHandlers) authenticateCurrentApp(c *gin.Context) (*model.RegisteredApp, bool) {
	apiKey := c.GetHeader("X-API-Key")
	db := data.GetDB()
	app, err := h.appsService.AuthenticateAPIKey(db, apiKey)
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
	return app, true
}

func (h *ServerHandlers) ListSkills(c *gin.Context) {
	query := c.Query("s")
	db := data.GetDB()
	res, err := h.skillsService.ListSkills(db, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusOK, res)
}

func (h *ServerHandlers) GetSkill(c *gin.Context) {
	skillIDOrName := c.Param("skill_id_or_name")
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
	app, ok := h.authenticateCurrentApp(c)
	if !ok {
		return
	}

	var req model.SkillCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	db := data.GetDB()
	res, err := h.skillsService.CreateSkill(db, app.AppID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"detail": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, res)
}

func (h *ServerHandlers) ReplaceSkill(c *gin.Context) {
	app, ok := h.authenticateCurrentApp(c)
	if !ok {
		return
	}

	skillID := c.Param("skill_id")
	var req model.SkillUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	db := data.GetDB()
	res, err := h.skillsService.UpdateSkill(db, skillID, app.AppID, req, true)
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
	app, ok := h.authenticateCurrentApp(c)
	if !ok {
		return
	}

	skillID := c.Param("skill_id")
	var req model.SkillUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	db := data.GetDB()
	res, err := h.skillsService.UpdateSkill(db, skillID, app.AppID, req, false)
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
	app, ok := h.authenticateCurrentApp(c)
	if !ok {
		return
	}

	skillID := c.Param("skill_id")
	db := data.GetDB()
	res, err := h.skillsService.DeleteSkill(db, skillID, app.AppID)
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
