package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"manage-job/appscript"

	"github.com/spf13/cobra"
)

var PatchCmd = &cobra.Command{
	Use:   "patch --matchBy '<json>' --update '<json>'",
	Short: "Update an existing job application",
	Long: `Update an existing job application in the spreadsheet.
Uses matchBy to find the row, then applies the update fields.

--matchBy: JSON object with at least one field to identify the row.
  Example: '{"companyName":"Acme Corp"}'

--update: JSON object with at least one field to change.
  Example: '{"status":"Interview!"}'`,
	Example: `  manage-job patch --matchBy '{"companyName":"Acme Corp"}' --update '{"status":"Interview!"}'
  manage-job patch --matchBy '{"companyName":"Acme Corp","link":"https://example.com"}' --update '{"status":"Didn't Get It","notes":"Rejected"}'`,
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

		app := appscript.NewAppScript()
		result, err := app.Patch(matchBy, update)
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
