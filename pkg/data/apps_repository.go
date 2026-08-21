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

package data

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/retail-cortex/castor/pkg/model"
	"gorm.io/gorm"
)

var (
	ErrAppAlreadyRegistered     = errors.New("application with email is already registered")
	ErrInvalidVerificationToken = errors.New("invalid or expired verification token")
	ErrMissingAPIKey            = errors.New("missing API key header X-API-Key")
	ErrInvalidAPIKey            = errors.New("invalid API key provided")
	ErrAppNotVerified           = errors.New("application is pending email verification")
	ErrFreemailDomainProhibited = errors.New("freemail accounts (e.g. @gmail.com) cannot claim enterprise domain names")
	ErrInvalidRegistrationEmail = errors.New("invalid email address format for registration")
	ErrMemberAlreadyExists      = errors.New("collaborator already exists in this application")
	ErrMemberNotFound           = errors.New("collaborator not found")
	ErrKeyNotFound              = errors.New("scoped API key not found")
	ErrInsufficientPermission   = errors.New("insufficient role permission for this operation")
	ErrInvalidInvitationToken   = errors.New("invalid or expired invitation token")
)

var freemailDomains = map[string]bool{
	"gmail.com":      true,
	"yahoo.com":      true,
	"hotmail.com":    true,
	"outlook.com":    true,
	"icloud.com":     true,
	"aol.com":        true,
	"protonmail.com": true,
	"zoho.com":       true,
}

// AppRepository defines the contract for application registration, verification, and multi-user RBAC.
type AppRepository interface {
	HashAPIKey(apiKey string) string
	RegisterApp(db *gorm.DB, req model.AppRegisterRequest, baseURL string) (*model.AppRegisterResponse, error)
	VerifyApp(db *gorm.DB, token string) (*model.AppVerifyResponse, error)
	AuthenticateAPIKey(db *gorm.DB, apiKey string) (*model.RegisteredApp, error)
	AuthenticateContext(db *gorm.DB, apiKey string) (*model.AuthContext, error)

	// RBAC Collaborator Management
	InviteMember(db *gorm.DB, appID, inviterEmail, targetEmail string, role model.AppRole, baseURL string) (*model.MemberInviteResponse, error)
	AcceptInvitation(db *gorm.DB, token string) (*model.AppMember, error)
	ListMembers(db *gorm.DB, appID string) ([]model.AppMember, error)
	UpdateMemberRole(db *gorm.DB, appID, memberID string, newRole model.AppRole) (*model.AppMember, error)
	RemoveMember(db *gorm.DB, appID, memberID string) error

	// Scoped API Key Management
	CreateScopedAPIKey(db *gorm.DB, appID, memberEmail, keyName string, role model.AppRole, expiresInDays int) (*model.CreateAPIKeyResponse, error)
	ListAPIKeys(db *gorm.DB, appID string) ([]model.AppAPIKeySummary, error)
	RevokeAPIKey(db *gorm.DB, appID, keyID string) error
}

type AppsRepository struct{}

func NewAppsRepository() *AppsRepository {
	return &AppsRepository{}
}

// HashAPIKey computes SHA-256 hash of API key for secure DB storage.
func (r *AppsRepository) HashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}

// generateAPIKey creates a secure random API key string cstr_live_...
func (r *AppsRepository) generateAPIKey() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("cstr_live_%s", hex.EncodeToString(bytes)), nil
}

// RegisterApp registers a new application with domain scoping and seeds the creator as OWNER.
func (r *AppsRepository) RegisterApp(db *gorm.DB, req model.AppRegisterRequest, baseURL string) (*model.AppRegisterResponse, error) {
	emailParts := strings.Split(req.Email, "@")
	if len(emailParts) != 2 || strings.TrimSpace(emailParts[0]) == "" || strings.TrimSpace(emailParts[1]) == "" {
		return nil, ErrInvalidRegistrationEmail
	}
	emailDomain := strings.ToLower(strings.TrimSpace(emailParts[1]))

	requestedDomain := strings.ToLower(strings.TrimSpace(req.Domain))
	if requestedDomain == "" {
		requestedDomain = emailDomain
	}

	if freemailDomains[emailDomain] && requestedDomain != emailDomain {
		return nil, fmt.Errorf("%w: %s cannot claim %s", ErrFreemailDomainProhibited, req.Email, requestedDomain)
	}

	var existing model.RegisteredApp
	if err := db.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("%w: %s", ErrAppAlreadyRegistered, req.Email)
	}

	rawAPIKey, err := r.generateAPIKey()
	if err != nil {
		return nil, err
	}
	apiKeyHash := r.HashAPIKey(rawAPIKey)

	appID := uuid.New().String()
	verificationToken := uuid.New().String()
	appURN := fmt.Sprintf("urn:castor:app:%s:%s", requestedDomain, req.AppName)

	var domainStatus model.DomainVerificationStatus
	var dnsChallenge string

	if emailDomain == requestedDomain {
		domainStatus = model.DomainStatusVerifiedSSO
	} else {
		domainStatus = model.DomainStatusPendingDNS
		dnsChallenge = fmt.Sprintf("castor-domain-verify-%s", uuid.New().String())
	}

	now := time.Now().UTC()
	app := model.RegisteredApp{
		AppID:                    appID,
		AppName:                  req.AppName,
		Domain:                   requestedDomain,
		AppURN:                   appURN,
		OrganizationID:           req.OrganizationID,
		Email:                    req.Email,
		DomainVerificationStatus: domainStatus,
		DNSTXTChallenge:          dnsChallenge,
		APIKeyHash:               apiKeyHash,
		IsActive:                 false,
		VerificationToken:        verificationToken,
		CreatedAt:                now,
	}

	ownerMember := model.AppMember{
		ID:         uuid.New().String(),
		AppID:      appID,
		Email:      req.Email,
		Role:       model.RoleOwner,
		InvitedBy:  "system_registration",
		Status:     "ACTIVE",
		CreatedAt:  now,
		AcceptedAt: &now,
	}

	err = db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&app).Error; err != nil {
			return err
		}
		if err := tx.Create(&ownerMember).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	defaultBaseURL := os.Getenv("BASE_URL")
	if defaultBaseURL == "" {
		defaultBaseURL = "http://localhost:8000"
	}
	hostURL := strings.TrimRight(baseURL, "/")
	if hostURL == "" {
		hostURL = strings.TrimRight(defaultBaseURL, "/")
	}

	verificationURL := fmt.Sprintf("%s/api/v1/apps/verify?token=%s", hostURL, verificationToken)

	return &model.AppRegisterResponse{
		AppID:                    app.AppID,
		AppName:                  app.AppName,
		Domain:                   app.Domain,
		AppURN:                   app.AppURN,
		OrganizationID:           app.OrganizationID,
		Email:                    app.Email,
		DomainVerificationStatus: app.DomainVerificationStatus,
		DNSTXTChallenge:          app.DNSTXTChallenge,
		APIKey:                   rawAPIKey,
		VerificationToken:        app.VerificationToken,
		VerificationURL:          verificationURL,
	}, nil
}

// VerifyApp activates a registered application using its verification token.
func (r *AppsRepository) VerifyApp(db *gorm.DB, token string) (*model.AppVerifyResponse, error) {
	var app model.RegisteredApp
	err := db.Where("verification_token = ?", token).First(&app).Error
	if err != nil {
		return nil, ErrInvalidVerificationToken
	}

	now := time.Now().UTC()
	app.IsActive = true
	app.VerifiedAt = &now

	if err := db.Save(&app).Error; err != nil {
		return nil, err
	}

	return &model.AppVerifyResponse{
		AppID:                    app.AppID,
		AppName:                  app.AppName,
		Domain:                   app.Domain,
		AppURN:                   app.AppURN,
		Email:                    app.Email,
		DomainVerificationStatus: app.DomainVerificationStatus,
		IsActive:                 app.IsActive,
		Message:                  "Application email verified successfully. Account is now active.",
	}, nil
}

// AuthenticateAPIKey validates API key and ensures the associated application is active.
func (r *AppsRepository) AuthenticateAPIKey(db *gorm.DB, apiKey string) (*model.RegisteredApp, error) {
	ctx, err := r.AuthenticateContext(db, apiKey)
	if err != nil {
		return nil, err
	}
	return ctx.App, nil
}

// AuthenticateContext validates either root or scoped API key and returns full AuthContext with assigned role.
func (r *AppsRepository) AuthenticateContext(db *gorm.DB, apiKey string) (*model.AuthContext, error) {
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	keyHash := r.HashAPIKey(apiKey)

	// 1. Check Scoped API Keys (User/Team/Service Keys)
	var scopedKey model.AppAPIKey
	if err := db.Where("api_key_hash = ? AND revoked_at IS NULL", keyHash).First(&scopedKey).Error; err == nil {
		if scopedKey.ExpiresAt != nil && scopedKey.ExpiresAt.Before(time.Now().UTC()) {
			return nil, ErrInvalidAPIKey
		}

		var app model.RegisteredApp
		if err := db.Where("app_id = ?", scopedKey.AppID).First(&app).Error; err != nil {
			return nil, ErrInvalidAPIKey
		}
		if !app.IsActive {
			return nil, ErrAppNotVerified
		}

		now := time.Now().UTC()
		_ = db.Model(&scopedKey).Update("last_used_at", now).Error

		return &model.AuthContext{
			App:         &app,
			MemberEmail: scopedKey.MemberEmail,
			Role:        scopedKey.Role,
			KeyID:       scopedKey.ID,
		}, nil
	}

	// 2. Check Root Application API Key (Full Owner Access)
	var app model.RegisteredApp
	if err := db.Where("api_key_hash = ?", keyHash).First(&app).Error; err == nil {
		if !app.IsActive {
			return nil, ErrAppNotVerified
		}

		return &model.AuthContext{
			App:         &app,
			MemberEmail: app.Email,
			Role:        model.RoleOwner,
			KeyID:       "root",
		}, nil
	}

	return nil, ErrInvalidAPIKey
}

// InviteMember creates a new pending collaborator invitation for an application.
func (r *AppsRepository) InviteMember(
	db *gorm.DB,
	appID, inviterEmail, targetEmail string,
	role model.AppRole,
	baseURL string,
) (*model.MemberInviteResponse, error) {
	targetEmail = strings.ToLower(strings.TrimSpace(targetEmail))
	if !strings.Contains(targetEmail, "@") {
		return nil, ErrInvalidRegistrationEmail
	}

	var existing model.AppMember
	if err := db.Where("app_id = ? AND email = ?", appID, targetEmail).First(&existing).Error; err == nil {
		return nil, fmt.Errorf("%w: %s", ErrMemberAlreadyExists, targetEmail)
	}

	invitationToken := uuid.New().String()
	now := time.Now().UTC()

	member := model.AppMember{
		ID:              uuid.New().String(),
		AppID:           appID,
		Email:           targetEmail,
		Role:            role,
		InvitedBy:       inviterEmail,
		Status:          "PENDING_INVITE",
		InvitationToken: invitationToken,
		CreatedAt:       now,
	}

	if err := db.Create(&member).Error; err != nil {
		return nil, err
	}

	defaultBaseURL := os.Getenv("BASE_URL")
	if defaultBaseURL == "" {
		defaultBaseURL = "http://localhost:8000"
	}
	hostURL := strings.TrimRight(baseURL, "/")
	if hostURL == "" {
		hostURL = strings.TrimRight(defaultBaseURL, "/")
	}
	inviteURL := fmt.Sprintf("%s/api/v1/apps/members/accept?token=%s", hostURL, invitationToken)

	return &model.MemberInviteResponse{
		ID:              member.ID,
		AppID:           member.AppID,
		Email:           member.Email,
		Role:            member.Role,
		Status:          member.Status,
		InvitedBy:       member.InvitedBy,
		InvitationToken: invitationToken,
		InvitationURL:   inviteURL,
		CreatedAt:       member.CreatedAt,
	}, nil
}

// AcceptInvitation activates an invited collaborator via token.
func (r *AppsRepository) AcceptInvitation(db *gorm.DB, token string) (*model.AppMember, error) {
	var member model.AppMember
	if err := db.Where("invitation_token = ? AND status = 'PENDING_INVITE'", token).First(&member).Error; err != nil {
		return nil, ErrInvalidInvitationToken
	}

	now := time.Now().UTC()
	member.Status = "ACTIVE"
	member.AcceptedAt = &now

	if err := db.Save(&member).Error; err != nil {
		return nil, err
	}

	return &member, nil
}

// ListMembers lists all registered collaborators for an application.
func (r *AppsRepository) ListMembers(db *gorm.DB, appID string) ([]model.AppMember, error) {
	var members []model.AppMember
	if err := db.Where("app_id = ?", appID).Order("created_at asc").Find(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

// UpdateMemberRole updates the RBAC role for an existing collaborator.
func (r *AppsRepository) UpdateMemberRole(db *gorm.DB, appID, memberID string, newRole model.AppRole) (*model.AppMember, error) {
	var member model.AppMember
	if err := db.Where("id = ? AND app_id = ?", memberID, appID).First(&member).Error; err != nil {
		return nil, ErrMemberNotFound
	}

	member.Role = newRole
	if err := db.Save(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

// RemoveMember revokes and deletes a collaborator from an application.
func (r *AppsRepository) RemoveMember(db *gorm.DB, appID, memberID string) error {
	res := db.Where("id = ? AND app_id = ?", memberID, appID).Delete(&model.AppMember{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrMemberNotFound
	}
	return nil
}

// CreateScopedAPIKey provisions a new scoped API key for a team member or automated service.
func (r *AppsRepository) CreateScopedAPIKey(
	db *gorm.DB,
	appID, memberEmail, keyName string,
	role model.AppRole,
	expiresInDays int,
) (*model.CreateAPIKeyResponse, error) {
	rawKey, err := r.generateAPIKey()
	if err != nil {
		return nil, err
	}
	keyHash := r.HashAPIKey(rawKey)
	now := time.Now().UTC()

	var expiresAt *time.Time
	if expiresInDays > 0 {
		exp := now.AddDate(0, 0, expiresInDays)
		expiresAt = &exp
	}

	if role == "" {
		role = model.RoleEditor
	}

	apiKey := model.AppAPIKey{
		ID:          uuid.New().String(),
		AppID:       appID,
		MemberEmail: memberEmail,
		Name:        keyName,
		APIKeyHash:  keyHash,
		Role:        role,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
	}

	if err := db.Create(&apiKey).Error; err != nil {
		return nil, err
	}

	return &model.CreateAPIKeyResponse{
		ID:          apiKey.ID,
		AppID:       apiKey.AppID,
		MemberEmail: apiKey.MemberEmail,
		Name:        apiKey.Name,
		APIKey:      rawKey,
		Role:        apiKey.Role,
		CreatedAt:   apiKey.CreatedAt,
		ExpiresAt:   apiKey.ExpiresAt,
	}, nil
}

// ListAPIKeys lists all active and revoked scoped API keys for an application.
func (r *AppsRepository) ListAPIKeys(db *gorm.DB, appID string) ([]model.AppAPIKeySummary, error) {
	var keys []model.AppAPIKey
	if err := db.Where("app_id = ?", appID).Order("created_at desc").Find(&keys).Error; err != nil {
		return nil, err
	}

	summaries := make([]model.AppAPIKeySummary, 0, len(keys))
	now := time.Now().UTC()
	for _, k := range keys {
		isActive := k.RevokedAt == nil
		if k.ExpiresAt != nil && k.ExpiresAt.Before(now) {
			isActive = false
		}
		summaries = append(summaries, model.AppAPIKeySummary{
			ID:          k.ID,
			AppID:       k.AppID,
			MemberEmail: k.MemberEmail,
			Name:        k.Name,
			Role:        k.Role,
			CreatedAt:   k.CreatedAt,
			LastUsedAt:  k.LastUsedAt,
			ExpiresAt:   k.ExpiresAt,
			IsActive:    isActive,
		})
	}
	return summaries, nil
}

// RevokeAPIKey immediately revokes an API key.
func (r *AppsRepository) RevokeAPIKey(db *gorm.DB, appID, keyID string) error {
	now := time.Now().UTC()
	res := db.Model(&model.AppAPIKey{}).
		Where("id = ? AND app_id = ? AND revoked_at IS NULL", keyID, appID).
		Update("revoked_at", now)

	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrKeyNotFound
	}
	return nil
}
