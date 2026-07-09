package main

import (
	"fmt"
	"os"

	"github.com/LordFarquaadtheCreator/create-story/internal/mcpserver"
)

func main() {
	if err := mcpserver.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
		os.Exit(1)
	}
}
