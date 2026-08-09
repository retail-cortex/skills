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

package model

import (
	"time"
)

// DomainVerificationStatus classifies domain ownership validation state.
type DomainVerificationStatus string

const (
	DomainStatusUnspecified DomainVerificationStatus = "UNSPECIFIED"
	DomainStatusVerifiedSSO DomainVerificationStatus = "VERIFIED_SSO"
	DomainStatusVerifiedDNS DomainVerificationStatus = "VERIFIED_DNS"
	DomainStatusPendingDNS  DomainVerificationStatus = "PENDING_DNS"
	DomainStatusRejected    DomainVerificationStatus = "REJECTED"
)

// AppRole defines Role-Based Access Control (RBAC) levels for applications and skills.
type AppRole string

const (
	RoleOwner  AppRole = "OWNER"
	RoleEditor AppRole = "EDITOR"
	RoleViewer AppRole = "VIEWER"
)

// HasPermission checks if the current role meets or exceeds the required role.
func (r AppRole) HasPermission(required AppRole) bool {
	switch required {
	case RoleViewer:
		return r == RoleViewer || r == RoleEditor || r == RoleOwner
	case RoleEditor:
		return r == RoleEditor || r == RoleOwner
	case RoleOwner:
		return r == RoleOwner
	default:
		return false
	}
}

// RegisteredApp represents a registered application authorized to manage skills.
type RegisteredApp struct {
	AppID                    string                   `gorm:"primaryKey;column:app_id" json:"app_id"`
	AppName                  string                   `gorm:"index;column:app_name" json:"app_name"`
	Domain                   string                   `gorm:"index;column:domain" json:"domain"`
	AppURN                   string                   `gorm:"uniqueIndex;column:app_urn" json:"app_urn"`
	OrganizationID           string                   `gorm:"index;column:organization_id" json:"organization_id,omitempty"`
	Email                    string                   `gorm:"index;column:email" json:"email"`
	DomainVerificationStatus DomainVerificationStatus `gorm:"column:domain_verification_status" json:"domain_verification_status"`
	DNSTXTChallenge          string                   `gorm:"column:dns_txt_challenge" json:"dns_txt_challenge,omitempty"`
	APIKeyHash               string                   `gorm:"index;column:api_key_hash" json:"-"`
	IsActive                 bool                     `gorm:"index;column:is_active" json:"is_active"`
	VerificationToken        string                   `gorm:"index;column:verification_token" json:"verification_token"`
	CreatedAt                time.Time                `gorm:"column:created_at" json:"created_at"`
	VerifiedAt               *time.Time               `gorm:"column:verified_at" json:"verified_at,omitempty"`
}

func (RegisteredApp) TableName() string {
	return "registered_apps"
}

// AppMember represents an individual collaborator/user assigned a role within an application.
type AppMember struct {
	ID              string     `gorm:"primaryKey;column:id" json:"id"`
	AppID           string     `gorm:"uniqueIndex:idx_app_member_email;column:app_id" json:"app_id"`
	Email           string     `gorm:"uniqueIndex:idx_app_member_email;column:email" json:"email"`
	Role            AppRole    `gorm:"column:role;default:'EDITOR'" json:"role"`
	InvitedBy       string     `gorm:"column:invited_by" json:"invited_by"`
	Status          string     `gorm:"column:status;default:'ACTIVE'" json:"status"` // "ACTIVE", "PENDING_INVITE", "REVOKED"
	InvitationToken string     `gorm:"index;column:invitation_token" json:"-"`
	CreatedAt       time.Time  `gorm:"column:created_at" json:"created_at"`
	AcceptedAt      *time.Time `gorm:"column:accepted_at" json:"accepted_at,omitempty"`
}

func (AppMember) TableName() string {
	return "app_members"
}

// AppAPIKey represents a scoped API key provisioned for a user, team, or CI/CD pipeline.
type AppAPIKey struct {
	ID          string     `gorm:"primaryKey;column:id" json:"id"`
	AppID       string     `gorm:"index;column:app_id" json:"app_id"`
	MemberEmail string     `gorm:"index;column:member_email" json:"member_email"`
	Name        string     `gorm:"column:name" json:"name"`
	APIKeyHash  string     `gorm:"uniqueIndex;column:api_key_hash" json:"-"`
	Role        AppRole    `gorm:"column:role;default:'EDITOR'" json:"role"`
	ExpiresAt   *time.Time `gorm:"column:expires_at" json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `gorm:"column:last_used_at" json:"last_used_at,omitempty"`
	CreatedAt   time.Time  `gorm:"column:created_at" json:"created_at"`
	RevokedAt   *time.Time `gorm:"column:revoked_at" json:"revoked_at,omitempty"`
}

func (AppAPIKey) TableName() string {
	return "app_api_keys"
}

// AuthContext captures the validated identity and permission tier for an authenticated request.
type AuthContext struct {
	App         *RegisteredApp `json:"app"`
	MemberEmail string         `json:"member_email"`
	Role        AppRole        `json:"role"`
	KeyID       string         `json:"key_id,omitempty"`
}

// DTOs for Application & Collaborator RBAC

type AppRegisterRequest struct {
	AppName        string `json:"app_name" binding:"required"`
	Domain         string `json:"domain"`
	OrganizationID string `json:"organization_id"`
	Email          string `json:"email" binding:"required"`
}

type AppRegisterResponse struct {
	AppID                    string                   `json:"app_id"`
	AppName                  string                   `json:"app_name"`
	Domain                   string                   `json:"domain"`
	AppURN                   string                   `json:"app_urn"`
	OrganizationID           string                   `json:"organization_id,omitempty"`
	Email                    string                   `json:"email"`
	DomainVerificationStatus DomainVerificationStatus `json:"domain_verification_status"`
	DNSTXTChallenge          string                   `json:"dns_txt_challenge,omitempty"`
	APIKey                   string                   `json:"api_key"`
	VerificationToken        string                   `json:"verification_token"`
	VerificationURL          string                   `json:"verification_url"`
}

type AppVerifyResponse struct {
	AppID                    string                   `json:"app_id"`
	AppName                  string                   `json:"app_name"`
	Domain                   string                   `json:"domain"`
	AppURN                   string                   `json:"app_urn"`
	Email                    string                   `json:"email"`
	DomainVerificationStatus DomainVerificationStatus `json:"domain_verification_status"`
	IsActive                 bool                     `json:"is_active"`
	Message                  string                   `json:"message"`
}

// Member Management DTOs

type MemberInviteRequest struct {
	Email string  `json:"email" binding:"required"`
	Role  AppRole `json:"role" binding:"required"`
}

type MemberInviteResponse struct {
	ID              string     `json:"id"`
	AppID           string     `json:"app_id"`
	Email           string     `json:"email"`
	Role            AppRole    `json:"role"`
	Status          string     `json:"status"`
	InvitedBy       string     `json:"invited_by"`
	InvitationToken string     `json:"invitation_token,omitempty"`
	InvitationURL   string     `json:"invitation_url,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	AcceptedAt      *time.Time `json:"accepted_at,omitempty"`
}

type MemberAcceptRequest struct {
	Token string `json:"token" binding:"required"`
}

type MemberUpdateRoleRequest struct {
	Role AppRole `json:"role" binding:"required"`
}

type CreateAPIKeyRequest struct {
	Name          string  `json:"name" binding:"required"`
	Role          AppRole `json:"role"`
	ExpiresInDays int     `json:"expires_in_days,omitempty"`
}

type CreateAPIKeyResponse struct {
	ID          string     `json:"id"`
	AppID       string     `json:"app_id"`
	MemberEmail string     `json:"member_email"`
	Name        string     `json:"name"`
	APIKey      string     `json:"api_key"`
	Role        AppRole    `json:"role"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type AppAPIKeySummary struct {
	ID          string     `json:"id"`
	AppID       string     `json:"app_id"`
	MemberEmail string     `json:"member_email"`
	Name        string     `json:"name"`
	Role        AppRole    `json:"role"`
	CreatedAt   time.Time  `json:"created_at"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	IsActive    bool       `json:"is_active"`
}
