package data_test

import (
	"testing"

	"github.com/retail-cortex/skills/pkg/data"
	"github.com/retail-cortex/skills/pkg/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) {
	data.ResetEngine()
	db, err := data.InitDB("file::memory:?cache=shared")
	require.NoError(t, err)
	require.NotNil(t, db)
}

func TestAppsRepository(t *testing.T) {
	setupTestDB(t)
	db := data.GetDB()
	repo := data.NewAppsRepository()

	// 1. Register App
	req := model.AppRegisterRequest{
		AppName: "test-app",
		Email:   "test@example.com",
	}
	res, err := repo.RegisterApp(db, req, "http://localhost:8000")
	require.NoError(t, err)
	assert.NotEmpty(t, res.AppID)
	assert.NotEmpty(t, res.APIKey)
	assert.NotEmpty(t, res.VerificationToken)
	assert.Contains(t, res.VerificationURL, res.VerificationToken)

	// Duplicate registration attempt
	_, err = repo.RegisterApp(db, req, "http://localhost:8000")
	assert.ErrorIs(t, err, data.ErrAppAlreadyRegistered)

	// 2. Authenticate unverified app should fail
	_, err = repo.AuthenticateAPIKey(db, res.APIKey)
	assert.ErrorIs(t, err, data.ErrAppNotVerified)

	// Missing API Key
	_, err = repo.AuthenticateAPIKey(db, "")
	assert.ErrorIs(t, err, data.ErrMissingAPIKey)

	// Invalid API Key
	_, err = repo.AuthenticateAPIKey(db, "sk_live_invalid")
	assert.ErrorIs(t, err, data.ErrInvalidAPIKey)

	// 3. Verify App
	verifyRes, err := repo.VerifyApp(db, res.VerificationToken)
	require.NoError(t, err)
	assert.True(t, verifyRes.IsActive)

	// 4. Authenticate verified app
	app, err := repo.AuthenticateAPIKey(db, res.APIKey)
	require.NoError(t, err)
	assert.Equal(t, res.AppID, app.AppID)
	assert.True(t, app.IsActive)

	// 5. Invalid token verify
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
	skillRes, err := skillsRepo.CreateSkill(db, appRes.AppID, createReq, []float64{0.1, 0.2, 0.3}, "text-embedding-004")
	require.NoError(t, err)
	assert.Equal(t, "test-skill", skillRes.Name)
	assert.Equal(t, appRes.AppID, skillRes.AppID)
	assert.Equal(t, "1.0.0", skillRes.Version)
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

	// Non-existent skill get -> error
	_, err = skillsRepo.GetSkill(db, "non-existent-skill")
	assert.ErrorIs(t, err, data.ErrSkillNotFound)

	// 3. List Skills
	skills, err := skillsRepo.ListSkills(db, "", nil)
	require.NoError(t, err)
	assert.Len(t, skills, 1)

	// Keyword search
	matched, err := skillsRepo.ListSkills(db, "testing", nil)
	require.NoError(t, err)
	assert.Len(t, matched, 1)

	// Vector search
	vectorMatches, err := skillsRepo.ListSkills(db, "testing", []float64{0.1, 0.2, 0.3})
	require.NoError(t, err)
	assert.Len(t, vectorMatches, 1)
	assert.NotNil(t, vectorMatches[0].SimilarityScore)

	// 4. Update Skill (partial)
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
	updated, err := skillsRepo.UpdateSkill(db, skillRes.ID, appRes.AppID, updateReq, false, []float64{0.4, 0.5, 0.6}, "text-embedding-004")
	require.NoError(t, err)
	assert.Equal(t, "Updated description", updated.Description)
	assert.Equal(t, "1.1.0", updated.Version)
	assert.Equal(t, "newval", updated.Metadata["newkey"])

	// Update Skill (full replace)
	fullReplaceDesc := "Full replace desc"
	fullReq := model.SkillUpdateRequest{
		Description: &fullReplaceDesc,
	}
	fullUpdated, err := skillsRepo.UpdateSkill(db, skillRes.ID, appRes.AppID, fullReq, true, nil, "")
	require.NoError(t, err)
	assert.Equal(t, "Full replace desc", fullUpdated.Description)

	// Update non-existent skill
	_, err = skillsRepo.UpdateSkill(db, "non-existent", appRes.AppID, updateReq, false, nil, "")
	assert.ErrorIs(t, err, data.ErrSkillNotFound)

	// Update with wrong app ID
	_, err = skillsRepo.UpdateSkill(db, skillRes.ID, "wrong-app", updateReq, false, nil, "")
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
