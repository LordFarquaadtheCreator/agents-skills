package pats

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var patsPath = filepath.Join(os.Getenv("HOME"), "agents-data", "config", "gh-pats.yaml")

func modeToKey(mode string) (string, error) {
	switch mode {
	case "work_mode":
		return "work_PAT", nil
	case "personal_mode":
		return "personal_PAT", nil
	default:
		return "", fmt.Errorf("invalid mode: %s. Must be 'work_mode' or 'personal_mode'", mode)
	}
}

func LoadToken(mode string) (string, error) {
	key, err := modeToKey(mode)
	if err != nil {
		return "", err
	}

	data, err := os.ReadFile(patsPath)
	if err != nil {
		return "", fmt.Errorf("PATs file not found: %s", patsPath)
	}

	var pats struct {
		PersonalPAT string `yaml:"personal_PAT"`
		WorkPAT     string `yaml:"work_PAT"`
	}
	if err := yaml.Unmarshal(data, &pats); err != nil {
		return "", fmt.Errorf("invalid YAML in PATs file: %v", err)
	}

	var token string
	switch key {
	case "work_PAT":
		token = pats.WorkPAT
	case "personal_PAT":
		token = pats.PersonalPAT
	}
	if token == "" {
		return "", fmt.Errorf("key '%s' not found in %s", key, patsPath)
	}

	return token, nil
}
