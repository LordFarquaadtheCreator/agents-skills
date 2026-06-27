package main

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "manage-job",
	Short: "Track and retrieve job applications",
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
