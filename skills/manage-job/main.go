package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"manage-job/validate"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
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

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Retrieve all tracked job applications",
	Long:  "Fetches all job applications from the spreadsheet. Returns JSON to stdout. Pass key-value pairs as arguments to filter.",
	Example: `  manage-job get
  manage-job get page 1 pageSize 10 search "Acme" industry "Tech" status "Applied Only" order "desc"`,
	Run: func(cmd *cobra.Command, args []string) {
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
			os.Exit(1)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading response: %v\n", err)
			os.Exit(1)
		}

		fmt.Print(string(body))
	},
}

var trackCmd = &cobra.Command{
	Use:   "track <companyName> <link> <industry> <status> [email] [phone] [notes]",
	Short: "Record a new job application",
	Long: `Record a new job application in the spreadsheet.
Creates a new row with today's date.

Required args (in order):
  companyName — Name of the company
  link        — Job posting URL (must start with http:// or https://)
  industry    — Tech, Health Care, Retail, Finance, Gig, Other
  status      — Not Started, Applied Only, Applied + Emailed, Applied + Called, Applied + Emailed + Called, Interview!, Done

Optional args (in order):
  email — Employer contact email
  phone — Contact phone number (10-15 digits)
  notes — Free-form notes (all remaining args joined)`,
	Example: `  manage-job track "Acme Corp" "https://fakejobs.com/quantum-ai-analyst" "Tech" "Not Started"
  manage-job track "Acme Corp" "https://fakejobs.com/quantum-ai-analyst" "Tech" "Not Started" "email@email.com"
  manage-job track "Acme Corp" "https://fakejobs.com/quantum-ai-analyst" "Tech" "Not Started" "email@email.com" "917-999-1234" "They said to email John"`,
	Args: cobra.MinimumNArgs(4),
	Run: func(cmd *cobra.Command, args []string) {
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

		if err := validate.URL(link); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		if !validIndustries[industry] {
			fmt.Fprintf(os.Stderr, "Error: industry must be one of: Tech, Health Care, Retail, Finance, Gig, Other\n")
			os.Exit(1)
		}
		if !validStatuses[status] {
			fmt.Fprintf(os.Stderr, "Error: status must be one of: Not Started, Applied Only, Applied + Emailed, Applied + Called, Applied + Emailed + Called, Interview!, Done\n")
			os.Exit(1)
		}
		if email != "" {
			if err := validate.Email(email); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
		}
		if phone != "" {
			if err := validate.Phone(phone); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
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
			os.Exit(1)
		}

		scriptURL := loadScriptURL()
		result, err := postFollowRedirect(scriptURL, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if strings.Contains(result, `"error"`) {
			fmt.Fprintf(os.Stderr, "Fail: %s\n", result)
			os.Exit(1)
		}

		fmt.Printf("Success: %s\n", result)
	},
}

var rootCmd = &cobra.Command{
	Use:   "manage-job",
	Short: "Track and retrieve job applications",
}

var patchCmd = &cobra.Command{
	Use:   "patch --matchBy '<json>' --update '<json>'",
	Short: "Update an existing job application",
	Long: `Update an existing job application in the spreadsheet.
Uses matchBy to find the row, then applies the update fields.

--matchBy: JSON object with at least one field to identify the row.
  Example: '{"companyName":"Acme Corp"}'

--update: JSON object with at least one field to change.
  Example: '{"status":"Interview!"}'`,
	Example: `  manage-job patch --matchBy '{"companyName":"Acme Corp"}' --update '{"status":"Interview!"}'
  manage-job patch --matchBy '{"companyName":"Acme Corp","link":"https://example.com"}' --update '{"status":"Done","notes":"Rejected"}'`,
	Run: func(cmd *cobra.Command, args []string) {
		matchByStr, _ := cmd.Flags().GetString("matchBy")
		updateStr, _ := cmd.Flags().GetString("update")

		if matchByStr == "" {
			fmt.Fprintln(os.Stderr, "Error: --matchBy is required")
			os.Exit(1)
		}
		if updateStr == "" {
			fmt.Fprintln(os.Stderr, "Error: --update is required")
			os.Exit(1)
		}

		var matchBy map[string]interface{}
		if err := json.Unmarshal([]byte(matchByStr), &matchBy); err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid --matchBy JSON: %v\n", err)
			os.Exit(1)
		}

		var update map[string]interface{}
		if err := json.Unmarshal([]byte(updateStr), &update); err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid --update JSON: %v\n", err)
			os.Exit(1)
		}

		payload := map[string]interface{}{
			"action":  "patch",
			"matchBy": matchBy,
			"update":  update,
		}

		body, err := json.Marshal(payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding payload: %v\n", err)
			os.Exit(1)
		}

		scriptURL := loadScriptURL()
		result, err := postFollowRedirect(scriptURL, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if strings.Contains(result, `"error"`) {
			fmt.Fprintf(os.Stderr, "Fail: %s\n", result)
			os.Exit(1)
		}

		fmt.Printf("Success: %s\n", result)
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete --matchBy '<json>'",
	Short: "Delete a job application",
	Long: `Delete a job application from the spreadsheet.
Uses matchBy to find the row to delete.

--matchBy: JSON object with at least one field to identify the row.
  Example: '{"companyName":"Acme Corp"}'`,
	Example: `  manage-job delete --matchBy '{"companyName":"Acme Corp"}'
  manage-job delete --matchBy '{"companyName":"Acme Corp","link":"https://example.com"}'`,
	Run: func(cmd *cobra.Command, args []string) {
		matchByStr, _ := cmd.Flags().GetString("matchBy")

		if matchByStr == "" {
			fmt.Fprintln(os.Stderr, "Error: --matchBy is required")
			os.Exit(1)
		}

		var matchBy map[string]interface{}
		if err := json.Unmarshal([]byte(matchByStr), &matchBy); err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid --matchBy JSON: %v\n", err)
			os.Exit(1)
		}

		payload := map[string]interface{}{
			"action":  "delete",
			"matchBy": matchBy,
		}

		body, err := json.Marshal(payload)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding payload: %v\n", err)
			os.Exit(1)
		}

		scriptURL := loadScriptURL()
		result, err := postFollowRedirect(scriptURL, body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if strings.Contains(result, `"error"`) {
			fmt.Fprintf(os.Stderr, "Fail: %s\n", result)
			os.Exit(1)
		}

		fmt.Printf("Success: %s\n", result)
	},
}

func main() {
	patchCmd.Flags().String("matchBy", "", "JSON object to identify the row (required)")
	patchCmd.Flags().String("update", "", "JSON object with fields to change (required)")
	deleteCmd.Flags().String("matchBy", "", "JSON object to identify the row (required)")

	rootCmd.AddCommand(getCmd, trackCmd, patchCmd, deleteCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
