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

package embedding_test

import (
	"context"
	"testing"

	"github.com/retail-cortex/skills/pkg/embedding"
	"github.com/retail-cortex/skills/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockProvider struct {
	name      string
	dimension int
}

func (m *mockProvider) Name() string     { return m.name }
func (m *mockProvider) Dimension() int  { return m.dimension }
func (m *mockProvider) CosineSimilarity(a, b []float64) float64 {
	return embedding.CosineSimilarity(a, b)
}

func (m *mockProvider) GenerateEmbedding(ctx context.Context, text string) ([]float64, error) {
	if text == "" {
		return nil, nil
	}
	return embedding.GenerateDeterministicVector(text, m.dimension), nil
}

func (m *mockProvider) GenerateImageEmbedding(ctx context.Context, base64Image string) ([]float64, error) {
	if base64Image == "" {
		return nil, nil
	}
	return embedding.GenerateDeterministicVector(base64Image, m.dimension), nil
}

func (m *mockProvider) GenerateSkillEmbeddings(
	ctx context.Context,
	name, description, instructions string,
	triggers []string,
	references map[string]string,
	examples map[string]string,
) ([]model.SkillEmbeddingChunk, error) {
	var chunks []model.SkillEmbeddingChunk
	for _, text := range append([]string{name, description, instructions}, triggers...) {
		if text == "" {
			continue
		}
		vec, err := m.GenerateEmbedding(ctx, text)
		if err != nil {
			return nil, err
		}
		chunks = append(chunks, model.SkillEmbeddingChunk{
			TargetType: "skill",
			TargetName: "SKILL.md",
			Vector:     vec,
			ModelName:  m.name,
		})
	}
	return chunks, nil
}

func TestCosineSimilarity(t *testing.T) {
	t.Parallel()

	// Identical vectors -> 1.0
	v1 := []float64{1.0, 0.0, 0.0}
	v2 := []float64{1.0, 0.0, 0.0}
	assert.InDelta(t, 1.0, embedding.CosineSimilarity(v1, v2), 0.0001)

	// Orthogonal vectors -> 0.0
	v3 := []float64{0.0, 1.0, 0.0}
	assert.InDelta(t, 0.0, embedding.CosineSimilarity(v1, v3), 0.0001)

	// Opposite vectors -> -1.0
	v4 := []float64{-1.0, 0.0, 0.0}
	assert.InDelta(t, -1.0, embedding.CosineSimilarity(v1, v4), 0.0001)

	// Empty vectors -> 0.0
	assert.Equal(t, 0.0, embedding.CosineSimilarity(nil, v1))
	assert.Equal(t, 0.0, embedding.CosineSimilarity(v1, nil))
	assert.Equal(t, 0.0, embedding.CosineSimilarity([]float64{1.0}, []float64{1.0, 2.0}))
}

func TestSplitTextIntoChunks(t *testing.T) {
	t.Parallel()

	// Empty text
	assert.Nil(t, embedding.SplitTextIntoChunks("", 100))
	assert.Nil(t, embedding.SplitTextIntoChunks("   ", 100))

	// Short text within limit
	short := "This is a short text block."
	chunks := embedding.SplitTextIntoChunks(short, 100)
	require.Len(t, chunks, 1)
	assert.Equal(t, short, chunks[0])

	// Multi-paragraph text
	para := "First paragraph.\n\nSecond paragraph.\n\nThird paragraph."
	pChunks := embedding.SplitTextIntoChunks(para, 30)
	assert.GreaterOrEqual(t, len(pChunks), 2)
}

func TestGenerateDeterministicVector(t *testing.T) {
	t.Parallel()

	vec768 := embedding.GenerateDeterministicVector("hello world test", 768)
	assert.Len(t, vec768, 768)

	vec1408 := embedding.GenerateDeterministicVector("hello world test", 1408)
	assert.Len(t, vec1408, 1408)

	// Vector should be unit-normalized (length ≈ 1.0)
	var sumSquares float64
	for _, v := range vec768 {
		sumSquares += v * v
	}
	assert.InDelta(t, 1.0, sumSquares, 0.001)

	// Identical input produces identical vectors
	vec768_2 := embedding.GenerateDeterministicVector("hello world test", 768)
	assert.Equal(t, vec768, vec768_2)
}

func TestMockProviderIntegration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	p := &mockProvider{name: "mock-768", dimension: 768}
	assert.Equal(t, "mock-768", p.Name())
	assert.Equal(t, 768, p.Dimension())

	vec, err := p.GenerateEmbedding(ctx, "sample text")
	require.NoError(t, err)
	assert.Len(t, vec, 768)

	imgVec, err := p.GenerateImageEmbedding(ctx, "base64sample")
	require.NoError(t, err)
	assert.Len(t, imgVec, 768)

	chunks, err := p.GenerateSkillEmbeddings(ctx, "skill-a", "desc", "instructions", []string{"trig1"}, nil, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, chunks)
}
