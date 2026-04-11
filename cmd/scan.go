package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/vypher-io/cli/pkg/config"
	"github.com/vypher-io/cli/pkg/report"
	"github.com/vypher-io/cli/pkg/scanner"
)

var (
	targetDir   string
	outputFmt   string
	exclude     []string
	failOnMatch bool
	rules       []string
	maxDepth    int
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan a directory for PII/PHI",
	Long: `Scan a directory recursively for files containing PII (Personally Identifiable Information) 
or PHI (Protected Health Information).

The following directories are ignored by default: .git, node_modules, vendor, .venv, __pycache__

Configuration can be loaded from a YAML file with --config. CLI flags override config file values.

Examples:
  vypher scan --target ./src
  vypher scan -t ./src -o json
  vypher scan -t ./src -o sarif
  vypher scan -t ./src --exclude "*_test.go" --exclude "*.log"
  vypher scan -t ./src --rules finance,phi
  vypher scan -t ./src --max-depth 3
  vypher scan --config .vypher.yaml --fail-on-match`,
	Run: func(cmd *cobra.Command, args []string) {
		// Load config file if specified
		var cfg *config.Config
		if cfgFile != "" {
			var err error
			cfg, err = config.Load(cfgFile)
			if err != nil {
				fmt.Printf("Error loading config file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Loaded config from: %s\n", cfgFile)
		}

		// Merge: CLI flags override config file values
		mergedExclude := mergeStringSlice(exclude, cfgSlice(cfg, "exclude"))
		mergedRules := mergeStringSlice(rules, cfgSlice(cfg, "rules"))
		mergedOutput := mergeString(cmd, "output", outputFmt, cfgString(cfg, "output"))
		mergedMaxDepth := mergeInt(cmd, "max-depth", maxDepth, cfgInt(cfg, "max_depth"))
		mergedFailOnMatch := mergeBool(cmd, "fail-on-match", failOnMatch, cfgBool(cfg, "fail_on_match"))

		if targetDir == "" {
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Printf("Error getting current directory: %v\n", err)
				os.Exit(1)
			}
			targetDir = cwd
		}

		fmt.Printf("Scanning directory: %s\n", targetDir)
		results := scanner.Scan(targetDir, scanner.ScanOptions{
			ExcludePatterns: mergedExclude,
			MaxDepth:        mergedMaxDepth,
			RuleTags:        mergedRules,
		})
		report.Print(results, mergedOutput)

		if mergedFailOnMatch && len(results) > 0 {
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(scanCmd)

	scanCmd.Flags().StringVarP(&targetDir, "target", "t", "", "Directory to scan (defaults to current directory)")
	scanCmd.Flags().StringVarP(&outputFmt, "output", "o", "console", "Output format (console, json, sarif)")
	scanCmd.Flags().StringSliceVarP(&exclude, "exclude", "e", nil, "Glob patterns to exclude (e.g. \"*_test.go\")")
	scanCmd.Flags().BoolVar(&failOnMatch, "fail-on-match", false, "Exit with code 1 if any issues are found (for CI/CD)")
	scanCmd.Flags().StringSliceVar(&rules, "rules", nil, "Rule tags to enable (e.g. \"finance,phi\")")
	scanCmd.Flags().IntVar(&maxDepth, "max-depth", 0, "Maximum directory recursion depth (0 = unlimited)")
}

// --- Config merge helpers ---
// CLI flags take precedence; config provides defaults for unset flags.

func cfgSlice(cfg *config.Config, field string) []string {
	if cfg == nil {
		return nil
	}
	switch field {
	case "exclude":
		return cfg.Exclude
	case "rules":
		return cfg.Rules
	}
	return nil
}

func cfgString(cfg *config.Config, field string) string {
	if cfg == nil {
		return ""
	}
	switch field {
	case "output":
		return cfg.Output
	}
	return ""
}

func cfgInt(cfg *config.Config, field string) int {
	if cfg == nil {
		return 0
	}
	switch field {
	case "max_depth":
		return cfg.MaxDepth
	}
	return 0
}

func cfgBool(cfg *config.Config, field string) bool {
	if cfg == nil {
		return false
	}
	switch field {
	case "fail_on_match":
		return cfg.FailOnMatch
	}
	return false
}

// mergeStringSlice combines CLI and config slices (both are additive).
func mergeStringSlice(cli, cfg []string) []string {
	if len(cli) > 0 {
		return cli
	}
	return cfg
}

// mergeString returns CLI value if the flag was explicitly set, otherwise config value.
func mergeString(cmd *cobra.Command, flag, cliVal, cfgVal string) string {
	if cmd.Flags().Changed(flag) {
		return cliVal
	}
	if cfgVal != "" {
		return cfgVal
	}
	return cliVal // return default
}

// mergeInt returns CLI value if the flag was explicitly set, otherwise config value.
func mergeInt(cmd *cobra.Command, flag string, cliVal, cfgVal int) int {
	if cmd.Flags().Changed(flag) {
		return cliVal
	}
	if cfgVal != 0 {
		return cfgVal
	}
	return cliVal
}

// mergeBool returns CLI value if the flag was explicitly set, otherwise config value.
func mergeBool(cmd *cobra.Command, flag string, cliVal, cfgVal bool) bool {
	if cmd.Flags().Changed(flag) {
		return cliVal
	}
	return cfgVal
}
