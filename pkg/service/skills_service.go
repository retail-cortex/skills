package service

import (
	"fmt"
	"strings"

	"github.com/retail-cortex/skills/pkg/data"
	"github.com/retail-cortex/skills/pkg/model"
	"gorm.io/gorm"
)

type SkillsService struct {
	repo         *data.SkillsRepository
	embedService *EmbeddingService
}

func NewSkillsService(repo ...*data.SkillsRepository) *SkillsService {
	var r *data.SkillsRepository
	if len(repo) > 0 && repo[0] != nil {
		r = repo[0]
	} else {
		r = data.NewSkillsRepository()
	}
	return &SkillsService{
		repo:         r,
		embedService: NewEmbeddingService(),
	}
}

func (s *SkillsService) ListSkills(db *gorm.DB, query string) ([]*model.SkillResponse, error) {
	var queryVector []float64
	if strings.TrimSpace(query) != "" {
		queryVector = s.embedService.GenerateEmbedding(query)
	}
	return s.repo.ListSkills(db, query, queryVector)
}

func (s *SkillsService) GetSkill(db *gorm.DB, skillIDOrName string) (*model.SkillResponse, error) {
	return s.repo.GetSkill(db, skillIDOrName)
}

func (s *SkillsService) CreateSkill(db *gorm.DB, appID string, req model.SkillCreateRequest) (*model.SkillResponse, error) {
	trigStr := strings.Join(req.TriggerPhrases, " ")
	textForEmbedding := fmt.Sprintf("%s %s %s %s", req.Name, req.Description, req.Instructions, trigStr)
	vector := s.embedService.GenerateEmbedding(textForEmbedding)

	return s.repo.CreateSkill(db, appID, req, vector, s.embedService.ModelName)
}

func (s *SkillsService) UpdateSkill(
	db *gorm.DB,
	skillID string,
	appID string,
	req model.SkillUpdateRequest,
	fullReplace bool,
) (*model.SkillResponse, error) {
	existing, err := s.repo.GetSkill(db, skillID)
	if err != nil {
		return nil, err
	}

	name := existing.Name
	desc := existing.Description
	if req.Description != nil {
		desc = *req.Description
	}
	inst := existing.Instructions
	if req.Instructions != nil {
		inst = *req.Instructions
	}
	trig := existing.TriggerPhrases
	if req.TriggerPhrases != nil {
		trig = *req.TriggerPhrases
	}
	trigStr := strings.Join(trig, " ")
	textForEmbedding := fmt.Sprintf("%s %s %s %s", name, desc, inst, trigStr)
	vector := s.embedService.GenerateEmbedding(textForEmbedding)

	return s.repo.UpdateSkill(db, skillID, appID, req, fullReplace, vector, s.embedService.ModelName)
}

func (s *SkillsService) DeleteSkill(db *gorm.DB, skillID string, appID string) (map[string]interface{}, error) {
	return s.repo.DeleteSkill(db, skillID, appID)
}
