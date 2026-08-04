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

// RegisteredApp represents a registered application authorized to manage skills.
type RegisteredApp struct {
	AppID                  string                   `gorm:"primaryKey;column:app_id" json:"app_id"`
	AppName                string                   `gorm:"index;column:app_name" json:"app_name"`
	Domain                 string                   `gorm:"index;column:domain" json:"domain"`
	AppURN                 string                   `gorm:"uniqueIndex;column:app_urn" json:"app_urn"`
	OrganizationID         string                   `gorm:"index;column:organization_id" json:"organization_id,omitempty"`
	Email                  string                   `gorm:"index;column:email" json:"email"`
	DomainVerificationStatus DomainVerificationStatus `gorm:"column:domain_verification_status" json:"domain_verification_status"`
	DNSTXTChallenge        string                   `gorm:"column:dns_txt_challenge" json:"dns_txt_challenge,omitempty"`
	APIKeyHash             string                   `gorm:"index;column:api_key_hash" json:"-"`
	IsActive               bool                     `gorm:"index;column:is_active" json:"is_active"`
	VerificationToken      string                   `gorm:"index;column:verification_token" json:"verification_token"`
	CreatedAt              time.Time                `gorm:"column:created_at" json:"created_at"`
	VerifiedAt             *time.Time               `gorm:"column:verified_at" json:"verified_at,omitempty"`
}

// TableName specifies table name for GORM.
func (RegisteredApp) TableName() string {
	return "registered_apps"
}

// AppRegisterRequest is the payload for registering a new application.
type AppRegisterRequest struct {
	AppName        string `json:"app_name" binding:"required"`
	Domain         string `json:"domain"`
	OrganizationID string `json:"organization_id"`
	Email          string `json:"email" binding:"required"`
}

// AppRegisterResponse is the response returned upon application registration.
type AppRegisterResponse struct {
	AppID                  string                   `json:"app_id"`
	AppName                string                   `json:"app_name"`
	Domain                 string                   `json:"domain"`
	AppURN                 string                   `json:"app_urn"`
	OrganizationID         string                   `json:"organization_id,omitempty"`
	Email                  string                   `json:"email"`
	DomainVerificationStatus DomainVerificationStatus `json:"domain_verification_status"`
	DNSTXTChallenge        string                   `json:"dns_txt_challenge,omitempty"`
	APIKey                 string                   `json:"api_key"`
	VerificationToken      string                   `json:"verification_token"`
	VerificationURL        string                   `json:"verification_url"`
}

// AppVerifyResponse is the response returned upon email verification.
type AppVerifyResponse struct {
	AppID                  string                   `json:"app_id"`
	AppName                string                   `json:"app_name"`
	Domain                 string                   `json:"domain"`
	AppURN                 string                   `json:"app_urn"`
	Email                  string                   `json:"email"`
	DomainVerificationStatus DomainVerificationStatus `json:"domain_verification_status"`
	IsActive               bool                     `json:"is_active"`
	Message                string                   `json:"message"`
}
