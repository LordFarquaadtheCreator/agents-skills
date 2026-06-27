package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

var ValidIndustries = map[string]bool{
	"Tech": true, "Health Care": true, "Retail": true,
	"Finance": true, "Gig": true, "Other": true,
}

var ValidStatuses = map[string]bool{
	"Not Started": true, "Applied Only": true, "Applied + Emailed": true,
	"Applied + Called": true, "Applied + Emailed + Called": true,
	"Interview!": true, "Done": true,
}

type sheetsConfig struct {
	DeploymentID string `yaml:"deploymentId"`
}

func repoRoot() string {
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot resolve executable path: %v\n", err)
		os.Exit(1)
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(exe)))
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

// AppScript encapsulates communication with the deployed Apps Script web app.
// Commands call methods on this struct — they don't deal with URLs,
// redirects, or payload construction directly.
type AppScript struct {
	url string
}

func NewAppScript() *AppScript {
	return &AppScript{url: loadScriptURL()}
}

// postFollowRedirect handles Apps Script's 302 redirect on POST.
// Go's default client converts POST to GET on redirect, dropping the body.
// We capture the Location header, then GET it to retrieve the actual response.
func (a *AppScript) postFollowRedirect(body []byte) (string, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("POST", a.url, bytes.NewReader(body))
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

func (a *AppScript) Get(params url.Values) (string, error) {
	target := a.url
	if len(params) > 0 {
		target = target + "?" + params.Encode()
	}

	resp, err := http.Get(target)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (a *AppScript) Create(entry map[string]interface{}) (string, error) {
	entry["action"] = "create"
	body, err := json.Marshal(entry)
	if err != nil {
		return "", err
	}
	return a.postFollowRedirect(body)
}

func (a *AppScript) Patch(matchBy, update map[string]interface{}) (string, error) {
	payload := map[string]interface{}{
		"action":  "patch",
		"matchBy": matchBy,
		"update":  update,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return a.postFollowRedirect(body)
}

func (a *AppScript) Delete(matchBy map[string]interface{}) (string, error) {
	payload := map[string]interface{}{
		"action":  "delete",
		"matchBy": matchBy,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return a.postFollowRedirect(body)
}
