package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	err := Run(context.Background())
	assert.NoError(t, err)
}

func TestLoadConfiguration(t *testing.T) {
	cfg, err := LoadConfiguration()
	assert.NoError(t, err)
	assert.Equal(t, "http://localhost:8080", cfg.SKM.ServerURL)
	assert.Equal(t, "example-secret-key-12345", cfg.SKM.APIKey)
}
