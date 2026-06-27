package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var DeleteCmd = &cobra.Command{
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

		app := NewAppScript()
		result, err := app.Delete(matchBy)
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
