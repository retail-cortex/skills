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
	"context"
	"log"
	"strings"
	"sync"

	"github.com/retail-cortex/castor/pkg/data"
	"github.com/retail-cortex/castor/pkg/embedding"
	"github.com/retail-cortex/castor/pkg/embedding/vertex"
	"github.com/retail-cortex/castor/pkg/model"
	"gorm.io/gorm"
)

type embeddingJob struct {
	DB      *gorm.DB
	SkillID string
	Name    string
	Desc    string
	Inst    string
	Trig    []string
	Refs    map[string]string
	Exs     map[string]string
}

type CastorService struct {
	repo         data.Repository
	provider     embedding.Provider
	embedJobChan chan embeddingJob
	workerOnce   sync.Once
}

func NewCastorService(repo ...data.Repository) *CastorService {
	return NewCastorServiceWithProvider(vertex.NewProvider(), repo...)
}

func NewCastorServiceWithProvider(provider embedding.Provider, repo ...data.Repository) *CastorService {
	var r data.Repository
	if len(repo) > 0 && repo[0] != nil {
		r = repo[0]
	} else {
		r = data.NewCastorRepository()
	}
	if provider == nil {
		provider = vertex.NewProvider()
	}
	svc := &CastorService{
		repo:         r,
		provider:     provider,
		embedJobChan: make(chan embeddingJob, 100),
	}
	svc.startBackgroundWorkers(4)
	return svc
}

func NewCastorServiceWithConfig(cfg EmbeddingConfig, repo ...data.Repository) *CastorService {
	vCfg := vertex.Config{
		ModelName:    cfg.ModelName,
		ProjectID:    cfg.ProjectID,
		Region:       cfg.Region,
		GeminiAPIKey: cfg.GeminiAPIKey,
		BaseURL:      cfg.BaseURL,
	}
	return NewCastorServiceWithProvider(vertex.NewProvider(vCfg), repo...)
}

// Deprecated: SkillsService is an alias for CastorService.
type SkillsService = CastorService

// Deprecated: NewSkillsService is an alias for NewCastorService.
func NewSkillsService(repo ...data.Repository) *CastorService {
	return NewCastorService(repo...)
}

// Deprecated: NewSkillsServiceWithProvider is an alias for NewCastorServiceWithProvider.
func NewSkillsServiceWithProvider(provider embedding.Provider, repo ...data.Repository) *CastorService {
	return NewCastorServiceWithProvider(provider, repo...)
}

// Deprecated: NewSkillsServiceWithConfig is an alias for NewCastorServiceWithConfig.
func NewSkillsServiceWithConfig(cfg EmbeddingConfig, repo ...data.Repository) *CastorService {
	return NewCastorServiceWithConfig(cfg, repo...)
}

func (s *CastorService) startBackgroundWorkers(concurrency int) {
	s.workerOnce.Do(func() {
		for i := 0; i < concurrency; i++ {
			go func() {
				for job := range s.embedJobChan {
					chunks, err := s.provider.GenerateSkillEmbeddings(
						context.Background(),
						job.Name,
						job.Desc,
						job.Inst,
						job.Trig,
						job.Refs,
						job.Exs,
					)
					if err == nil && len(chunks) > 0 {
						if err := s.repo.SaveSkillEmbeddings(job.DB, job.SkillID, chunks); err != nil {
							log.Printf("Background embedding persistence error for skill %s: %v", job.SkillID, err)
						}
					}
				}
			}()
		}
	})
}

func (s *CastorService) ListSkills(
	db *gorm.DB,
	query string,
	pagination ...model.PaginationParams,
) (*model.PaginatedSkillResponse, error) {
	var queryVector []float64
	if strings.TrimSpace(query) != "" {
		vec, err := s.provider.GenerateEmbedding(context.Background(), query)
		if err == nil {
			queryVector = vec
		}
	}
	return s.repo.ListSkills(db, query, queryVector, pagination...)
}

func (s *CastorService) GetSkill(db *gorm.DB, skillIDOrName string) (*model.SkillResponse, error) {
	return s.repo.GetSkill(db, skillIDOrName)
}

func (s *CastorService) CreateSkill(db *gorm.DB, appID string, req model.SkillCreateRequest) (*model.SkillResponse, error) {
	resp, err := s.repo.CreateSkill(db, appID, req, nil)
	if err != nil {
		return nil, err
	}

	select {
	case s.embedJobChan <- embeddingJob{
		DB:      db,
		SkillID: resp.ID,
		Name:    req.Name,
		Desc:    req.Description,
		Inst:    req.Instructions,
		Trig:    req.TriggerPhrases,
		Refs:    req.References,
		Exs:     req.Examples,
	}:
	default:
		go func() {
			chunks, _ := s.provider.GenerateSkillEmbeddings(
				context.Background(),
				req.Name,
				req.Description,
				req.Instructions,
				req.TriggerPhrases,
				req.References,
				req.Examples,
			)
			_ = s.repo.SaveSkillEmbeddings(db, resp.ID, chunks)
		}()
	}

	return resp, nil
}

func (s *CastorService) CreateSkillSync(db *gorm.DB, appID string, req model.SkillCreateRequest) (*model.SkillResponse, error) {
	chunks, _ := s.provider.GenerateSkillEmbeddings(
		context.Background(),
		req.Name,
		req.Description,
		req.Instructions,
		req.TriggerPhrases,
		req.References,
		req.Examples,
	)
	return s.repo.CreateSkill(db, appID, req, chunks)
}

func (s *CastorService) UpdateSkill(
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

	resp, err := s.repo.UpdateSkill(db, skillID, appID, req, fullReplace, nil)
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
	refs := existing.References
	if req.References != nil {
		refs = *req.References
	}
	exs := existing.Examples
	if req.Examples != nil {
		exs = *req.Examples
	}

	select {
	case s.embedJobChan <- embeddingJob{
		DB:      db,
		SkillID: resp.ID,
		Name:    name,
		Desc:    desc,
		Inst:    inst,
		Trig:    trig,
		Refs:    refs,
		Exs:     exs,
	}:
	default:
		go func() {
			chunks, _ := s.provider.GenerateSkillEmbeddings(context.Background(), name, desc, inst, trig, refs, exs)
			_ = s.repo.SaveSkillEmbeddings(db, resp.ID, chunks)
		}()
	}

	return resp, nil
}

func (s *SkillsService) DeleteSkill(db *gorm.DB, skillID string, appID string) (map[string]interface{}, error) {
	return s.repo.DeleteSkill(db, skillID, appID)
}
