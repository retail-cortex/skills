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

package data_test

import (
	"path/filepath"
	"testing"

	"github.com/retail-cortex/skills/pkg/data"
	"github.com/retail-cortex/skills/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) {
	data.ResetEngine()
	db, err := data.InitDB(filepath.Join(t.TempDir(), "test.db"))
	require.NoError(t, err)
	require.NotNil(t, db)
}

func TestAppsRepository(t *testing.T) {
	setupTestDB(t)
	db := data.GetDB()
	repo := data.NewAppsRepository()

	// 1. Register App with matching domain (SSO Verification)
	req := model.AppRegisterRequest{
		AppName: "test-app",
		Domain:  "example.com",
		Email:   "test@example.com",
	}
	res, err := repo.RegisterApp(db, req, "http://localhost:8000")
	require.NoError(t, err)
	assert.NotEmpty(t, res.AppID)
	assert.Equal(t, "example.com", res.Domain)
	assert.Equal(t, "urn:skm:app:example.com:test-app", res.AppURN)
	assert.Equal(t, model.DomainStatusVerifiedSSO, res.DomainVerificationStatus)
	assert.NotEmpty(t, res.APIKey)
	assert.NotEmpty(t, res.VerificationToken)
	assert.Contains(t, res.VerificationURL, res.VerificationToken)

	// 2. Register App with custom non-matching domain (DNS Challenge)
	reqCustom := model.AppRegisterRequest{
		AppName: "custom-app",
		Domain:  "customdomain.org",
		Email:   "dev@corp.com",
	}
	resCustom, err := repo.RegisterApp(db, reqCustom, "http://localhost:8000")
	require.NoError(t, err)
	assert.Equal(t, "customdomain.org", resCustom.Domain)
	assert.Equal(t, "urn:skm:app:customdomain.org:custom-app", resCustom.AppURN)
	assert.Equal(t, model.DomainStatusPendingDNS, resCustom.DomainVerificationStatus)
	assert.Contains(t, resCustom.DNSTXTChallenge, "skm-domain-verify-")

	// 3. Prohibit freemail claiming enterprise domain
	reqFreemail := model.AppRegisterRequest{
		AppName: "spoof-app",
		Domain:  "google.com",
		Email:   "hacker@gmail.com",
	}
	_, err = repo.RegisterApp(db, reqFreemail, "http://localhost:8000")
	assert.ErrorIs(t, err, data.ErrFreemailDomainProhibited)

	// Duplicate registration attempt
	_, err = repo.RegisterApp(db, req, "http://localhost:8000")
	assert.ErrorIs(t, err, data.ErrAppAlreadyRegistered)

	// 4. Authenticate unverified app should fail
	_, err = repo.AuthenticateAPIKey(db, res.APIKey)
	assert.ErrorIs(t, err, data.ErrAppNotVerified)

	// Missing API Key
	_, err = repo.AuthenticateAPIKey(db, "")
	assert.ErrorIs(t, err, data.ErrMissingAPIKey)

	// 4. Invalid API Key
	_, err = repo.AuthenticateAPIKey(db, "skm_live_invalid")
	assert.ErrorIs(t, err, data.ErrInvalidAPIKey)

	// 5. Verify App
	verifyRes, err := repo.VerifyApp(db, res.VerificationToken)
	require.NoError(t, err)
	assert.True(t, verifyRes.IsActive)
	assert.Equal(t, "example.com", verifyRes.Domain)
	assert.Equal(t, "urn:skm:app:example.com:test-app", verifyRes.AppURN)

	// 6. Authenticate verified app
	app, err := repo.AuthenticateAPIKey(db, res.APIKey)
	require.NoError(t, err)
	assert.Equal(t, res.AppID, app.AppID)
	assert.True(t, app.IsActive)

	// 7. Invalid token verify
	_, err = repo.VerifyApp(db, "invalid-token")
	assert.ErrorIs(t, err, data.ErrInvalidVerificationToken)
}

func TestSkillsRepository(t *testing.T) {
	setupTestDB(t)
	db := data.GetDB()
	appsRepo := data.NewAppsRepository()
	skillsRepo := data.NewSkillsRepository()

	// Register and verify app
	appRes, err := appsRepo.RegisterApp(db, model.AppRegisterRequest{
		AppName: "skills-app",
		Email:   "skills@example.com",
	}, "http://localhost:8000")
	require.NoError(t, err)
	_, err = appsRepo.VerifyApp(db, appRes.VerificationToken)
	require.NoError(t, err)

	category := "testing"
	createReq := model.SkillCreateRequest{
		Name:         "test-skill",
		Description:  "Skill for testing",
		Instructions: "Run unit tests carefully",
		Category:     &category,
		Tags:         []string{"test", "go"},
		Metadata:     map[string]string{"key": "value"},
		References:   map[string]string{"ref1": "content1"},
		Examples:     map[string]string{"ex1": "example1"},
	}

	// 1. Create Skill
	testChunks := []model.SkillEmbeddingChunk{
		{TargetType: "skill", TargetName: "SKILL.md", Vector: []float64{0.1, 0.2, 0.3}, ModelName: "multimodalembedding"},
		{TargetType: "reference", TargetName: "ref1", Vector: []float64{0.4, 0.5, 0.6}, ModelName: "multimodalembedding"},
	}
	skillRes, err := skillsRepo.CreateSkill(db, appRes.AppID, createReq, testChunks)
	require.NoError(t, err)
	assert.Equal(t, "test-skill", skillRes.Name)
	assert.Equal(t, appRes.AppID, skillRes.AppID)
	assert.Equal(t, "1.0.0", skillRes.Version)
	assert.Equal(t, "skm://skills/example.com/testing/test-skill/1.0.0", skillRes.URI)
	assert.Equal(t, "value", skillRes.Metadata["key"])
	assert.Equal(t, "content1", skillRes.References["ref1"])
	assert.Equal(t, "example1", skillRes.Examples["ex1"])

	// 2. Get Skill
	fetched, err := skillsRepo.GetSkill(db, skillRes.ID)
	require.NoError(t, err)
	assert.Equal(t, skillRes.Name, fetched.Name)

	fetchedByName, err := skillsRepo.GetSkill(db, "test-skill")
	require.NoError(t, err)
	assert.Equal(t, skillRes.ID, fetchedByName.ID)

	fetchedByURI, err := skillsRepo.GetSkill(db, "skm://skills/example.com/testing/test-skill/1.0.0")
	require.NoError(t, err)
	assert.Equal(t, skillRes.ID, fetchedByURI.ID)

	// Non-existent skill get -> error
	_, err = skillsRepo.GetSkill(db, "non-existent-skill")
	assert.ErrorIs(t, err, data.ErrSkillNotFound)

	// 3. Idempotent Upsert (New Version 1.1.0 on same skill)
	ver11 := "1.1.0"
	createReq.Version = &ver11
	skillResV11, err := skillsRepo.CreateSkill(db, appRes.AppID, createReq, nil)
	require.NoError(t, err)
	assert.Equal(t, skillRes.ID, skillResV11.ID) // Same root Skill ID
	assert.Equal(t, "1.1.0", skillResV11.Version)
	assert.Equal(t, "skm://skills/example.com/testing/test-skill/1.1.0", skillResV11.URI)

	// 4. List Skills (should still only be 1 root skill)
	list, err := skillsRepo.ListSkills(db, "", nil)
	require.NoError(t, err)
	assert.Equal(t, int64(1), list.TotalCount)
	assert.Len(t, list.Items, 1)

	// List Skills by semantic query
	listSem, err := skillsRepo.ListSkills(db, "unit test", []float64{0.1, 0.2, 0.3})
	require.NoError(t, err)
	assert.Equal(t, int64(1), listSem.TotalCount)
	assert.Len(t, listSem.Items, 1)
	assert.NotNil(t, listSem.Items[0].SimilarityScore)
	assert.NotNil(t, listSem.Items[0].MatchingChunk)
	assert.Equal(t, "SKILL.md", *listSem.Items[0].MatchingChunk)

	// Update Skill
	newDesc := "Updated description"
	newVer := "1.1.0"
	newMeta := map[string]string{"newkey": "newval"}
	newRefs := map[string]string{"newref": "newrefcontent"}
	newExs := map[string]string{"newex": "newexcontent"}
	updateReq := model.SkillUpdateRequest{
		Description: &newDesc,
		Version:     &newVer,
		Metadata:    &newMeta,
		References:  &newRefs,
		Examples:    &newExs,
	}
	updChunks := []model.SkillEmbeddingChunk{
		{TargetType: "skill", TargetName: "SKILL.md", Vector: []float64{0.4, 0.5, 0.6}, ModelName: "multimodalembedding"},
	}
	updated, err := skillsRepo.UpdateSkill(db, skillRes.ID, appRes.AppID, updateReq, false, updChunks)
	require.NoError(t, err)
	assert.Equal(t, "Updated description", updated.Description)
	assert.Equal(t, "1.1.0", updated.Version)
	assert.Equal(t, "newval", updated.Metadata["newkey"])

	// Update Skill (full replace)
	fullReplaceDesc := "Full replace desc"
	fullReq := model.SkillUpdateRequest{
		Description: &fullReplaceDesc,
	}
	fullUpdated, err := skillsRepo.UpdateSkill(db, skillRes.ID, appRes.AppID, fullReq, true, nil)
	require.NoError(t, err)
	assert.Equal(t, "Full replace desc", fullUpdated.Description)

	// Update non-existent skill
	_, err = skillsRepo.UpdateSkill(db, "non-existent", appRes.AppID, updateReq, false, nil)
	assert.ErrorIs(t, err, data.ErrSkillNotFound)

	// Update with wrong app ID
	_, err = skillsRepo.UpdateSkill(db, skillRes.ID, "wrong-app", updateReq, false, nil)
	assert.ErrorIs(t, err, data.ErrUnauthorizedSkillAccess)

	// 5. Delete Skill with wrong app ID
	_, err = skillsRepo.DeleteSkill(db, skillRes.ID, "wrong-app-id")
	assert.ErrorIs(t, err, data.ErrUnauthorizedSkillAccess)

	// 6. Delete Skill
	deleteRes, err := skillsRepo.DeleteSkill(db, skillRes.ID, appRes.AppID)
	require.NoError(t, err)
	assert.Equal(t, "success", deleteRes["status"])

	// Verify deletion
	_, err = skillsRepo.GetSkill(db, skillRes.ID)
	assert.ErrorIs(t, err, data.ErrSkillNotFound)

	// Delete non-existent skill
	_, err = skillsRepo.DeleteSkill(db, skillRes.ID, appRes.AppID)
	assert.ErrorIs(t, err, data.ErrSkillNotFound)
}

func TestCosineSimilarity(t *testing.T) {
	v1 := []float64{1.0, 0.0, 0.0}
	v2 := []float64{1.0, 0.0, 0.0}
	sim := data.CosineSimilarity(v1, v2)
	assert.InDelta(t, 1.0, sim, 0.0001)

	v3 := []float64{0.0, 1.0, 0.0}
	simOrthogonal := data.CosineSimilarity(v1, v3)
	assert.InDelta(t, 0.0, simOrthogonal, 0.0001)

	assert.Equal(t, 0.0, data.CosineSimilarity(nil, nil))
	assert.Equal(t, 0.0, data.CosineSimilarity([]float64{1.0}, []float64{1.0, 2.0}))
}
