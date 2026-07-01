package main

import (
	"os"

	"github.com/spf13/cobra"

	"set-gh-token/cmd"
)

var rootCmd = &cobra.Command{
	Use:   "set-gh-token",
	Short: "Swap GitHub tokens between work and personal modes",
}

func main() {
	rootCmd.AddCommand(cmd.McpCmd, cmd.CliCmd)
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
