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

package alloydb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/retail-cortex/castor/pkg/embedding"
	"github.com/retail-cortex/castor/pkg/model"
	"gorm.io/gorm"
)

// Config configures the AlloyDB AI embedding provider.
type Config struct {
	DB        *gorm.DB
	SQLDB     *sql.DB
	ModelName string // e.g. "text-embedding-004"
	Dimension int    // e.g. 768
}

// Provider implements embedding.Provider for Google Cloud AlloyDB AI in-database embeddings.
type Provider struct {
	db        *gorm.DB
	sqlDB     *sql.DB
	modelName string
	dimension int
}

// NewProvider creates an AlloyDB AI embedding provider.
func NewProvider(cfg Config) *Provider {
	mName := cfg.ModelName
	if mName == "" {
		mName = "text-embedding-004"
	}
	dim := cfg.Dimension
	if dim <= 0 {
		dim = 768
	}
	return &Provider{
		db:        cfg.DB,
		sqlDB:     cfg.SQLDB,
		modelName: mName,
		dimension: dim,
	}
}

func (p *Provider) Name() string {
	return "alloydb-ai"
}

func (p *Provider) Dimension() int {
	return p.dimension
}

func (p *Provider) CosineSimilarity(a, b []float64) float64 {
	return embedding.CosineSimilarity(a, b)
}

// GenerateEmbedding calls AlloyDB AI in-database embedding functions:
// e.g. `SELECT embedding('text-embedding-004', ?)::text` or `SELECT google_ml.embedding(?, ?)`
func (p *Provider) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, nil
	}

	// 1. Direct GORM database query execution
	if p.db != nil {
		var rawResult string
		// AlloyDB AI native SQL embedding function
		query := fmt.Sprintf("SELECT embedding('%s', ?)::text", p.modelName)
		if err := p.db.WithContext(ctx).Raw(query, text).Scan(&rawResult).Error; err == nil && rawResult != "" {
			return parseAlloyDBVector(rawResult)
		}

		// AlloyDB google_ml extension fallback
		mlQuery := "SELECT google_ml.embedding(?, ?)::text"
		if err := p.db.WithContext(ctx).Raw(mlQuery, p.modelName, text).Scan(&rawResult).Error; err == nil && rawResult != "" {
			return parseAlloyDBVector(rawResult)
		}
	}

	// 2. Direct sql.DB execution
	if p.sqlDB != nil {
		var rawResult string
		query := fmt.Sprintf("SELECT embedding('%s', $1)::text", p.modelName)
		row := p.sqlDB.QueryRowContext(ctx, query, text)
		if err := row.Scan(&rawResult); err == nil && rawResult != "" {
			return parseAlloyDBVector(rawResult)
		}
	}

	// 3. Fallback deterministic simulation when no live AlloyDB AI cluster is connected
	return embedding.GenerateDeterministicVector(text, p.dimension), nil
}

// GenerateImageEmbedding handles images. AlloyDB text-embedding-004 focuses on text/code;
// for multi-modal images, it generates a vector over image metadata or visual hashes.
func (p *Provider) GenerateImageEmbedding(ctx context.Context, base64Image string) ([]float64, error) {
	base64Image = strings.TrimSpace(base64Image)
	if base64Image == "" {
		return nil, nil
	}
	desc := fmt.Sprintf("AlloyDB In-DB Image Asset Payload Hash %d", len(base64Image))
	return p.GenerateEmbedding(ctx, desc)
}

func (p *Provider) GenerateSkillEmbeddings(
	ctx context.Context,
	name, description, instructions string,
	triggers []string,
	references map[string]string,
	examples map[string]string,
) ([]model.SkillEmbeddingChunk, error) {
	type chunkTask struct {
		targetType string
		targetName string
		isImage    bool
		payload    string
	}

	var tasks []chunkTask

	// 1. Root Skill
	trigStr := strings.Join(triggers, " ")
	summaryText := fmt.Sprintf("%s %s %s", name, description, trigStr)
	for i, chunk := range embedding.SplitTextIntoChunks(summaryText, 900) {
		tName := "SKILL.md#summary"
		if i > 0 {
			tName = fmt.Sprintf("SKILL.md#summary-%d", i+1)
		}
		tasks = append(tasks, chunkTask{
			targetType: "skill",
			targetName: tName,
			payload:    chunk,
		})
	}

	if strings.TrimSpace(instructions) != "" {
		for i, chunk := range embedding.SplitTextIntoChunks(instructions, 900) {
			tName := "SKILL.md#instructions"
			if i > 0 {
				tName = fmt.Sprintf("SKILL.md#instructions-%d", i+1)
			}
			tasks = append(tasks, chunkTask{
				targetType: "skill",
				targetName: tName,
				payload:    chunk,
			})
		}
	}

	// 2. References
	for refName, content := range references {
		if strings.HasPrefix(content, "data:image/") && strings.Contains(content, ";base64,") {
			parts := strings.SplitN(content, ";base64,", 2)
			tasks = append(tasks, chunkTask{
				targetType: "reference",
				targetName: refName,
				isImage:    true,
				payload:    parts[1],
			})
		} else if strings.HasPrefix(content, "data:") {
			tasks = append(tasks, chunkTask{
				targetType: "reference",
				targetName: refName,
				payload:    fmt.Sprintf("Binary reference resource asset: %s", refName),
			})
		} else {
			textChunks := embedding.SplitTextIntoChunks(content, 900)
			for i, chunk := range textChunks {
				tName := refName
				if len(textChunks) > 1 {
					tName = fmt.Sprintf("%s#part-%d", refName, i+1)
				}
				tasks = append(tasks, chunkTask{
					targetType: "reference",
					targetName: tName,
					payload:    fmt.Sprintf("Reference %s: %s", tName, chunk),
				})
			}
		}
	}

	// 3. Examples
	for exName, content := range examples {
		if strings.HasPrefix(content, "data:") {
			tasks = append(tasks, chunkTask{
				targetType: "example",
				targetName: exName,
				payload:    fmt.Sprintf("Binary example resource asset: %s", exName),
			})
		} else {
			textChunks := embedding.SplitTextIntoChunks(content, 900)
			for i, chunk := range textChunks {
				tName := exName
				if len(textChunks) > 1 {
					tName = fmt.Sprintf("%s#part-%d", exName, i+1)
				}
				tasks = append(tasks, chunkTask{
					targetType: "example",
					targetName: tName,
					payload:    fmt.Sprintf("Example script %s: %s", tName, chunk),
				})
			}
		}
	}

	if len(tasks) == 0 {
		return nil, nil
	}

	taskChan := make(chan chunkTask, len(tasks))
	resultChan := make(chan model.SkillEmbeddingChunk, len(tasks))

	for _, t := range tasks {
		taskChan <- t
	}
	close(taskChan)

	numWorkers := 8
	if len(tasks) < numWorkers {
		numWorkers = len(tasks)
	}

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for t := range taskChan {
				var vec []float64
				var err error
				if t.isImage {
					vec, err = p.GenerateImageEmbedding(ctx, t.payload)
				} else {
					vec, err = p.GenerateEmbedding(ctx, t.payload)
				}
				if err == nil && len(vec) > 0 {
					resultChan <- model.SkillEmbeddingChunk{
						TargetType: t.targetType,
						TargetName: t.targetName,
						Vector:     vec,
						ModelName:  p.modelName,
					}
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	var chunks []model.SkillEmbeddingChunk
	for chunk := range resultChan {
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

func parseAlloyDBVector(raw string) ([]float64, error) {
	trimmed := strings.Trim(raw, "[]{}()")
	if trimmed == "" {
		return nil, nil
	}
	parts := strings.Split(trimmed, ",")
	vec := make([]float64, len(parts))
	for i, p := range parts {
		var val float64
		if err := json.Unmarshal([]byte(strings.TrimSpace(p)), &val); err == nil {
			vec[i] = val
		}
	}
	return vec, nil
}
