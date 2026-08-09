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

package model

// PaginationParams represents standard request pagination settings.
type PaginationParams struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

// DefaultPagination returns standard pagination settings: page 1, page_size 5 (max 25).
func DefaultPagination() PaginationParams {
	return PaginationParams{
		Page:     1,
		PageSize: 5,
	}
}

// Normalize ensures page >= 1 and 1 <= page_size <= 25 (default 5).
func (p *PaginationParams) Normalize() {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 5
	}
	if p.PageSize > 25 {
		p.PageSize = 25
	}
}

// Offset returns the zero-based SQL/slice offset.
func (p PaginationParams) Offset() int {
	page := p.Page
	if page < 1 {
		page = 1
	}
	pageSize := p.Limit()
	return (page - 1) * pageSize
}

// Limit returns the bounded limit (1 <= limit <= 25).
func (p PaginationParams) Limit() int {
	if p.PageSize <= 0 {
		return 5
	}
	if p.PageSize > 25 {
		return 25
	}
	return p.PageSize
}

// PaginatedSkillResponse wraps a list of skills with pagination metadata.
type PaginatedSkillResponse struct {
	Items      []*SkillResponse `json:"items"`
	TotalCount int64            `json:"total_count"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}
