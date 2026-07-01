package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	apiURL := os.Getenv("COMFYUI_API_URL")
	if apiURL == "" {
		log.Fatal("COMFYUI_API_URL env var is required")
	}

	if err := runMCPServer(apiURL); err != nil {
		fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
		os.Exit(1)
	}
}
