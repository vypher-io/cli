package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValid(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".vypher.yaml")

	content := `exclude:
  - "*_test.go"
  - "*.log"
rules:
  - finance
  - phi
output: sarif
max_depth: 5
fail_on_match: true
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Exclude) != 2 {
		t.Errorf("Exclude count = %d, want 2", len(cfg.Exclude))
	}
	if cfg.Exclude[0] != "*_test.go" {
		t.Errorf("Exclude[0] = %s, want *_test.go", cfg.Exclude[0])
	}

	if len(cfg.Rules) != 2 {
		t.Errorf("Rules count = %d, want 2", len(cfg.Rules))
	}
	if cfg.Rules[0] != "finance" {
		t.Errorf("Rules[0] = %s, want finance", cfg.Rules[0])
	}

	if cfg.Output != "sarif" {
		t.Errorf("Output = %s, want sarif", cfg.Output)
	}
	if cfg.MaxDepth != 5 {
		t.Errorf("MaxDepth = %d, want 5", cfg.MaxDepth)
	}
	if !cfg.FailOnMatch {
		t.Error("FailOnMatch = false, want true")
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/.vypher.yaml")
	if err == nil {
		t.Error("Load() expected error for missing file, got nil")
	}
}

func TestLoadPartialConfig(t *testing.T) {
	tmpDir := t.TempDir()
	cfgPath := filepath.Join(tmpDir, ".vypher.yaml")

	content := `exclude:
  - "*.tmp"
`
	if err := os.WriteFile(cfgPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}

	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if len(cfg.Exclude) != 1 {
		t.Errorf("Exclude count = %d, want 1", len(cfg.Exclude))
	}
	if cfg.Output != "" {
		t.Errorf("Output = %s, want empty string", cfg.Output)
	}
	if cfg.MaxDepth != 0 {
		t.Errorf("MaxDepth = %d, want 0", cfg.MaxDepth)
	}
	if cfg.FailOnMatch {
		t.Error("FailOnMatch = true, want false")
	}
	if len(cfg.Rules) != 0 {
		t.Errorf("Rules count = %d, want 0", len(cfg.Rules))
	}
}
