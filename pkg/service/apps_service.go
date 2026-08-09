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

package service

import (
	"github.com/retail-cortex/skills/pkg/data"
	"github.com/retail-cortex/skills/pkg/model"
	"gorm.io/gorm"
)

type AppsService struct {
	repo data.AppRepository
}

func NewAppsService(repo ...data.AppRepository) *AppsService {
	var r data.AppRepository
	if len(repo) > 0 && repo[0] != nil {
		r = repo[0]
	} else {
		r = data.NewAppsRepository()
	}
	return &AppsService{repo: r}
}

func (s *AppsService) RegisterApp(db *gorm.DB, req model.AppRegisterRequest, baseURL string) (*model.AppRegisterResponse, error) {
	return s.repo.RegisterApp(db, req, baseURL)
}

func (s *AppsService) VerifyApp(db *gorm.DB, token string) (*model.AppVerifyResponse, error) {
	return s.repo.VerifyApp(db, token)
}

func (s *AppsService) AuthenticateAPIKey(db *gorm.DB, apiKey string) (*model.RegisteredApp, error) {
	return s.repo.AuthenticateAPIKey(db, apiKey)
}

func (s *AppsService) AuthenticateContext(db *gorm.DB, apiKey string) (*model.AuthContext, error) {
	return s.repo.AuthenticateContext(db, apiKey)
}

func (s *AppsService) InviteMember(db *gorm.DB, appID, inviterEmail, targetEmail string, role model.AppRole, baseURL string) (*model.MemberInviteResponse, error) {
	return s.repo.InviteMember(db, appID, inviterEmail, targetEmail, role, baseURL)
}

func (s *AppsService) AcceptInvitation(db *gorm.DB, token string) (*model.AppMember, error) {
	return s.repo.AcceptInvitation(db, token)
}

func (s *AppsService) ListMembers(db *gorm.DB, appID string) ([]model.AppMember, error) {
	return s.repo.ListMembers(db, appID)
}

func (s *AppsService) UpdateMemberRole(db *gorm.DB, appID, memberID string, newRole model.AppRole) (*model.AppMember, error) {
	return s.repo.UpdateMemberRole(db, appID, memberID, newRole)
}

func (s *AppsService) RemoveMember(db *gorm.DB, appID, memberID string) error {
	return s.repo.RemoveMember(db, appID, memberID)
}

func (s *AppsService) CreateScopedAPIKey(db *gorm.DB, appID, memberEmail, keyName string, role model.AppRole, expiresInDays int) (*model.CreateAPIKeyResponse, error) {
	return s.repo.CreateScopedAPIKey(db, appID, memberEmail, keyName, role, expiresInDays)
}

func (s *AppsService) ListAPIKeys(db *gorm.DB, appID string) ([]model.AppAPIKeySummary, error) {
	return s.repo.ListAPIKeys(db, appID)
}

func (s *AppsService) RevokeAPIKey(db *gorm.DB, appID, keyID string) error {
	return s.repo.RevokeAPIKey(db, appID, keyID)
}
