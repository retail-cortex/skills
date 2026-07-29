package skillsloader

import (
	"sort"
)

// AuthorDetails represents structured attribution for a skill contributor.
type AuthorDetails struct {
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
	URL   string `json:"url,omitempty"`
}

// ToolRequirement represents a structured permission contract for agent tool execution.
type ToolRequirement struct {
	Name        string   `json:"name"`
	Scopes      []string `json:"scopes,omitempty"`
	Description string   `json:"description,omitempty"`
}

// ExecutionHints provides operational parameters for LLM agents, routers, and orchestrators.
type ExecutionHints struct {
	PreferredModel        string            `json:"preferred_model,omitempty"`
	RequiresHumanApproval bool              `json:"requires_human_approval,omitempty"`
	EnvironmentVariables  []string          `json:"environment_variables,omitempty"`
	TimeoutSeconds        int               `json:"timeout_seconds,omitempty"`
	CustomHints           map[string]string `json:"custom_hints,omitempty"`
}

// SkillDefinition represents a loaded enterprise skill definition.
type SkillDefinition struct {
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	Instructions     string            `json:"instructions"`
	License          string            `json:"license,omitempty"`
	Author           string            `json:"author,omitempty"`
	Authors          []AuthorDetails   `json:"authors,omitempty"`
	Version          string            `json:"version,omitempty"`
	Compatibility    string            `json:"compatibility,omitempty"`
	AllowedTools     string            `json:"allowed_tools,omitempty"`
	ToolRequirements []ToolRequirement `json:"tool_requirements,omitempty"`
	Category         string            `json:"category,omitempty"`
	Tags             []string          `json:"tags,omitempty"`
	TriggerPhrases   []string          `json:"trigger_phrases,omitempty"`
	ExecutionHints   *ExecutionHints   `json:"execution_hints,omitempty"`
	Metadata         map[string]string `json:"metadata"`
	References       map[string]string `json:"references"`
	Examples         map[string]string `json:"examples"`
	Path             string            `json:"path"`
}

// GetReferenceContent retrieves content of a reference file on demand.
func (s *SkillDefinition) GetReferenceContent(refName string) string {
	if s.References != nil {
		return s.References[refName]
	}
	return ""
}

// GetExampleContent retrieves content of an example file on demand.
func (s *SkillDefinition) GetExampleContent(exName string) string {
	if s.Examples != nil {
		return s.Examples[exName]
	}
	return ""
}

// ToMap serializes skill definition to a dictionary format matching the Python implementation.
func (s *SkillDefinition) ToMap() map[string]any {
	refKeys := make([]string, 0, len(s.References))
	for k := range s.References {
		refKeys = append(refKeys, k)
	}
	sort.Strings(refKeys)

	exKeys := make([]string, 0, len(s.Examples))
	for k := range s.Examples {
		exKeys = append(exKeys, k)
	}
	sort.Strings(exKeys)

	meta := s.Metadata
	if meta == nil {
		meta = make(map[string]string)
	}

	return map[string]any{
		"name":              s.Name,
		"description":       s.Description,
		"instructions":      s.Instructions,
		"license":           s.License,
		"author":            s.Author,
		"authors":           s.Authors,
		"version":           s.Version,
		"compatibility":     s.Compatibility,
		"allowed_tools":     s.AllowedTools,
		"tool_requirements": s.ToolRequirements,
		"category":          s.Category,
		"tags":              s.Tags,
		"trigger_phrases":   s.TriggerPhrases,
		"execution_hints":   s.ExecutionHints,
		"metadata":          meta,
		"references":        refKeys,
		"examples":          exKeys,
		"path":              s.Path,
	}
}

// SkillSummary represents high-level summary metadata for a registered skill.
type SkillSummary struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	ReferenceCount int      `json:"reference_count"`
	ExampleCount   int      `json:"example_count"`
	Path           string   `json:"path"`
	Category       string   `json:"category,omitempty"`
	Tags           []string `json:"tags,omitempty"`
	TriggerPhrases []string `json:"trigger_phrases,omitempty"`
}
