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
