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

	"github.com/retail-cortex/skills/pkg/embedding"
	"github.com/retail-cortex/skills/pkg/embedding/vertex"
	"github.com/retail-cortex/skills/pkg/model"
)

// EmbeddingConfig configures Vertex AI and Gemini embedding generation.
type EmbeddingConfig struct {
	ModelName    string
	ProjectID    string
	Region       string
	GeminiAPIKey string
	BaseURL      string
}

// EmbeddingService is a service adapter over vertex.Provider.
type EmbeddingService struct {
	provider *vertex.Provider
}

// NewEmbeddingService creates a new EmbeddingService wrapping vertex.Provider.
func NewEmbeddingService(cfg ...EmbeddingConfig) *EmbeddingService {
	var vCfg vertex.Config
	if len(cfg) > 0 {
		vCfg = vertex.Config{
			ModelName:    cfg[0].ModelName,
			ProjectID:    cfg[0].ProjectID,
			Region:       cfg[0].Region,
			GeminiAPIKey: cfg[0].GeminiAPIKey,
			BaseURL:      cfg[0].BaseURL,
		}
	}
	return &EmbeddingService{
		provider: vertex.NewProvider(vCfg),
	}
}

func (s *EmbeddingService) Provider() embedding.Provider {
	return s.provider
}

func (s *EmbeddingService) ModelName() string {
	return s.provider.ModelName
}

func (s *EmbeddingService) GenerateEmbedding(text string) []float64 {
	vec, _ := s.provider.GenerateEmbedding(context.Background(), text)
	return vec
}

func (s *EmbeddingService) GenerateImageEmbedding(base64Image string) []float64 {
	vec, _ := s.provider.GenerateImageEmbedding(context.Background(), base64Image)
	return vec
}

func (s *EmbeddingService) GenerateSkillEmbeddings(
	name, description, instructions string,
	triggers []string,
	references map[string]string,
	examples map[string]string,
) []model.SkillEmbeddingChunk {
	chunks, _ := s.provider.GenerateSkillEmbeddings(
		context.Background(),
		name,
		description,
		instructions,
		triggers,
		references,
		examples,
	)
	return chunks
}

func (s *EmbeddingService) CosineSimilarity(a, b []float64) float64 {
	return s.provider.CosineSimilarity(a, b)
}
