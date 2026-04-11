package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra/doc"
	"github.com/vypher-io/cli/cmd"
)

func main() {
	rootCmd := cmd.GetRootCmd()

	outputDir := "./docs/cli"
	if len(os.Args) > 1 {
		outputDir = os.Args[1]
	}

	// Sanitize: resolve to absolute path and reject path traversal attempts
	absOutput, err := filepath.Abs(outputDir)
	if err != nil {
		log.Fatalf("Invalid output directory: %v", err)
	}
	if strings.Contains(absOutput, "..") {
		log.Fatal("Output directory must not contain path traversal sequences")
	}

	if err := os.MkdirAll(absOutput, 0750); err != nil { // #nosec G703 -- path is sanitized via filepath.Abs above
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Generate Markdown documentation
	// Using standard Markdown (not man pages)
	err = doc.GenMarkdownTree(rootCmd, absOutput)
	if err != nil {
		log.Fatalf("Failed to generate documentation: %v", err)
	}

	log.Printf("Documentation generated in %s", absOutput) // #nosec G706 -- absOutput is a resolved absolute path, not raw user input
}
