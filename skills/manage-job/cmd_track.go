package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"manage-job/validate"

	"github.com/spf13/cobra"
)

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
