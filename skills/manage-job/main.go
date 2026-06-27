package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
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

func loadScriptURL() string {
	configPath := filepath.Join(repoRoot(), "config", "sheets-deployment.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot read config/sheets-deployment.yaml: %v\n", err)
		os.Exit(1)
	}
	// Parse simple YAML: "deploymentId: <value>"
	var deploymentID string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "deploymentId:") {
			deploymentID = strings.TrimSpace(strings.TrimPrefix(line, "deploymentId:"))
			break
		}
	}
	if deploymentID == "" {
		fmt.Fprintf(os.Stderr, "Error: deploymentId not found in config/sheets-deployment.yaml\n")
		os.Exit(1)
	}
	return fmt.Sprintf("https://script.google.com/macros/s/%s/exec", deploymentID)
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

func cmdGet(args []string) int {
	scriptURL := loadScriptURL()

	if len(args) > 0 {
		params := url.Values{}
		for i := 0; i+1 < len(args); i += 2 {
			params.Set(args[i], args[i+1])
		}
		if len(params) > 0 {
			scriptURL = scriptURL + "?" + params.Encode()
		}
	}

	resp, err := http.Get(scriptURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
		return 1
	}

	fmt.Print(string(body))
	return 0
}

func validateURL(s string) error {
	matched, _ := regexp.MatchString(`^https?://`, s)
	if !matched {
		return fmt.Errorf("link must be a valid URL starting with http:// or https://")
	}
	return nil
}

func validateEmail(s string) error {
	if !strings.Contains(s, "@") || !strings.Contains(s, ".") {
		return fmt.Errorf("email must be a valid email address")
	}
	return nil
}

func validatePhone(s string) error {
	digits := regexp.MustCompile(`[^\d]`).ReplaceAllString(s, "")
	if len(digits) < 10 || len(digits) > 15 {
		return fmt.Errorf("phone number must be 10-15 digits")
	}
	return nil
}

func cmdTrack(args []string) int {
	if len(args) < 4 {
		fmt.Fprintln(os.Stderr, "Usage: manage-job track <companyName> <link> <industry> <status> [email] [phone] [notes]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Required:")
		fmt.Fprintln(os.Stderr, "  companyName — Name of the company")
		fmt.Fprintln(os.Stderr, "  link        — Job posting URL")
		fmt.Fprintln(os.Stderr, "  industry    — Tech, Health Care, Retail, Finance, Gig, Other")
		fmt.Fprintln(os.Stderr, "  status      — Not Started, Applied Only, Applied + Emailed, Applied + Called, Applied + Emailed + Called, Interview!, Done")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Optional:")
		fmt.Fprintln(os.Stderr, "  email       — Employer contact email")
		fmt.Fprintln(os.Stderr, "  phone       — Contact phone number")
		fmt.Fprintln(os.Stderr, "  notes       — Free-form notes")
		return 1
	}

	companyName := args[0]
	link := args[1]
	industry := args[2]
	status := args[3]
	email := ""
	phone := ""
	notes := ""

	if len(args) > 4 {
		email = args[4]
	}
	if len(args) > 5 {
		phone = args[5]
	}
	if len(args) > 6 {
		notes = strings.Join(args[6:], " ")
	}

	if err := validateURL(link); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	if !validIndustries[industry] {
		fmt.Fprintf(os.Stderr, "Error: industry must be one of: Tech, Health Care, Retail, Finance, Gig, Other\n")
		return 1
	}
	if !validStatuses[status] {
		fmt.Fprintf(os.Stderr, "Error: status must be one of: Not Started, Applied Only, Applied + Emailed, Applied + Called, Applied + Emailed + Called, Interview!, Done\n")
		return 1
	}
	if email != "" {
		if err := validateEmail(email); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
	}
	if phone != "" {
		if err := validatePhone(phone); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return 1
		}
	}

	today := time.Now().Format("2006-01-02")

	payload := map[string]interface{}{
		"action":      "create",
		"companyName": companyName,
		"link":        link,
		"dateApplied": today,
		"industry":    industry,
		"status":      status,
	}
	if email != "" {
		payload["email"] = email
	}
	if phone != "" {
		payload["phoneNumber"] = phone
	}
	if notes != "" {
		payload["notes"] = notes
	}

	body, err := json.Marshal(payload)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding payload: %v\n", err)
		return 1
	}

	scriptURL := loadScriptURL()
	result, err := postFollowRedirect(scriptURL, body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}

	if strings.Contains(result, `"error"`) {
		fmt.Fprintf(os.Stderr, "Fail: %s\n", result)
		return 1
	}

	fmt.Printf("Success: %s\n", result)
	return 0
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Usage: manage-job <get|track> [args...]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "Commands:")
		fmt.Fprintln(os.Stderr, "  get    — Retrieve all tracked job applications (JSON to stdout)")
		fmt.Fprintln(os.Stderr, "  track  — Record a new job application")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "get":
		os.Exit(cmdGet(os.Args[2:]))
	case "track":
		os.Exit(cmdTrack(os.Args[2:]))
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		fmt.Fprintln(os.Stderr, "Usage: manage-job <get|track> [args...]")
		os.Exit(1)
	}
}
