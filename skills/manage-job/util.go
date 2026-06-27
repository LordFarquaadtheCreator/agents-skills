package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

var validIndustries = map[string]bool{
	"Tech": true, "Health Care": true, "Retail": true,
	"Finance": true, "Gig": true, "Other": true,
}

var validStatuses = map[string]bool{
	"Not Started": true, "Applied Only": true, "Applied + Emailed": true,
	"Applied + Called": true, "Applied + Emailed + Called": true,
	"Interview!": true, "Done": true,
}

func repoRoot() string {
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot resolve executable path: %v\n", err)
		os.Exit(1)
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(exe)))
}

type sheetsConfig struct {
	DeploymentID string `yaml:"deploymentId"`
}

func loadScriptURL() string {
	configPath := filepath.Join(repoRoot(), "config", "sheets-deployment.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot read config/sheets-deployment.yaml: %v\n", err)
		os.Exit(1)
	}
	var cfg sheetsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid YAML: %v\n", err)
		os.Exit(1)
	}
	if cfg.DeploymentID == "" {
		fmt.Fprintf(os.Stderr, "Error: deploymentId not found in config/sheets-deployment.yaml\n")
		os.Exit(1)
	}
	return fmt.Sprintf("https://script.google.com/macros/s/%s/exec", cfg.DeploymentID)
}

// postFollowRedirect handles Apps Script's 302 redirect on POST.
// Go's default client converts POST to GET on redirect, dropping the body.
// We capture the Location header, then GET it to retrieve the actual response.
func postFollowRedirect(scriptURL string, body []byte) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("POST", scriptURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 302 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("expected 302, got %d: %s", resp.StatusCode, string(respBody))
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return "", fmt.Errorf("no redirect Location header")
	}

	redirectResp, err := http.Get(location)
	if err != nil {
		return "", err
	}
	defer redirectResp.Body.Close()

	respBody, err := io.ReadAll(redirectResp.Body)
	if err != nil {
		return "", err
	}

	return string(respBody), nil
}
