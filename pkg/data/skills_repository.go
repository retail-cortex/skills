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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/retail-cortex/skills/pkg/embedding"
	"github.com/retail-cortex/skills/pkg/model"
	"gorm.io/gorm"
)

var (
	ErrSkillNotFound           = errors.New("skill not found")
	ErrUnauthorizedSkillAccess = errors.New("unauthorized: skill belongs to a different application")
)

// Repository defines the contract for skill storage, indexing, and retrieval.
type Repository interface {
	BuildSkillResponse(db *gorm.DB, skill *model.Skill, similarityScore *float64) (*model.SkillResponse, error)
	CreateSkill(db *gorm.DB, appID string, req model.SkillCreateRequest, embeddingChunks []model.SkillEmbeddingChunk) (*model.SkillResponse, error)
	GetSkill(db *gorm.DB, skillIDOrNameOrURI string) (*model.SkillResponse, error)
	ListSkills(db *gorm.DB, query string, queryVector []float64, pagination ...model.PaginationParams) (*model.PaginatedSkillResponse, error)
	UpdateSkill(db *gorm.DB, skillID string, appID string, req model.SkillUpdateRequest, fullReplace bool, embeddingChunks []model.SkillEmbeddingChunk) (*model.SkillResponse, error)
	SaveSkillEmbeddings(db *gorm.DB, skillID string, embeddingChunks []model.SkillEmbeddingChunk) error
	DeleteSkill(db *gorm.DB, skillID string, appID string) (map[string]interface{}, error)
}

type SkillsRepository struct{}

func NewSkillsRepository() *SkillsRepository {
	return &SkillsRepository{}
}

// ComputeSkillHash computes a SHA256 hex string over skill components.
func ComputeSkillHash(name, description, instructions string) string {
	content := fmt.Sprintf("%s\n%s\n%s", name, description, instructions)
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// CosineSimilarity computes cosine similarity between two float64 vectors.
func CosineSimilarity(vec1, vec2 []float64) float64 {
	return embedding.CosineSimilarity(vec1, vec2)
}

// buildSkillEmbeddingRecord creates a SkillEmbedding model populated with poly-column vectors.
func buildSkillEmbeddingRecord(skillID, targetType, targetName, modelName string, vector []float64, now time.Time) model.SkillEmbedding {
	embBytes, _ := json.Marshal(vector)
	vecStr := string(embBytes)
	dim := len(vector)
	if modelName == "" {
		modelName = "multimodalembedding"
	}
	if targetType == "" {
		targetType = "skill"
	}
	if targetName == "" {
		targetName = "SKILL.md"
	}

	record := model.SkillEmbedding{
		ID:            uuid.New().String(),
		SkillID:       skillID,
		TargetType:    targetType,
		TargetName:    targetName,
		EmbeddingJSON: vecStr,
		ModelName:     modelName,
		Dimension:     dim,
		CreatedAt:     now,
	}

	switch dim {
	case 768:
		record.Embedding768 = &vecStr
	case 1408:
		record.Embedding1408 = &vecStr
	case 3072:
		record.Embedding3072 = &vecStr
	}

	return record
}

// BuildSkillResponse constructs a complete SkillResponse DTO with linked sub-entities.
func (r *SkillsRepository) BuildSkillResponse(db *gorm.DB, skill *model.Skill, similarityScore *float64) (*model.SkillResponse, error) {
	var latestVersion model.SkillVersion
	versionStr := "1.0.0"
	if skill.LatestVersion != "" {
		versionStr = skill.LatestVersion
	}
	jsonSchema := make(map[string]interface{})
	responseURI := skill.URI

	if err := db.Where("skill_id = ?", skill.ID).Order("created_at desc").First(&latestVersion).Error; err == nil {
		versionStr = latestVersion.Version
		if latestVersion.URI != "" {
			responseURI = latestVersion.URI
		}
		_ = json.Unmarshal([]byte(latestVersion.JSONSchemaJSON), &jsonSchema)
	}

	var metaRecords []model.SkillMetadata
	metadataMap := make(map[string]string)
	if err := db.Where("skill_id = ?", skill.ID).Find(&metaRecords).Error; err == nil {
		for _, m := range metaRecords {
			metadataMap[m.Key] = m.Value
		}
	}

	var resRecords []model.SkillResource
	resourcesMap := make(map[string]string)
	if err := db.Where("skill_id = ?", skill.ID).Find(&resRecords).Error; err == nil {
		for _, res := range resRecords {
			resourcesMap[res.Name] = res.Content
		}
	}

	var exRecords []model.SkillExample
	examplesMap := make(map[string]string)
	if err := db.Where("skill_id = ?", skill.ID).Find(&exRecords).Error; err == nil {
		for _, ex := range exRecords {
			examplesMap[ex.Name] = ex.Content
		}
	}

	var tagsList []string
	if skill.TagsJSON != "" {
		_ = json.Unmarshal([]byte(skill.TagsJSON), &tagsList)
	}
	if tagsList == nil {
		tagsList = []string{}
	}

	var triggerPhrasesList []string
	if skill.TriggerPhrasesJSON != "" {
		_ = json.Unmarshal([]byte(skill.TriggerPhrasesJSON), &triggerPhrasesList)
	}
	if triggerPhrasesList == nil {
		triggerPhrasesList = []string{}
	}

	return &model.SkillResponse{
		ID:              skill.ID,
		AppID:           skill.AppID,
		Name:            skill.Name,
		URI:             responseURI,
		SourceURI:       skill.SourceURI,
		Description:     skill.Description,
		Instructions:    skill.Instructions,
		License:         skill.License,
		Author:          skill.Author,
		Category:        skill.Category,
		Tags:            tagsList,
		TriggerPhrases:  triggerPhrasesList,
		Version:         versionStr,
		SHA256Hash:      skill.SHA256Hash,
		HitlTier:        skill.HitlTier,
		JSONSchema:      jsonSchema,
		Metadata:        metadataMap,
		References:      resourcesMap,
		Examples:        examplesMap,
		CreatedAt:       skill.CreatedAt,
		UpdatedAt:       skill.UpdatedAt,
		SimilarityScore: similarityScore,
	}, nil
}

// CreateSkill creates or updates a skill entity, its version, resources, and multi-modal embedding chunks.
func (r *SkillsRepository) CreateSkill(
	db *gorm.DB,
	appID string,
	req model.SkillCreateRequest,
	embeddingChunks []model.SkillEmbeddingChunk,
) (*model.SkillResponse, error) {
	// Sanitize text inputs against null bytes
	req.Name = strings.ReplaceAll(req.Name, "\x00", "")
	req.Description = strings.ReplaceAll(req.Description, "\x00", "")
	req.Instructions = strings.ReplaceAll(req.Instructions, "\x00", "")

	// 1. Resolve Application Domain
	domain := "global"
	var app model.RegisteredApp
	if err := db.Where("app_id = ?", appID).First(&app).Error; err == nil && app.Domain != "" {
		domain = app.Domain
	}

	// 2. Resolve Category & Version
	category := "general"
	if req.Category != nil && strings.TrimSpace(*req.Category) != "" {
		category = strings.ToLower(strings.TrimSpace(*req.Category))
	}
	reqCategory := category

	versionStr := "1.0.0"
	if req.Version != nil && strings.TrimSpace(*req.Version) != "" {
		versionStr = strings.TrimSpace(*req.Version)
	}

	// 3. Construct Canonical Base and Versioned URIs
	canonicalSkillURI := fmt.Sprintf("skm://skills/%s/%s/%s", domain, category, req.Name)
	versionedURI := fmt.Sprintf("skm://skills/%s/%s/%s/%s", domain, category, req.Name, versionStr)
	shaHash := ComputeSkillHash(req.Name, req.Description, req.Instructions)

	tagsBytes, _ := json.Marshal(req.Tags)
	triggersBytes, _ := json.Marshal(req.TriggerPhrases)
	now := time.Now().UTC()

	var skill model.Skill

	// Execute entire persistence in an atomic ACID transaction
	err := db.Transaction(func(tx *gorm.DB) error {
		// 4. Check if Skill already exists under this app, category, and name
		err := tx.Where("app_id = ? AND category = ? AND name = ?", appID, category, req.Name).First(&skill).Error
		if err != nil {
			// Create new skill
			skill = model.Skill{
				ID:                 uuid.New().String(),
				AppID:              appID,
				Category:           &reqCategory,
				Name:               req.Name,
				URI:                canonicalSkillURI,
				LatestVersion:      versionStr,
				SourceURI:          req.SourceURI,
				Description:        req.Description,
				Instructions:       req.Instructions,
				License:            req.License,
				Author:             req.Author,
				SHA256Hash:         shaHash,
				HitlTier:           "TIER_1_AUTO_READ",
				TagsJSON:           string(tagsBytes),
				TriggerPhrasesJSON: string(triggersBytes),
				CreatedAt:          now,
				UpdatedAt:          now,
			}
			if err := tx.Create(&skill).Error; err != nil {
				return err
			}
		} else {
			// Update existing root skill entity
			skill.Category = &reqCategory
			skill.URI = canonicalSkillURI
			skill.LatestVersion = versionStr
			skill.SourceURI = req.SourceURI
			skill.Description = req.Description
			skill.Instructions = req.Instructions
			skill.License = req.License
			skill.Author = req.Author
			skill.SHA256Hash = shaHash
			skill.TagsJSON = string(tagsBytes)
			skill.TriggerPhrasesJSON = string(triggersBytes)
			skill.UpdatedAt = now
			if err := tx.Save(&skill).Error; err != nil {
				return err
			}
		}

		// 5. Upsert Version Entry
		defaultSchema := map[string]interface{}{
			"type":        "object",
			"title":       req.Name,
			"description": req.Description,
			"properties":  map[string]interface{}{},
		}
		schemaBytes, _ := json.Marshal(defaultSchema)

		var versionEntry model.SkillVersion
		verErr := tx.Where("skill_id = ? AND version = ?", skill.ID, versionStr).First(&versionEntry).Error
		if verErr != nil {
			versionEntry = model.SkillVersion{
				ID:             uuid.New().String(),
				SkillID:        skill.ID,
				Version:        versionStr,
				URI:            versionedURI,
				JSONSchemaJSON: string(schemaBytes),
				SHA256Hash:     shaHash,
				CreatedAt:      now,
			}
			if err := tx.Create(&versionEntry).Error; err != nil {
				return err
			}
		} else {
			versionEntry.URI = versionedURI
			versionEntry.JSONSchemaJSON = string(schemaBytes)
			versionEntry.SHA256Hash = shaHash
			if err := tx.Save(&versionEntry).Error; err != nil {
				return err
			}
		}

		// 6. Refresh Metadata, Resources, Examples & Embeddings atomically
		if err := tx.Where("skill_id = ?", skill.ID).Delete(&model.SkillMetadata{}).Error; err != nil {
			return err
		}
		for k, v := range req.Metadata {
			m := model.SkillMetadata{
				ID:      uuid.New().String(),
				SkillID: skill.ID,
				Key:     strings.ReplaceAll(k, "\x00", ""),
				Value:   strings.ReplaceAll(v, "\x00", ""),
			}
			if err := tx.Create(&m).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("skill_id = ?", skill.ID).Delete(&model.SkillResource{}).Error; err != nil {
			return err
		}
		for name, content := range req.References {
			res := model.SkillResource{
				ID:      uuid.New().String(),
				SkillID: skill.ID,
				Name:    strings.ReplaceAll(name, "\x00", ""),
				Content: strings.ReplaceAll(content, "\x00", ""),
			}
			if err := tx.Create(&res).Error; err != nil {
				return err
			}
		}

		if err := tx.Where("skill_id = ?", skill.ID).Delete(&model.SkillExample{}).Error; err != nil {
			return err
		}
		for name, content := range req.Examples {
			ex := model.SkillExample{
				ID:      uuid.New().String(),
				SkillID: skill.ID,
				Name:    strings.ReplaceAll(name, "\x00", ""),
				Content: strings.ReplaceAll(content, "\x00", ""),
			}
			if err := tx.Create(&ex).Error; err != nil {
				return err
			}
		}

		if embeddingChunks != nil {
			if err := tx.Where("skill_id = ?", skill.ID).Delete(&model.SkillEmbedding{}).Error; err != nil {
				return err
			}
			for _, chunk := range embeddingChunks {
				if len(chunk.Vector) == 0 {
					continue
				}
				embRecord := buildSkillEmbeddingRecord(skill.ID, chunk.TargetType, chunk.TargetName, chunk.ModelName, chunk.Vector, now)
				if err := tx.Create(&embRecord).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return r.BuildSkillResponse(db, &skill, nil)
}

// GetSkill retrieves a skill by ID, unique name, or namespaced skm:// URI.
func (r *SkillsRepository) GetSkill(db *gorm.DB, skillIDOrNameOrURI string) (*model.SkillResponse, error) {
	var skill model.Skill
	target := strings.TrimSpace(skillIDOrNameOrURI)
	target = strings.TrimPrefix(target, "/")

	candidateURI := target
	if !strings.HasPrefix(target, "skm://") {
		candidateURI = fmt.Sprintf("skm://skills/%s", strings.TrimPrefix(target, "skills/"))
	}

	// 1. Direct match on Skill ID, Name, or base URI
	err := db.Where("id = ? OR name = ? OR uri = ? OR uri = ?", target, target, target, candidateURI).First(&skill).Error
	if err == nil {
		return r.BuildSkillResponse(db, &skill, nil)
	}

	// 2. Check match on SkillVersion URI (e.g. skm://skills/domain/cat/name/1.0.0)
	var ver model.SkillVersion
	if errVer := db.Where("uri = ? OR uri = ?", target, candidateURI).First(&ver).Error; errVer == nil {
		if errSkill := db.Where("id = ?", ver.SkillID).First(&skill).Error; errSkill == nil {
			return r.BuildSkillResponse(db, &skill, nil)
		}
	}

	return nil, fmt.Errorf("%w: %s", ErrSkillNotFound, target)
}

type ScoredSkill struct {
	Skill         *model.Skill
	Score         float64
	MatchingChunk string
}

// ListSkills lists all skills, performing vector semantic matching across multi-modal chunks or keyword matching, with bounded pagination.
func (r *SkillsRepository) ListSkills(
	db *gorm.DB,
	query string,
	queryVector []float64,
	pagination ...model.PaginationParams,
) (*model.PaginatedSkillResponse, error) {
	params := model.DefaultPagination()
	if len(pagination) > 0 {
		params = pagination[0]
	}
	params.Normalize()

	var skills []model.Skill
	if err := db.Find(&skills).Error; err != nil {
		return nil, err
	}

	if len(skills) == 0 {
		return &model.PaginatedSkillResponse{
			Items:      []*model.SkillResponse{},
			TotalCount: 0,
			Page:       params.Page,
			PageSize:   params.Limit(),
			TotalPages: 0,
		}, nil
	}

	var allMatches []*model.SkillResponse

	query = strings.TrimSpace(query)
	if query == "" {
		for i := range skills {
			resp, err := r.BuildSkillResponse(db, &skills[i], nil)
			if err == nil {
				allMatches = append(allMatches, resp)
			}
		}
	} else if len(queryVector) > 0 {
		type bestMatch struct {
			score         float64
			matchingChunk string
		}
		skillScores := make(map[string]bestMatch)

		// Check if native pgvector query on the matching dimension column is supported
		dim := len(queryVector)
		var colName string
		switch dim {
		case 768:
			colName = "embedding_768"
		case 1408:
			colName = "embedding_1408"
		case 3072:
			colName = "embedding_3072"
		}

		nativeQueried := false
		if colName != "" {
			type pgVectorMatch struct {
				SkillID    string  `gorm:"column:skill_id"`
				TargetName string  `gorm:"column:target_name"`
				Distance   float64 `gorm:"column:distance"`
			}
			var pgMatches []pgVectorMatch
			vecBytes, _ := json.Marshal(queryVector)
			querySQL := fmt.Sprintf("SELECT skill_id, target_name, (%s <=> ?) AS distance FROM skill_embeddings WHERE %s IS NOT NULL ORDER BY distance ASC LIMIT 50", colName, colName)
			if err := db.Raw(querySQL, string(vecBytes)).Scan(&pgMatches).Error; err == nil && len(pgMatches) > 0 {
				nativeQueried = true
				for _, m := range pgMatches {
					score := 1.0 - m.Distance
					curr := skillScores[m.SkillID]
					if score > curr.score || curr.matchingChunk == "" {
						skillScores[m.SkillID] = bestMatch{
							score:         score,
							matchingChunk: m.TargetName,
						}
					}
				}
			}
		}

		if !nativeQueried {
			var allEmbeddings []model.SkillEmbedding
			_ = db.Find(&allEmbeddings).Error
			for _, emb := range allEmbeddings {
				var vector []float64
				if err := json.Unmarshal([]byte(emb.EmbeddingJSON), &vector); err == nil {
					score := CosineSimilarity(queryVector, vector)
					curr := skillScores[emb.SkillID]
					if score > curr.score || curr.matchingChunk == "" {
						skillScores[emb.SkillID] = bestMatch{
							score:         score,
							matchingChunk: emb.TargetName,
						}
					}
				}
			}
		}

		qLower := strings.ToLower(query)
		qTokens := strings.Fields(qLower)

		var maxScore float64
		scored := make([]ScoredSkill, 0, len(skills))
		for i := range skills {
			s := &skills[i]
			match := skillScores[s.ID]
			finalScore := match.score

			// Lexical boost: if skill Name, Category, Tags, Description, Instructions, or MatchingChunk contains exact query terms
			nameLower := strings.ToLower(s.Name)
			descLower := strings.ToLower(s.Description)
			instLower := strings.ToLower(s.Instructions)
			chunkLower := strings.ToLower(match.matchingChunk)
			catLower := ""
			if s.Category != nil {
				catLower = strings.ToLower(*s.Category)
			}
			tagsLower := strings.ToLower(s.TagsJSON)

			lexicalMatches := 0
			for _, tok := range qTokens {
				if len(tok) < 3 {
					continue
				}
				if strings.Contains(nameLower, tok) || strings.Contains(catLower, tok) || strings.Contains(tagsLower, tok) || strings.Contains(chunkLower, tok) {
					lexicalMatches += 2
				} else if strings.Contains(descLower, tok) || strings.Contains(instLower, tok) {
					lexicalMatches++
				}
			}

			if lexicalMatches > 0 {
				boost := float64(lexicalMatches) * 0.08
				if boost > 0.35 {
					boost = 0.35
				}
				finalScore += boost
				if finalScore > 1.0 {
					finalScore = 1.0
				}
			}

			if finalScore > maxScore {
				maxScore = finalScore
			}

			scored = append(scored, ScoredSkill{
				Skill:         s,
				Score:         finalScore,
				MatchingChunk: match.matchingChunk,
			})
		}

		// Prune results with low semantic similarity while preserving valid lexical/keyword matches
		threshold := 0.10
		if maxScore > 0.45 {
			threshold = maxScore - 0.20
		} else if maxScore > 0 {
			threshold = maxScore * 0.4
		}

		filtered := make([]ScoredSkill, 0, len(scored))
		for _, item := range scored {
			if item.Score >= threshold {
				filtered = append(filtered, item)
			}
		}

		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].Score > filtered[j].Score
		})

		for _, item := range filtered {
			scoreVal := item.Score
			var chunkPtr *string
			if item.MatchingChunk != "" {
				chunkPtr = &item.MatchingChunk
			}
			resp, err := r.BuildSkillResponse(db, item.Skill, &scoreVal)
			if err == nil {
				resp.MatchingChunk = chunkPtr
				allMatches = append(allMatches, resp)
			}
		}
	} else {
		qLower := strings.ToLower(query)
		for i := range skills {
			s := &skills[i]
			catStr := ""
			if s.Category != nil {
				catStr = strings.ToLower(*s.Category)
			}
			if strings.Contains(strings.ToLower(s.Name), qLower) ||
				strings.Contains(strings.ToLower(s.Description), qLower) ||
				strings.Contains(strings.ToLower(s.Instructions), qLower) ||
				(catStr != "" && strings.Contains(catStr, qLower)) {
				resp, err := r.BuildSkillResponse(db, s, nil)
				if err == nil {
					allMatches = append(allMatches, resp)
				}
			}
		}
	}

	totalCount := int64(len(allMatches))
	limit := params.Limit()
	offset := params.Offset()

	totalPages := 0
	if totalCount > 0 {
		totalPages = int(math.Ceil(float64(totalCount) / float64(limit)))
	}

	var paged []*model.SkillResponse
	if offset < len(allMatches) {
		end := offset + limit
		if end > len(allMatches) {
			end = len(allMatches)
		}
		paged = allMatches[offset:end]
	} else {
		paged = []*model.SkillResponse{}
	}

	return &model.PaginatedSkillResponse{
		Items:      paged,
		TotalCount: totalCount,
		Page:       params.Page,
		PageSize:   limit,
		TotalPages: totalPages,
	}, nil
}

// UpdateSkill updates or replaces an existing skill, enforcing app ownership.
func (r *SkillsRepository) UpdateSkill(
	db *gorm.DB,
	skillID string,
	appID string,
	req model.SkillUpdateRequest,
	fullReplace bool,
	embeddingChunks []model.SkillEmbeddingChunk,
) (*model.SkillResponse, error) {
	var skill model.Skill
	if err := db.Where("id = ?", skillID).First(&skill).Error; err != nil {
		return nil, fmt.Errorf("%w: %s", ErrSkillNotFound, skillID)
	}

	if skill.AppID != appID {
		return nil, ErrUnauthorizedSkillAccess
	}

	if req.Description != nil || fullReplace {
		if req.Description != nil {
			skill.Description = *req.Description
		}
	}
	if req.Instructions != nil || fullReplace {
		if req.Instructions != nil {
			skill.Instructions = *req.Instructions
		}
	}
	if req.License != nil || fullReplace {
		skill.License = req.License
	}
	if req.Category != nil || fullReplace {
		skill.Category = req.Category
	}
	if req.Tags != nil || fullReplace {
		tags := []string{}
		if req.Tags != nil {
			tags = *req.Tags
		}
		tagsBytes, _ := json.Marshal(tags)
		skill.TagsJSON = string(tagsBytes)
	}
	if req.TriggerPhrases != nil || fullReplace {
		triggers := []string{}
		if req.TriggerPhrases != nil {
			triggers = *req.TriggerPhrases
		}
		trigBytes, _ := json.Marshal(triggers)
		skill.TriggerPhrasesJSON = string(trigBytes)
	}

	now := time.Now().UTC()

	err := db.Transaction(func(tx *gorm.DB) error {
		skill.UpdatedAt = now
		skill.SHA256Hash = ComputeSkillHash(skill.Name, skill.Description, skill.Instructions)

		if err := tx.Save(&skill).Error; err != nil {
			return err
		}

		if req.Version != nil && *req.Version != "" {
			var versionEntry model.SkillVersion
			verErr := tx.Where("skill_id = ? AND version = ?", skill.ID, *req.Version).First(&versionEntry).Error
			if verErr != nil {
				versionEntry = model.SkillVersion{
					ID:             uuid.New().String(),
					SkillID:        skill.ID,
					Version:        *req.Version,
					URI:            fmt.Sprintf("%s/%s", skill.URI, *req.Version),
					JSONSchemaJSON: fmt.Sprintf(`{"type":"object","title":%q,"description":%q}`, skill.Name, skill.Description),
					SHA256Hash:     skill.SHA256Hash,
					CreatedAt:      now,
				}
				if err := tx.Create(&versionEntry).Error; err != nil {
					return err
				}
			} else {
				versionEntry.SHA256Hash = skill.SHA256Hash
				versionEntry.JSONSchemaJSON = fmt.Sprintf(`{"type":"object","title":%q,"description":%q}`, skill.Name, skill.Description)
				if err := tx.Save(&versionEntry).Error; err != nil {
					return err
				}
			}
		}

		if req.Metadata != nil {
			if err := tx.Where("skill_id = ?", skill.ID).Delete(&model.SkillMetadata{}).Error; err != nil {
				return err
			}
			for k, v := range *req.Metadata {
				m := model.SkillMetadata{
					ID:      uuid.New().String(),
					SkillID: skill.ID,
					Key:     k,
					Value:   v,
				}
				if err := tx.Create(&m).Error; err != nil {
					return err
				}
			}
		}

		if req.References != nil {
			if err := tx.Where("skill_id = ?", skill.ID).Delete(&model.SkillResource{}).Error; err != nil {
				return err
			}
			for name, content := range *req.References {
				res := model.SkillResource{
					ID:      uuid.New().String(),
					SkillID: skill.ID,
					Name:    name,
					Content: content,
				}
				if err := tx.Create(&res).Error; err != nil {
					return err
				}
			}
		}

		if req.Examples != nil {
			if err := tx.Where("skill_id = ?", skill.ID).Delete(&model.SkillExample{}).Error; err != nil {
				return err
			}
			for name, content := range *req.Examples {
				ex := model.SkillExample{
					ID:      uuid.New().String(),
					SkillID: skill.ID,
					Name:    name,
					Content: content,
				}
				if err := tx.Create(&ex).Error; err != nil {
					return err
				}
			}
		}

		if len(embeddingChunks) > 0 {
			if err := tx.Where("skill_id = ?", skill.ID).Delete(&model.SkillEmbedding{}).Error; err != nil {
				return err
			}
			for _, chunk := range embeddingChunks {
				if len(chunk.Vector) == 0 {
					continue
				}
				embRecord := buildSkillEmbeddingRecord(skill.ID, chunk.TargetType, chunk.TargetName, chunk.ModelName, chunk.Vector, now)
				if err := tx.Create(&embRecord).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return r.BuildSkillResponse(db, &skill, nil)
}

// SaveSkillEmbeddings persists embedding chunks for a skill inside an atomic ACID transaction.
func (r *SkillsRepository) SaveSkillEmbeddings(
	db *gorm.DB,
	skillID string,
	embeddingChunks []model.SkillEmbeddingChunk,
) error {
	now := time.Now().UTC()
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("skill_id = ?", skillID).Delete(&model.SkillEmbedding{}).Error; err != nil {
			return err
		}
		for _, chunk := range embeddingChunks {
			if len(chunk.Vector) == 0 {
				continue
			}
			embRecord := buildSkillEmbeddingRecord(skillID, chunk.TargetType, chunk.TargetName, chunk.ModelName, chunk.Vector, now)
			if err := tx.Create(&embRecord).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteSkill deletes a skill and all linked sub-entities.
func (r *SkillsRepository) DeleteSkill(db *gorm.DB, skillID string, appID string) (map[string]interface{}, error) {
	var skill model.Skill
	if err := db.Where("id = ?", skillID).First(&skill).Error; err != nil {
		return nil, fmt.Errorf("%w: %s", ErrSkillNotFound, skillID)
	}

	if skill.AppID != appID {
		return nil, ErrUnauthorizedSkillAccess
	}

	_ = db.Where("skill_id = ?", skill.ID).Delete(&model.SkillVersion{})
	_ = db.Where("skill_id = ?", skill.ID).Delete(&model.SkillMetadata{})
	_ = db.Where("skill_id = ?", skill.ID).Delete(&model.SkillResource{})
	_ = db.Where("skill_id = ?", skill.ID).Delete(&model.SkillExample{})
	_ = db.Where("skill_id = ?", skill.ID).Delete(&model.SkillEmbedding{})

	if err := db.Delete(&skill).Error; err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"status":  "success",
		"message": fmt.Sprintf("Skill '%s' deleted successfully.", skillID),
	}, nil
}
