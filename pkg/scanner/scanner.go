package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/vypher-io/cli/pkg/engine"
	"github.com/vypher-io/cli/pkg/report"
)

// DefaultIgnoreDirs contains directory names that are skipped by default.
var DefaultIgnoreDirs = []string{".git", "node_modules", "vendor", ".venv", "__pycache__"}

// ScanOptions holds configurable options for the Scan function.
type ScanOptions struct {
	ExcludePatterns []string
	MaxDepth        int      // 0 means unlimited
	RuleTags        []string // empty means all rules
}

// Scan recursively scans the directory for PII/PHI with the given options.
// File scanning is parallelized across available CPU cores.
func Scan(rootDir string, opts ScanOptions) []report.FileResult {
	rootDir = filepath.Clean(rootDir)

	// Phase 1: Walk and collect file paths
	filePaths := collectFiles(rootDir, opts)

	// Phase 2: Parallel scan with worker pool
	numWorkers := runtime.NumCPU()
	if numWorkers < 1 {
		numWorkers = 1
	}
	if len(filePaths) < numWorkers {
		numWorkers = len(filePaths)
	}
	if numWorkers == 0 {
		return nil
	}

	pathCh := make(chan string, len(filePaths))
	resultCh := make(chan report.FileResult, len(filePaths))
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range pathCh {
				matches := scanFile(path, opts.RuleTags)
				if len(matches) > 0 {
					resultCh <- report.FileResult{
						FilePath: path,
						Matches:  matches,
					}
				}
			}
		}()
	}

	// Send files to workers
	for _, p := range filePaths {
		pathCh <- p
	}
	close(pathCh)

	// Wait for all workers to finish, then close results
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	// Collect results
	var results []report.FileResult
	for r := range resultCh {
		results = append(results, r)
	}

	// Sort by file path for deterministic output
	sort.Slice(results, func(i, j int) bool {
		return results[i].FilePath < results[j].FilePath
	})

	return results
}

// collectFiles walks the directory tree and returns a list of file paths to scan.
func collectFiles(rootDir string, opts ScanOptions) []string {
	var paths []string

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("Error accessing path %s: %v\n", path, err)
			return nil
		}

		// Calculate depth relative to rootDir
		relPath, _ := filepath.Rel(rootDir, path)
		depth := 0
		if relPath != "." {
			depth = strings.Count(relPath, string(filepath.Separator)) + 1
		}

		if d.IsDir() {
			if opts.MaxDepth > 0 && depth > opts.MaxDepth {
				return filepath.SkipDir
			}

			base := d.Name()
			for _, ignored := range DefaultIgnoreDirs {
				if base == ignored {
					return filepath.SkipDir
				}
			}
			for _, pattern := range opts.ExcludePatterns {
				if matched, _ := filepath.Match(pattern, base); matched {
					return filepath.SkipDir
				}
			}
			return nil
		}

		if opts.MaxDepth > 0 && depth > opts.MaxDepth {
			return nil
		}

		for _, pattern := range opts.ExcludePatterns {
			if matched, _ := filepath.Match(pattern, d.Name()); matched {
				return nil
			}
			if strings.Contains(pattern, string(filepath.Separator)) || strings.Contains(pattern, "/") {
				if matched, _ := filepath.Match(pattern, relPath); matched {
					return nil
				}
			}
		}

		paths = append(paths, path)
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
	}

	return paths
}

func scanFile(path string, tags []string) []engine.Match {
	content, err := os.ReadFile(path) // #nosec G304 -- by design: this tool scans arbitrary user-specified files
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", path, err)
		return nil
	}

	if isBinary(content) {
		return nil
	}

	return engine.ScanContentWithTags(string(content), tags)
}

func isBinary(content []byte) bool {
	for _, b := range content {
		if b == 0 {
			return true
		}
	}
	return false
}
