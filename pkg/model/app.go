package model

import (
	"time"
)

// RegisteredApp represents a registered application authorized to manage skills.
type RegisteredApp struct {
	AppID             string     `gorm:"primaryKey;column:app_id" json:"app_id"`
	AppName           string     `gorm:"index;column:app_name" json:"app_name"`
	Email             string     `gorm:"index;column:email" json:"email"`
	APIKeyHash        string     `gorm:"index;column:api_key_hash" json:"-"`
	IsActive          bool       `gorm:"index;column:is_active" json:"is_active"`
	VerificationToken string     `gorm:"index;column:verification_token" json:"verification_token"`
	CreatedAt         time.Time  `gorm:"column:created_at" json:"created_at"`
	VerifiedAt        *time.Time `gorm:"column:verified_at" json:"verified_at,omitempty"`
}

// TableName specifies table name for GORM.
func (RegisteredApp) TableName() string {
	return "registered_apps"
}

// AppRegisterRequest is the payload for registering a new application.
type AppRegisterRequest struct {
	AppName string `json:"app_name" binding:"required"`
	Email   string `json:"email" binding:"required"`
}

// AppRegisterResponse is the response returned upon application registration.
type AppRegisterResponse struct {
	AppID             string `json:"app_id"`
	AppName           string `json:"app_name"`
	Email             string `json:"email"`
	APIKey            string `json:"api_key"`
	VerificationToken string `json:"verification_token"`
	VerificationURL   string `json:"verification_url"`
}

// AppVerifyResponse is the response returned upon email verification.
type AppVerifyResponse struct {
	AppID    string `json:"app_id"`
	AppName  string `json:"app_name"`
	Email    string `json:"email"`
	IsActive bool   `json:"is_active"`
	Message  string `json:"message"`
}
