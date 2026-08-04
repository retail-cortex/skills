package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/retail-cortex/skills/internal/installer"
	"github.com/retail-cortex/skills/pkg/model"
)

func runRegister(args []string, stdout, stderr io.Writer) int {
	var sourceURI string
	var force = false

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-f" || arg == "--force":
			force = true
		default:
			if !strings.HasPrefix(arg, "-") && sourceURI == "" {
				sourceURI = arg
			}
		}
	}

	if sourceURI == "" {
		fmt.Fprintf(stderr, "Error: source URI required (e.g. skm register github://google/skills@main/tree/main/skills/cloud/gemini-api)\n")
		return 1
	}

	cfg, err := LoadCLIConfig()
	if err != nil || cfg.ServerURL == "" {
		fmt.Fprintf(stderr, "Error: SKM server URL not configured. Run 'skm config set server <URL>' first.\n")
		return 1
	}

	// Resolve local or remote skills at source URI
	skills, err := installer.ResolveFileSkills(sourceURI, nil)
	if err != nil || len(skills) == 0 {
		// Fallback to installer URI resolution
		results, loadErr := installer.AddSkills([]string{sourceURI}, ".tmp_skm_register", nil, force)
		if loadErr != nil || len(results) == 0 {
			fmt.Fprintf(stderr, "Error loading skill from source URI %s: %v\n", sourceURI, loadErr)
			return 1
		}
		_ = os.RemoveAll(".tmp_skm_register")
	}

	// Resolve skill definitions
	resolvedSkills, err := installer.ResolveFileSkills(sourceURI, nil)
	if err != nil || len(resolvedSkills) == 0 {
		fmt.Fprintf(stderr, "Failed to resolve valid skill at %s: %v\n", sourceURI, err)
		return 1
	}

	client := &http.Client{Timeout: 15 * time.Second}
	serverURL := strings.TrimRight(cfg.ServerURL, "/") + "/api/v1/skills"
	registeredCount := 0

	for _, skillDef := range resolvedSkills {
		reqPayload := model.SkillCreateRequest{
			Name:         skillDef.Name,
			SourceURI:    sourceURI,
			Description:  skillDef.Description,
			Instructions: skillDef.Instructions,
			License:      &skillDef.License,
			Author:       &skillDef.Author,
			Version:      &skillDef.Version,
			Metadata:     skillDef.Metadata,
			References:   skillDef.References,
			Examples:     skillDef.Examples,
		}

		payloadBytes, err := json.Marshal(reqPayload)
		if err != nil {
			fmt.Fprintf(stderr, "Error serializing skill request payload for '%s': %v\n", skillDef.Name, err)
			continue
		}

		req, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(payloadBytes))
		if err != nil {
			fmt.Fprintf(stderr, "Error creating HTTP request: %v\n", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		if cfg.APIKey != "" {
			req.Header.Set("X-API-Key", cfg.APIKey)
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Fprintf(stderr, "Error connecting to server %s: %v\n", serverURL, err)
			return 1
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)
		if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
			fmt.Fprintf(stderr, "[!] Server Registration Failed (%d): %s\n", resp.StatusCode, string(respBody))
			return 1
		}

		var skResp model.SkillResponse
		if err := json.Unmarshal(respBody, &skResp); err != nil {
			fmt.Fprintf(stderr, "Error parsing server response: %v\n", err)
			return 1
		}

		registeredCount++
		fmt.Fprintf(stdout, "\n[+] Successfully Registered Skill\n")
		fmt.Fprintf(stdout, "%s\n", strings.Repeat("=", 65))
		fmt.Fprintf(stdout, "Name:       %s\n", skResp.Name)
		fmt.Fprintf(stdout, "Skill ID:   %s\n", skResp.ID)
		fmt.Fprintf(stdout, "Server URI: %s\n", skResp.URI)
		fmt.Fprintf(stdout, "Source URI: %s\n", skResp.SourceURI)
		fmt.Fprintf(stdout, "Server:     %s\n\n", cfg.ServerURL)
	}

	if registeredCount == 0 {
		return 1
	}
	return 0
}
