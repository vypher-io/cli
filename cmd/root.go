package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "vypher",
	Short: "Vypher is a PII and PHI scanning tool",
	Long: `Vypher is a CLI tool designed to scan directories for Personally Identifiable Information (PII) 
and Protected Health Information (PHI) with a focus on finance and healthcare data.

It helps developers and security professionals identify sensitive data leaks in their codebase or file systems.

Use --config to load settings from a YAML configuration file.`,
}

func GetRootCmd() *cobra.Command {
	return rootCmd
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (e.g. .vypher.yaml)")
}
