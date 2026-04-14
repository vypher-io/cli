package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var cfgFile string
var appVersion = "dev"

// SetVersion is called from main to inject the build-time version.
func SetVersion(v string) {
	appVersion = v
}

var rootCmd = &cobra.Command{
	Use:   "vypher",
	Short: "Vypher is a PII and PHI scanning tool",
	Long: `Vypher is a CLI tool designed to scan directories for Personally Identifiable Information (PII) 
and Protected Health Information (PHI) with a focus on finance and healthcare data.

It helps developers and security professionals identify sensitive data leaks in their codebase or file systems.

Use --config to load settings from a YAML configuration file.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of vypher",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("vypher %s\n", appVersion)
	},
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

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Open the Vypher documentation in your browser",
	Run: func(cmd *cobra.Command, args []string) {
		url := "https://docs.vypher.io"
		var err error
		switch runtime.GOOS {
		case "darwin":
			err = exec.Command("open", url).Start()
		case "windows":
			err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
		default:
			err = exec.Command("xdg-open", url).Start()
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not open browser: %v\nVisit: %s\n", err, url)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (e.g. .vypher.yaml)")
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(docsCmd)
}
