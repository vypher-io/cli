package cmd

import (
	"testing"
)

func TestScanCmdFlags(t *testing.T) {
	targetFlag := scanCmd.Flags().Lookup("target")
	if targetFlag == nil {
		t.Error("scan command missing 'target' flag")
	}
	if targetFlag.DefValue != "" {
		t.Errorf("scan command 'target' flag default = %s, want empty", targetFlag.DefValue)
	}

	outputFlag := scanCmd.Flags().Lookup("output")
	if outputFlag == nil {
		t.Error("scan command missing 'output' flag")
	}
	if outputFlag.DefValue != "console" {
		t.Errorf("scan command 'output' flag default = %s, want 'console'", outputFlag.DefValue)
	}

	excludeFlag := scanCmd.Flags().Lookup("exclude")
	if excludeFlag == nil {
		t.Error("scan command missing 'exclude' flag")
	}
	if excludeFlag.DefValue != "[]" {
		t.Errorf("scan command 'exclude' flag default = %s, want '[]'", excludeFlag.DefValue)
	}

	failOnMatchFlag := scanCmd.Flags().Lookup("fail-on-match")
	if failOnMatchFlag == nil {
		t.Error("scan command missing 'fail-on-match' flag")
	}
	if failOnMatchFlag.DefValue != "false" {
		t.Errorf("scan command 'fail-on-match' flag default = %s, want 'false'", failOnMatchFlag.DefValue)
	}

	rulesFlag := scanCmd.Flags().Lookup("rules")
	if rulesFlag == nil {
		t.Error("scan command missing 'rules' flag")
	}
	if rulesFlag.DefValue != "[]" {
		t.Errorf("scan command 'rules' flag default = %s, want '[]'", rulesFlag.DefValue)
	}

	maxDepthFlag := scanCmd.Flags().Lookup("max-depth")
	if maxDepthFlag == nil {
		t.Error("scan command missing 'max-depth' flag")
	}
	if maxDepthFlag.DefValue != "0" {
		t.Errorf("scan command 'max-depth' flag default = %s, want '0'", maxDepthFlag.DefValue)
	}
}

func TestRootCmdConfigFlag(t *testing.T) {
	configFlag := rootCmd.PersistentFlags().Lookup("config")
	if configFlag == nil {
		t.Error("root command missing 'config' persistent flag")
	}
	if configFlag.DefValue != "" {
		t.Errorf("root command 'config' flag default = %s, want empty", configFlag.DefValue)
	}
	if configFlag.Shorthand != "c" {
		t.Errorf("root command 'config' shorthand = %s, want 'c'", configFlag.Shorthand)
	}
}

func TestScanCmdRun(t *testing.T) {
	if scanCmd.Use != "scan" {
		t.Errorf("scan command use = %s, want 'scan'", scanCmd.Use)
	}
}
