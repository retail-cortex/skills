package service

import (
	"github.com/retail-cortex/skills/pkg/data"
	"github.com/retail-cortex/skills/pkg/model"
	"gorm.io/gorm"
)

type AppsService struct {
	repo *data.AppsRepository
}

func NewAppsService(repo ...*data.AppsRepository) *AppsService {
	var r *data.AppsRepository
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
