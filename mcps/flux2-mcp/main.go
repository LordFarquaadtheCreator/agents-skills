package main

import (
	"fmt"
	"log"
	"os"

	"mcps/flux2-mcp/internal/mcpserver"
)

func main() {
	apiURL := os.Getenv("FLUX2_API_URL")
	if apiURL == "" {
		log.Fatal("FLUX2_API_URL env var is required")
	}

	mode := "mcp"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	switch mode {
	case "mcp":
		if err := mcpserver.Run(apiURL); err != nil {
			fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
			os.Exit(1)
		}
	case "cli":
		if err := mcpserver.RunCLI(apiURL, os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "cli error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "unknown mode %q: use 'cli' or 'mcp'\n", mode)
		os.Exit(1)
	}
}
