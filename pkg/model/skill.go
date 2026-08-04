package model

import (
	"time"
)

// Skill is the primary skill persistence record linked to a registered app.
type Skill struct {
	ID                 string    `gorm:"primaryKey;column:id" json:"id"`
	AppID              string    `gorm:"index;column:app_id" json:"app_id"`
	Name               string    `gorm:"index;column:name" json:"name"`
	URI                string    `gorm:"index;column:uri" json:"uri"`
	SourceURI          string    `gorm:"column:source_uri" json:"source_uri"`
	Description        string    `gorm:"column:description" json:"description"`
	Instructions       string    `gorm:"column:instructions" json:"instructions"`
	License            *string   `gorm:"column:license" json:"license,omitempty"`
	Author             *string   `gorm:"column:author" json:"author,omitempty"`
	Category           *string   `gorm:"index;column:category" json:"category,omitempty"`
	SHA256Hash         string    `gorm:"index;column:sha256_hash" json:"sha256_hash"`
	HitlTier           string    `gorm:"column:hitl_tier;default:'TIER_1_AUTO_READ'" json:"hitl_tier"`
	TagsJSON           string    `gorm:"column:tags_json;default:'[]'" json:"tags_json"`
	TriggerPhrasesJSON string    `gorm:"column:trigger_phrases_json;default:'[]'" json:"trigger_phrases_json"`
	CreatedAt          time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Skill) TableName() string { return "skills" }

// SkillVersion is a historical version entry for a skill.
type SkillVersion struct {
	ID             string    `gorm:"primaryKey;column:id" json:"id"`
	SkillID        string    `gorm:"index;column:skill_id" json:"skill_id"`
	Version        string    `gorm:"index;column:version" json:"version"`
	JSONSchemaJSON string    `gorm:"column:json_schema_json" json:"json_schema_json"`
	SHA256Hash     string    `gorm:"column:sha256_hash" json:"sha256_hash"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
}

func (SkillVersion) TableName() string { return "skill_versions" }

// SkillMetadata represents key-value metadata linked to a skill.
type SkillMetadata struct {
	ID      string `gorm:"primaryKey;column:id" json:"id"`
	SkillID string `gorm:"index;column:skill_id" json:"skill_id"`
	Key     string `gorm:"index;column:key" json:"key"`
	Value   string `gorm:"column:value" json:"value"`
}

func (SkillMetadata) TableName() string { return "skill_metadata" }

// SkillResource represents a reference file or resource linked to a skill.
type SkillResource struct {
	ID      string `gorm:"primaryKey;column:id" json:"id"`
	SkillID string `gorm:"index;column:skill_id" json:"skill_id"`
	Name    string `gorm:"index;column:name" json:"name"`
	Content string `gorm:"column:content" json:"content"`
}

func (SkillResource) TableName() string { return "skill_resources" }

// SkillExample represents a usage example content linked to a skill.
type SkillExample struct {
	ID      string `gorm:"primaryKey;column:id" json:"id"`
	SkillID string `gorm:"index;column:skill_id" json:"skill_id"`
	Name    string `gorm:"index;column:name" json:"name"`
	Content string `gorm:"column:content" json:"content"`
}

func (SkillExample) TableName() string { return "skill_examples" }

// SkillEmbedding represents a vector embedding record for semantic search.
type SkillEmbedding struct {
	ID            string    `gorm:"primaryKey;column:id" json:"id"`
	SkillID       string    `gorm:"index;column:skill_id" json:"skill_id"`
	EmbeddingJSON string    `gorm:"column:embedding_json" json:"embedding_json"`
	ModelName     string    `gorm:"column:model_name;default:'text-embedding-004'" json:"model_name"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
}

func (SkillEmbedding) TableName() string { return "skill_embeddings" }

// DTOs

// SkillCreateRequest is the payload for creating or registering a skill.
type SkillCreateRequest struct {
	Name           string            `json:"name" binding:"required"`
	SourceURI      string            `json:"source_uri,omitempty"`
	Description    string            `json:"description" binding:"required"`
	Instructions   string            `json:"instructions" binding:"required"`
	License        *string           `json:"license,omitempty"`
	Author         *string           `json:"author,omitempty"`
	Version        *string           `json:"version,omitempty"`
	Category       *string           `json:"category,omitempty"`
	Tags           []string          `json:"tags,omitempty"`
	TriggerPhrases []string          `json:"trigger_phrases,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
	References     map[string]string `json:"references,omitempty"`
	Examples       map[string]string `json:"examples,omitempty"`
}

// SkillUpdateRequest is the payload for updating an existing skill.
type SkillUpdateRequest struct {
	Description    *string            `json:"description,omitempty"`
	Instructions   *string            `json:"instructions,omitempty"`
	License        *string            `json:"license,omitempty"`
	Category       *string            `json:"category,omitempty"`
	Tags           *[]string          `json:"tags,omitempty"`
	TriggerPhrases *[]string          `json:"trigger_phrases,omitempty"`
	Version        *string            `json:"version,omitempty"`
	Metadata       *map[string]string `json:"metadata,omitempty"`
	References     *map[string]string `json:"references,omitempty"`
	Examples       *map[string]string `json:"examples,omitempty"`
}

// SkillResponse is the DTO returned when retrieving a skill.
type SkillResponse struct {
	ID              string                 `json:"id"`
	AppID           string                 `json:"app_id"`
	Name            string                 `json:"name"`
	URI             string                 `json:"uri"`
	SourceURI       string                 `json:"source_uri,omitempty"`
	Description     string                 `json:"description"`
	Instructions    string                 `json:"instructions"`
	License         *string                `json:"license,omitempty"`
	Author          *string                `json:"author,omitempty"`
	Category        *string                `json:"category,omitempty"`
	Tags            []string               `json:"tags"`
	TriggerPhrases  []string               `json:"trigger_phrases"`
	Version         string                 `json:"version"`
	SHA256Hash      string                 `json:"sha256_hash"`
	HitlTier        string                 `json:"hitl_tier"`
	JSONSchema      map[string]interface{} `json:"json_schema"`
	Metadata        map[string]string      `json:"metadata"`
	References      map[string]string      `json:"references"`
	Examples        map[string]string      `json:"examples"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
	SimilarityScore *float64               `json:"similarity_score,omitempty"`
}

