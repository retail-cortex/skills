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

	"github.com/retail-cortex/skills/pkg/model"
)

func runLogin(args []string, stdout, stderr io.Writer) int {
	var appName = "skm-cli"
	var email string
	var domain string
	var orgID string
	var serverURL string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-a" || arg == "--app" || arg == "--app-name":
			if i+1 < len(args) {
				appName = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--app="):
			appName = strings.TrimPrefix(arg, "--app=")
		case arg == "-e" || arg == "--email":
			if i+1 < len(args) {
				email = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--email="):
			email = strings.TrimPrefix(arg, "--email=")
		case arg == "-d" || arg == "--domain":
			if i+1 < len(args) {
				domain = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--domain="):
			domain = strings.TrimPrefix(arg, "--domain=")
		case arg == "-o" || arg == "--org" || arg == "--organization":
			if i+1 < len(args) {
				orgID = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--org="):
			orgID = strings.TrimPrefix(arg, "--org=")
		case arg == "-s" || arg == "--server":
			if i+1 < len(args) {
				serverURL = args[i+1]
				i++
			}
		case strings.HasPrefix(arg, "--server="):
			serverURL = strings.TrimPrefix(arg, "--server=")
		}
	}

	cfg, _ := LoadCLIConfig()
	if serverURL != "" {
		cfg.ServerURL = serverURL
	}

	if email == "" {
		if hostname, err := os.Hostname(); err == nil && hostname != "" {
			appName = fmt.Sprintf("skm-cli-%s", hostname)
		}
		fmt.Fprintf(stderr, "Error: email address required for registration (e.g. skm login --email dev@company.com [--domain company.com])\n")
		return 1
	}

	reqPayload := model.AppRegisterRequest{
		AppName:        appName,
		Email:          email,
		Domain:         domain,
		OrganizationID: orgID,
	}

	payloadBytes, err := json.Marshal(reqPayload)
	if err != nil {
		fmt.Fprintf(stderr, "Error serializing registration request: %v\n", err)
		return 1
	}

	targetURL := strings.TrimRight(cfg.ServerURL, "/") + "/api/v1/apps/register"
	client := &http.Client{Timeout: 15 * time.Second}

	req, err := http.NewRequest("POST", targetURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Fprintf(stderr, "Error creating HTTP request: %v\n", err)
		return 1
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(stderr, "Error connecting to SKM server %s: %v\n", targetURL, err)
		return 1
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		fmt.Fprintf(stderr, "\n[!] Application Registration Failed (%d)\n", resp.StatusCode)
		fmt.Fprintf(stderr, "    %s\n\n", string(respBody))
		return 1
	}

	var appResp model.AppRegisterResponse
	if err := json.Unmarshal(respBody, &appResp); err != nil {
		fmt.Fprintf(stderr, "Error parsing server response: %v\n", err)
		return 1
	}

	cfg.APIKey = appResp.APIKey
	cfg.Domain = appResp.Domain
	cfg.OrganizationID = appResp.OrganizationID

	if err := SaveCLIConfig(cfg); err != nil {
		fmt.Fprintf(stderr, "Warning: Failed to save credentials to %s: %v\n", GetConfigPath(), err)
	}

	fmt.Fprintf(stdout, "\n[+] Successfully Registered & Logged In\n")
	fmt.Fprintf(stdout, "%s\n", strings.Repeat("=", 65))
	fmt.Fprintf(stdout, "App Name:          %s\n", appResp.AppName)
	fmt.Fprintf(stdout, "App ID:            %s\n", appResp.AppID)
	fmt.Fprintf(stdout, "App URN:           %s\n", appResp.AppURN)
	fmt.Fprintf(stdout, "Domain:            %s\n", appResp.Domain)
	fmt.Fprintf(stdout, "Domain Status:     %s\n", appResp.DomainVerificationStatus)
	if appResp.DNSTXTChallenge != "" {
		fmt.Fprintf(stdout, "DNS TXT Challenge:  %s\n", appResp.DNSTXTChallenge)
	}
	fmt.Fprintf(stdout, "Server:            %s\n", cfg.ServerURL)
	fmt.Fprintf(stdout, "Config File:       %s\n\n", GetConfigPath())

	fmt.Fprintf(stdout, "[!] Action Required: Activate your account via email URL:\n")
	fmt.Fprintf(stdout, "    %s\n\n", appResp.VerificationURL)

	return 0
}
