package config

import (
	"os"

	"go.yaml.in/yaml/v3"
)

// Config holds the configuration loaded from a YAML file.
type Config struct {
	Exclude     []string `yaml:"exclude"`
	Rules       []string `yaml:"rules"`
	Output      string   `yaml:"output"`
	MaxDepth    int      `yaml:"max_depth"`
	FailOnMatch bool     `yaml:"fail_on_match"`
}

// Load reads and parses a YAML config file at the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- path is the user-supplied config file path from the CLI flag
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
