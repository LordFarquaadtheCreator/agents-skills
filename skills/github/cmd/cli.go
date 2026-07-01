package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"set-gh-token/pats"

	"github.com/spf13/cobra"
)

var CliCmd = &cobra.Command{
	Use:   "cli <work_mode|personal_mode>",
	Short: "Swap GitHub CLI token",
	Long:  "Swap GitHub CLI token using gh auth login --with-token based on mode.",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		mode := args[0]

		token, err := pats.LoadToken(mode)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		ghCmd := exec.Command("gh", "auth", "login", "--with-token")
		ghCmd.Stdin = strings.NewReader(token)
		ghCmd.Stdout = os.Stdout
		ghCmd.Stderr = os.Stderr

		if err := ghCmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: gh auth login failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Successfully updated gh CLI token to %s\n", mode)
	},
}
