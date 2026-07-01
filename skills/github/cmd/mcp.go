package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"set-gh-token/pats"

	"github.com/spf13/cobra"
)

var mcpConfigPath = filepath.Join(os.Getenv("HOME"), ".codeium", "windsurf", "mcp_config.json")

var McpCmd = &cobra.Command{
	Use:   "mcp <work_mode|personal_mode>",
	Short: "Swap GitHub MCP token in mcp_config.json",
	Long:  "Swap GitHub PAT in mcp_config.json based on mode.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mode := args[0]

		token, err := pats.LoadToken(mode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		data, err := os.ReadFile(mcpConfigPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: MCP config not found: %s\n", mcpConfigPath)
			os.Exit(1)
		}

		var config map[string]interface{}
		if err := json.Unmarshal(data, &config); err != nil {
			fmt.Fprintf(os.Stderr, "Error: invalid JSON in MCP config: %v\n", err)
			os.Exit(1)
		}

		mcpServers, ok := config["mcpServers"].(map[string]interface{})
		if !ok {
			fmt.Fprintln(os.Stderr, "Error: required key missing in config: mcpServers")
			os.Exit(1)
		}

		githubEntry, ok := mcpServers["github"].(map[string]interface{})
		if !ok {
			fmt.Fprintln(os.Stderr, "Error: required key missing in config: mcpServers.github")
			os.Exit(1)
		}

		headers, ok := githubEntry["headers"].(map[string]interface{})
		if !ok {
			fmt.Fprintln(os.Stderr, "Error: required key missing in config: mcpServers.github.headers")
			os.Exit(1)
		}

		headers["Authorization"] = fmt.Sprintf("Bearer %s", token)

		output, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to marshal config: %v\n", err)
			os.Exit(1)
		}

		if err := os.WriteFile(mcpConfigPath, output, 0644); err != nil {
			fmt.Fprintf(os.Stderr, "Error: cannot write to %s\n", mcpConfigPath)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated GitHub token to %s\n", mode)
	},
}
