package appscript

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type sheetsConfig struct {
	DeploymentID string `yaml:"deploymentId"`
}

func repoRoot() string {
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot resolve executable path: %v\n", err)
		os.Exit(1)
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(exe)))
}

func loadScriptURL() string {
	configPath := filepath.Join(repoRoot(), "config", "sheets-deployment.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: cannot read config/sheets-deployment.yaml: %v\n", err)
		os.Exit(1)
	}
	var cfg sheetsConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid YAML: %v\n", err)
		os.Exit(1)
	}
	if cfg.DeploymentID == "" {
		fmt.Fprintf(os.Stderr, "Error: deploymentId not found in config/sheets-deployment.yaml\n")
		os.Exit(1)
	}
	return fmt.Sprintf("https://script.google.com/macros/s/%s/exec", cfg.DeploymentID)
}
