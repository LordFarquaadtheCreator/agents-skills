package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var TrackCmd = &cobra.Command{
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
		entry := map[string]interface{}{
			"companyName": args[0],
			"link":        args[1],
			"industry":    args[2],
			"status":      args[3],
			"dateApplied": time.Now().Format("2006-01-02"),
		}
		if len(args) > 4 {
			entry["email"] = args[4]
		}
		if len(args) > 5 {
			entry["phoneNumber"] = args[5]
		}
		if len(args) > 6 {
			entry["notes"] = strings.Join(args[6:], " ")
		}

		app := NewAppScript()
		result, err := app.Create(entry)
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
