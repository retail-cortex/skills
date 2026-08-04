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
	"github.com/retail-cortex/skills/pkg/model"
	"gorm.io/gorm"
)

var (
	ErrSkillNotFound           = errors.New("skill not found")
	ErrUnauthorizedSkillAccess = errors.New("unauthorized: skill belongs to a different application")
)

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
	if len(vec1) == 0 || len(vec2) == 0 || len(vec1) != len(vec2) {
		return 0.0
	}
	var dotProduct, norm1, norm2 float64
	for i := range vec1 {
		dotProduct += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}
	if norm1 == 0.0 || norm2 == 0.0 {
		return 0.0
	}
	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// BuildSkillResponse constructs a complete SkillResponse DTO with linked sub-entities.
func (r *SkillsRepository) BuildSkillResponse(db *gorm.DB, skill *model.Skill, similarityScore *float64) (*model.SkillResponse, error) {
	var latestVersion model.SkillVersion
	versionStr := "1.0.0"
	jsonSchema := make(map[string]interface{})

	if err := db.Where("skill_id = ?", skill.ID).Order("created_at desc").First(&latestVersion).Error; err == nil {
		versionStr = latestVersion.Version
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
		URI:             skill.URI,
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

// CreateSkill persists a new skill, compiles schema, and stores optional embeddings.
func (r *SkillsRepository) CreateSkill(
	db *gorm.DB,
	appID string,
	req model.SkillCreateRequest,
	embeddingVector []float64,
	modelName string,
) (*model.SkillResponse, error) {
	skillID := uuid.New().String()
	skmURI := fmt.Sprintf("skm://skills/%s", skillID)
	shaHash := ComputeSkillHash(req.Name, req.Description, req.Instructions)

	tagsBytes, _ := json.Marshal(req.Tags)
	triggersBytes, _ := json.Marshal(req.TriggerPhrases)

	now := time.Now().UTC()
	skill := model.Skill{
		ID:                 skillID,
		AppID:              appID,
		Name:               req.Name,
		URI:                skmURI,
		SourceURI:          req.SourceURI,
		Description:        req.Description,
		Instructions:       req.Instructions,
		License:            req.License,
		Author:             req.Author,
		Category:           req.Category,
		SHA256Hash:         shaHash,
		HitlTier:           "TIER_1_AUTO_READ",
		TagsJSON:           string(tagsBytes),
		TriggerPhrasesJSON: string(triggersBytes),
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := db.Create(&skill).Error; err != nil {
		return nil, err
	}

	versionStr := "1.0.0"
	if req.Version != nil && *req.Version != "" {
		versionStr = *req.Version
	}

	defaultSchema := map[string]interface{}{
		"type":        "object",
		"title":       req.Name,
		"description": req.Description,
		"properties":  map[string]interface{}{},
	}
	schemaBytes, _ := json.Marshal(defaultSchema)

	versionEntry := model.SkillVersion{
		ID:             uuid.New().String(),
		SkillID:        skillID,
		Version:        versionStr,
		JSONSchemaJSON: string(schemaBytes),
		SHA256Hash:     shaHash,
		CreatedAt:      now,
	}
	if err := db.Create(&versionEntry).Error; err != nil {
		return nil, err
	}

	for k, v := range req.Metadata {
		m := model.SkillMetadata{
			ID:      uuid.New().String(),
			SkillID: skillID,
			Key:     k,
			Value:   v,
		}
		_ = db.Create(&m)
	}

	for name, content := range req.References {
		res := model.SkillResource{
			ID:      uuid.New().String(),
			SkillID: skillID,
			Name:    name,
			Content: content,
		}
		_ = db.Create(&res)
	}

	for name, content := range req.Examples {
		ex := model.SkillExample{
			ID:      uuid.New().String(),
			SkillID: skillID,
			Name:    name,
			Content: content,
		}
		_ = db.Create(&ex)
	}

	if len(embeddingVector) > 0 {
		embBytes, _ := json.Marshal(embeddingVector)
		mName := modelName
		if mName == "" {
			mName = "text-embedding-004"
		}
		embRecord := model.SkillEmbedding{
			ID:            uuid.New().String(),
			SkillID:       skillID,
			EmbeddingJSON: string(embBytes),
			ModelName:     mName,
			CreatedAt:     now,
		}
		_ = db.Create(&embRecord)
	}

	return r.BuildSkillResponse(db, &skill, nil)
}

// GetSkill retrieves a skill by ID, unique name, or skm:// URI.
func (r *SkillsRepository) GetSkill(db *gorm.DB, skillIDOrName string) (*model.SkillResponse, error) {
	var skill model.Skill
	err := db.Where("id = ? OR name = ? OR uri = ?", skillIDOrName, skillIDOrName, skillIDOrName).First(&skill).Error
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrSkillNotFound, skillIDOrName)
	}
	return r.BuildSkillResponse(db, &skill, nil)
}

type ScoredSkill struct {
	Skill *model.Skill
	Score float64
}

// ListSkills lists all skills, performing vector semantic matching or keyword matching.
func (r *SkillsRepository) ListSkills(
	db *gorm.DB,
	query string,
	queryVector []float64,
) ([]*model.SkillResponse, error) {
	var skills []model.Skill
	if err := db.Find(&skills).Error; err != nil {
		return nil, err
	}

	if len(skills) == 0 {
		return []*model.SkillResponse{}, nil
	}

	query = strings.TrimSpace(query)
	if query == "" {
		responses := make([]*model.SkillResponse, 0, len(skills))
		for i := range skills {
			resp, err := r.BuildSkillResponse(db, &skills[i], nil)
			if err == nil {
				responses = append(responses, resp)
			}
		}
		return responses, nil
	}

	if len(queryVector) > 0 {
		scored := make([]ScoredSkill, 0, len(skills))
		for i := range skills {
			var embRecord model.SkillEmbedding
			score := 0.0
			if err := db.Where("skill_id = ?", skills[i].ID).First(&embRecord).Error; err == nil {
				var vector []float64
				if err := json.Unmarshal([]byte(embRecord.EmbeddingJSON), &vector); err == nil {
					score = CosineSimilarity(queryVector, vector)
				}
			}
			scored = append(scored, ScoredSkill{Skill: &skills[i], Score: score})
		}

		sort.Slice(scored, func(i, j int) bool {
			return scored[i].Score > scored[j].Score
		})

		responses := make([]*model.SkillResponse, 0, len(scored))
		for _, item := range scored {
			scoreVal := item.Score
			resp, err := r.BuildSkillResponse(db, item.Skill, &scoreVal)
			if err == nil {
				responses = append(responses, resp)
			}
		}
		return responses, nil
	}

	qLower := strings.ToLower(query)
	responses := make([]*model.SkillResponse, 0)
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
				responses = append(responses, resp)
			}
		}
	}
	return responses, nil
}

// UpdateSkill updates or replaces an existing skill, enforcing app ownership.
func (r *SkillsRepository) UpdateSkill(
	db *gorm.DB,
	skillID string,
	appID string,
	req model.SkillUpdateRequest,
	fullReplace bool,
	embeddingVector []float64,
	modelName string,
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
	skill.UpdatedAt = now
	skill.SHA256Hash = ComputeSkillHash(skill.Name, skill.Description, skill.Instructions)

	if err := db.Save(&skill).Error; err != nil {
		return nil, err
	}

	if req.Version != nil && *req.Version != "" {
		versionEntry := model.SkillVersion{
			ID:             uuid.New().String(),
			SkillID:        skill.ID,
			Version:        *req.Version,
			JSONSchemaJSON: fmt.Sprintf(`{"type":"object","title":%q,"description":%q}`, skill.Name, skill.Description),
			SHA256Hash:     skill.SHA256Hash,
			CreatedAt:      now,
		}
		_ = db.Create(&versionEntry)
	}

	if req.Metadata != nil {
		_ = db.Where("skill_id = ?", skill.ID).Delete(&model.SkillMetadata{})
		for k, v := range *req.Metadata {
			m := model.SkillMetadata{
				ID:      uuid.New().String(),
				SkillID: skill.ID,
				Key:     k,
				Value:   v,
			}
			_ = db.Create(&m)
		}
	}

	if req.References != nil {
		_ = db.Where("skill_id = ?", skill.ID).Delete(&model.SkillResource{})
		for name, content := range *req.References {
			res := model.SkillResource{
				ID:      uuid.New().String(),
				SkillID: skill.ID,
				Name:    name,
				Content: content,
			}
			_ = db.Create(&res)
		}
	}

	if req.Examples != nil {
		_ = db.Where("skill_id = ?", skill.ID).Delete(&model.SkillExample{})
		for name, content := range *req.Examples {
			ex := model.SkillExample{
				ID:      uuid.New().String(),
				SkillID: skill.ID,
				Name:    name,
				Content: content,
			}
			_ = db.Create(&ex)
		}
	}

	if len(embeddingVector) > 0 {
		_ = db.Where("skill_id = ?", skill.ID).Delete(&model.SkillEmbedding{})
		embBytes, _ := json.Marshal(embeddingVector)
		mName := modelName
		if mName == "" {
			mName = "text-embedding-004"
		}
		embRecord := model.SkillEmbedding{
			ID:            uuid.New().String(),
			SkillID:       skill.ID,
			EmbeddingJSON: string(embBytes),
			ModelName:     mName,
			CreatedAt:     now,
		}
		_ = db.Create(&embRecord)
	}

	return r.BuildSkillResponse(db, &skill, nil)
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
