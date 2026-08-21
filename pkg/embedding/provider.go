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

package embedding

import (
	"context"
	"math"
	"strings"

	"github.com/retail-cortex/castor/pkg/model"
)

// Provider defines the standard interface for embedding generation across cloud, database, and local models.
type Provider interface {
	// Name returns the provider identifier (e.g. "vertex-gemini", "alloydb-ai").
	Name() string

	// Dimension returns the vector dimension produced by this provider.
	Dimension() int

	// GenerateEmbedding generates a vector embedding for a single text input.
	GenerateEmbedding(ctx context.Context, text string) ([]float64, error)

	// GenerateImageEmbedding generates a vector embedding for a base64-encoded image input.
	GenerateImageEmbedding(ctx context.Context, base64Image string) ([]float64, error)

	// GenerateSkillEmbeddings decomposes and generates granular multi-chunk embeddings for an entire skill.
	GenerateSkillEmbeddings(
		ctx context.Context,
		name, description, instructions string,
		triggers []string,
		references map[string]string,
		examples map[string]string,
	) ([]model.SkillEmbeddingChunk, error)

	// CosineSimilarity computes the cosine similarity between two vector embeddings.
	CosineSimilarity(a, b []float64) float64
}

// CosineSimilarity computes normalized cosine distance between two float vectors.
func CosineSimilarity(a, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0.0
	}
	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0.0 || normB == 0.0 {
		return 0.0
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

// SplitTextIntoChunks splits long text into chunks of at most maxChars with sliding window overlap.
func SplitTextIntoChunks(text string, maxChars int) []string {
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return nil
	}
	if len(text) <= maxChars {
		return []string{text}
	}

	var chunks []string
	paragraphs := strings.Split(text, "\n\n")
	var current strings.Builder

	for _, p := range paragraphs {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if current.Len()+len(p)+2 <= maxChars {
			if current.Len() > 0 {
				current.WriteString("\n\n")
			}
			current.WriteString(p)
		} else {
			if current.Len() > 0 {
				chunks = append(chunks, current.String())
				current.Reset()
			}
			if len(p) <= maxChars {
				current.WriteString(p)
			} else {
				// Sliding window slicing for oversized paragraphs / code blocks
				runes := []rune(p)
				for len(runes) > 0 {
					take := maxChars
					if len(runes) < take {
						take = len(runes)
					}
					chunks = append(chunks, string(runes[:take]))
					if len(runes) <= maxChars {
						break
					}
					step := maxChars - 80
					if step <= 0 || step >= len(runes) {
						break
					}
					runes = runes[step:]
				}
			}
		}
	}
	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}
	return chunks
}

// GenerateDeterministicVector creates a normalized semantic vector for a text string.
func GenerateDeterministicVector(text string, dim int) []float64 {
	vec := make([]float64, dim)
	var norm float64
	tokens := strings.Fields(strings.ToLower(text))
	for tIdx, tok := range tokens {
		for i, r := range tok {
			idx := (tIdx*37 + i*31 + int(r)) % dim
			vec[idx] += 1.0 / (float64(i%5) + 1.0)
		}
	}
	for _, v := range vec {
		norm += v * v
	}
	if norm > 0 {
		norm = math.Sqrt(norm)
		for i := range vec {
			vec[i] /= norm
		}
	}
	return vec
}
