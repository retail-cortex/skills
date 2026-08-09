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
	"github.com/retail-cortex/skills/pkg/model"
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

type AppsRepository struct{}

func NewAppsRepository() *AppsRepository {
	return &AppsRepository{}
}

// HashAPIKey computes SHA-256 hash of API key for secure DB storage.
func (r *AppsRepository) HashAPIKey(apiKey string) string {
	hash := sha256.Sum256([]byte(apiKey))
	return hex.EncodeToString(hash[:])
}

// generateAPIKey creates a secure random API key string skm_live_...
func (r *AppsRepository) generateAPIKey() (string, error) {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("skm_live_%s", hex.EncodeToString(bytes)), nil
}

// RegisterApp registers a new application with domain scoping, issuing an API key and email verification token.
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
	err := db.Where("email = ?", req.Email).First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("%w: %s", ErrAppAlreadyRegistered, req.Email)
	}

	rawAPIKey, err := r.generateAPIKey()
	if err != nil {
		return nil, err
	}
	apiKeyHash := r.HashAPIKey(rawAPIKey)

	appID := uuid.New().String()
	verificationToken := uuid.New().String()
	appURN := fmt.Sprintf("urn:skm:app:%s:%s", requestedDomain, req.AppName)

	var domainStatus model.DomainVerificationStatus
	var dnsChallenge string

	if emailDomain == requestedDomain {
		domainStatus = model.DomainStatusVerifiedSSO
	} else {
		domainStatus = model.DomainStatusPendingDNS
		dnsChallenge = fmt.Sprintf("skm-domain-verify-%s", uuid.New().String())
	}

	app := model.RegisteredApp{
		AppID:                  appID,
		AppName:                req.AppName,
		Domain:                 requestedDomain,
		AppURN:                 appURN,
		OrganizationID:         req.OrganizationID,
		Email:                  req.Email,
		DomainVerificationStatus: domainStatus,
		DNSTXTChallenge:        dnsChallenge,
		APIKeyHash:             apiKeyHash,
		IsActive:               false,
		VerificationToken:      verificationToken,
		CreatedAt:              time.Now().UTC(),
	}

	if err := db.Create(&app).Error; err != nil {
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
		AppID:                  app.AppID,
		AppName:                app.AppName,
		Domain:                 app.Domain,
		AppURN:                 app.AppURN,
		OrganizationID:         app.OrganizationID,
		Email:                  app.Email,
		DomainVerificationStatus: app.DomainVerificationStatus,
		DNSTXTChallenge:        app.DNSTXTChallenge,
		APIKey:                 rawAPIKey,
		VerificationToken:      app.VerificationToken,
		VerificationURL:        verificationURL,
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
		AppID:                  app.AppID,
		AppName:                app.AppName,
		Domain:                 app.Domain,
		AppURN:                 app.AppURN,
		Email:                  app.Email,
		DomainVerificationStatus: app.DomainVerificationStatus,
		IsActive:               app.IsActive,
		Message:                "Application email verified successfully. Account is now active.",
	}, nil
}

// AuthenticateAPIKey validates API key and ensures the associated application is active.
func (r *AppsRepository) AuthenticateAPIKey(db *gorm.DB, apiKey string) (*model.RegisteredApp, error) {
	if apiKey == "" {
		return nil, ErrMissingAPIKey
	}

	keyHash := r.HashAPIKey(apiKey)
	var app model.RegisteredApp
	err := db.Where("api_key_hash = ?", keyHash).First(&app).Error
	if err != nil {
		return nil, ErrInvalidAPIKey
	}

	if !app.IsActive {
		return nil, ErrAppNotVerified
	}

	return &app, nil
}
