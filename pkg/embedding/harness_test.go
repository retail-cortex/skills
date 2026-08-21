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
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/retail-cortex/castor/pkg/embedding"
	"github.com/retail-cortex/castor/pkg/embedding/alloydb"
	"github.com/retail-cortex/castor/pkg/embedding/vertex"
	"github.com/retail-cortex/castor/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type BenchmarkMetric struct {
	ProviderName       string
	Dimension          int
	AvgTextLatencyMS   float64
	P95TextLatencyMS   float64
	SkillEmbeddingMS   float64
	Top1RecallAccuracy float64
	Top3RecallAccuracy float64
	MRR                float64 // Mean Reciprocal Rank
}

type GroundTruthCase struct {
	Query         string
	ExpectedSkill string
}

func TestEmbeddingProvidersHarness(t *testing.T) {
	if os.Getenv("ENV") != "integration" && os.Getenv("MODENV_ENV") != "integration" && os.Getenv("APP_ENV") != "integration" {
		t.Skip("Skipping integration test: set ENV=integration to run against live AlloyDB/Vertex instance")
	}

	ctx := context.Background()

	// Initialize Candidate Providers
	prj := "wmt-lab-prj"
	providers := []embedding.Provider{
		vertex.NewProvider(vertex.Config{ProjectID: prj, Region: "us-central1"}),
		alloydb.NewProvider(alloydb.Config{ModelName: "text-embedding-004", Dimension: 768}),
	}

	// Standard Evaluation Test Corpus
	skillCorpus := []struct {
		Name         string
		Description  string
		Instructions string
		Triggers     []string
		References   map[string]string
		Examples     map[string]string
	}{
		{
			Name:         "configuration-modenv",
			Description:  "Elite hierarchical TOML configuration SDLC. Covers cascading precedence, HTTP 429 backoff, and XOR cipher security.",
			Instructions: "Implement .env.toml hierarchy with runtime staging and production overrides. Decrypt xor: secrets in memory.",
			Triggers:     []string{"load modenv configuration", "cascading TOML setup", "XOR cipher memory decryption"},
			References:   map[string]string{"modenv_spec.md": "Hierarchical configuration specification and XOR cipher security (CWE-798)."},
			Examples:     map[string]string{"main.go": "package main\nimport modenv\nfunc main() { cfg := modenv.Load() }"},
		},
		{
			Name:         "go-project-setup",
			Description:  "Meta-skill for scaffolding enterprise Go microservices using /cmd, /internal, /pkg layout wrapped in Bazel rules_go.",
			Instructions: "Scaffold Go repository layout with table-driven tests, golangci-lint, and distroless Docker packaging.",
			Triggers:     []string{"scaffold go project", "create go microservice", "bazel rules_go configuration"},
			References:   map[string]string{"architecture.md": "Go microservice architecture patterns and import restrictions."},
			Examples:     map[string]string{"service.go": "package internal\ntype Service struct{}"},
		},
		{
			Name:         "multi-content-suite",
			Description:  "Comprehensive multi-content validation skill testing binary encoding across PNG, WebP, PDF, WASM, and SQLite.",
			Instructions: "Verify image canvas rendering, wasm bytecode calculation, sqlite fixtures, and audio waveforms.",
			Triggers:     []string{"test all binary formats", "validate multi-content binary skill", "canvas raster graphics"},
			References:   map[string]string{"canvas.png": "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg=="},
			Examples:     map[string]string{"test_suite.py": "def test_binary_assets(): pass"},
		},
	}

	// Evaluation Queries with Ground Truth
	evalQueries := []GroundTruthCase{
		{Query: "XOR cipher security decryption", ExpectedSkill: "configuration-modenv"},
		{Query: "cascading TOML config settings", ExpectedSkill: "configuration-modenv"},
		{Query: "scaffold go microservice layout Bazel", ExpectedSkill: "go-project-setup"},
		{Query: "golangci-lint table driven testing", ExpectedSkill: "go-project-setup"},
		{Query: "canvas image rendering raster", ExpectedSkill: "multi-content-suite"},
		{Query: "wasm bytecode binary asset", ExpectedSkill: "multi-content-suite"},
	}

	var results []BenchmarkMetric

	fmt.Println("\n=========================================================================================")
	fmt.Println("  SKILL BUILDER EMBEDDING PROVIDER HARNESS: LATENCY & RECALL BENCHMARK")
	fmt.Println("=========================================================================================")

	for _, p := range providers {
		t.Run(p.Name(), func(t *testing.T) {
			assert.NotEmpty(t, p.Name())
			assert.Greater(t, p.Dimension(), 0)

			// 1. Measure Single Text Latencies (10 runs)
			var latencies []time.Duration
			for i := 0; i < 10; i++ {
				start := time.Now()
				_, _ = p.GenerateEmbedding(ctx, "Test microservice architecture with Go and Bazel")
				latencies = append(latencies, time.Since(start))
			}

			var totalDuration time.Duration
			for _, d := range latencies {
				totalDuration += d
			}
			avgTextMS := float64(totalDuration.Milliseconds()) / float64(len(latencies))
			sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
			p95TextMS := float64(latencies[len(latencies)*95/100].Milliseconds())

			// 2. Measure Skill Multi-Chunk Embedding Generation Time
			skillStart := time.Now()
			var skillIndex = make(map[string][]model.SkillEmbeddingChunk)
			for _, s := range skillCorpus {
				chunks, err := p.GenerateSkillEmbeddings(ctx, s.Name, s.Description, s.Instructions, s.Triggers, s.References, s.Examples)
				require.NoError(t, err)
				skillIndex[s.Name] = chunks
			}
			skillEmbeddingMS := float64(time.Since(skillStart).Milliseconds()) / float64(len(skillCorpus))

			// 3. Measure Recall Accuracy & MRR
			top1Hits := 0
			top3Hits := 0
			var reciprocalRankSum float64

			for _, eq := range evalQueries {
				qVec, err := p.GenerateEmbedding(ctx, eq.Query)
				if err != nil || len(qVec) == 0 {
					continue
				}

				type scored struct {
					skillName string
					score     float64
				}
				var scores []scored

				for skillName, chunks := range skillIndex {
					var maxScore float64
					for _, c := range chunks {
						s := p.CosineSimilarity(qVec, c.Vector)
						if s > maxScore {
							maxScore = s
						}
					}
					// Lexical boost bonus
					qTokens := strings.Fields(strings.ToLower(eq.Query))
					for _, tok := range qTokens {
						if len(tok) >= 3 && strings.Contains(strings.ToLower(skillName), tok) {
							maxScore += 0.15
						}
					}
					scores = append(scores, scored{skillName: skillName, score: maxScore})
				}

				sort.Slice(scores, func(i, j int) bool { return scores[i].score > scores[j].score })

				for rank, sc := range scores {
					if sc.skillName == eq.ExpectedSkill {
						if rank == 0 {
							top1Hits++
						}
						if rank < 3 {
							top3Hits++
						}
						reciprocalRankSum += 1.0 / float64(rank+1)
						break
					}
				}
			}

			top1Acc := float64(top1Hits) / float64(len(evalQueries)) * 100.0
			top3Acc := float64(top3Hits) / float64(len(evalQueries)) * 100.0
			mrr := reciprocalRankSum / float64(len(evalQueries))

			metric := BenchmarkMetric{
				ProviderName:       p.Name(),
				Dimension:          p.Dimension(),
				AvgTextLatencyMS:   avgTextMS,
				P95TextLatencyMS:   p95TextMS,
				SkillEmbeddingMS:   skillEmbeddingMS,
				Top1RecallAccuracy: top1Acc,
				Top3RecallAccuracy: top3Acc,
				MRR:                mrr,
			}
			results = append(results, metric)
		})
	}

	// Render Formatted Comparison Report
	fmt.Printf("\n%-18s | %-9s | %-12s | %-12s | %-14s | %-10s | %-6s\n",
		"Provider", "Dimension", "Avg Text Lat", "P95 Text Lat", "Skill Ingest", "Top-1 Rec", "MRR")
	fmt.Println(strings.Repeat("-", 92))
	for _, m := range results {
		fmt.Printf("%-18s | %-9d | %8.1f ms  | %8.1f ms  | %10.1f ms   | %7.1f %%  | %5.2f\n",
			m.ProviderName, m.Dimension, m.AvgTextLatencyMS, m.P95TextLatencyMS, m.SkillEmbeddingMS, m.Top1RecallAccuracy, m.MRR)
	}
	fmt.Println(strings.Repeat("=", 92))
}
