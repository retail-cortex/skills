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

package model_test

import (
	"testing"

	"github.com/retail-cortex/castor/pkg/model"
	"github.com/stretchr/testify/assert"
)

func TestPaginationParams_DefaultsAndBounding(t *testing.T) {
	tests := []struct {
		name         string
		input        model.PaginationParams
		expectedPage int
		expectedSize int
		expectedOff  int
		expectedLim  int
	}{
		{
			name:         "default settings",
			input:        model.DefaultPagination(),
			expectedPage: 1,
			expectedSize: 5,
			expectedOff:  0,
			expectedLim:  5,
		},
		{
			name:         "zero values normalized",
			input:        model.PaginationParams{Page: 0, PageSize: 0},
			expectedPage: 1,
			expectedSize: 5,
			expectedOff:  0,
			expectedLim:  5,
		},
		{
			name:         "negative values normalized",
			input:        model.PaginationParams{Page: -5, PageSize: -10},
			expectedPage: 1,
			expectedSize: 5,
			expectedOff:  0,
			expectedLim:  5,
		},
		{
			name:         "page size capped at 25",
			input:        model.PaginationParams{Page: 2, PageSize: 100},
			expectedPage: 2,
			expectedSize: 25,
			expectedOff:  25,
			expectedLim:  25,
		},
		{
			name:         "page 3 with size 10",
			input:        model.PaginationParams{Page: 3, PageSize: 10},
			expectedPage: 3,
			expectedSize: 10,
			expectedOff:  20,
			expectedLim:  10,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p := tc.input
			assert.Equal(t, tc.expectedOff, p.Offset())
			assert.Equal(t, tc.expectedLim, p.Limit())

			p.Normalize()
			assert.Equal(t, tc.expectedPage, p.Page)
			assert.Equal(t, tc.expectedSize, p.PageSize)
		})
	}
}
