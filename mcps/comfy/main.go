// comfy is a generic MCP server for the comfy-headless Modal app: every
// workflow in the shared store (~/senor-modal-apps/comfy-headless/workflows)
// is exposed as its own tool, with parameters generated from the workflow's
// {{token}} placeholders and LoadImage inputs.
package main

import (
	"fmt"
	"log"
	"os"

	"mcps/comfy/internal/config"
	"mcps/comfy/internal/headless"
	"mcps/comfy/internal/server"
	"mcps/comfy/internal/store"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatal(err)
	}
	if err := config.Validate(); err != nil {
		log.Fatal(err)
	}

	workflows, err := store.Load(config.Cfg.WorkflowsDir)
	if err != nil {
		log.Fatal(err)
	}

	client := headless.NewClient(config.Cfg.HeadlessURL, config.Cfg.ModalRegion, config.Cfg.WorkflowsDir)
	if config.Cfg.SubmitURL != "" {
		client.SubmitURL = config.Cfg.SubmitURL
		client.StatusURL = config.Cfg.StatusURL
	}

	fmt.Fprintf(os.Stderr, "comfy MCP: %d workflow tools from %s\n", len(workflows), config.Cfg.WorkflowsDir)

	if err := server.Run(client, workflows, config.Cfg.OutputDir); err != nil {
		fmt.Fprintf(os.Stderr, "mcp server error: %v\n", err)
		os.Exit(1)
	}
}
