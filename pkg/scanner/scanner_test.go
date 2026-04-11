package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestIsBinary(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		want    bool
	}{
		{
			name:    "Text Content",
			content: []byte("Hello, world!"),
			want:    false,
		},
		{
			name:    "Binary Content",
			content: []byte{0x00, 0x01, 0x02},
			want:    true,
		},
		{
			name:    "Mixed Content with Null",
			content: []byte("Hello\x00World"),
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isBinary(tt.content); got != tt.want {
				t.Errorf("isBinary() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScan(t *testing.T) {
	tmpDir := t.TempDir()

	piiFile := filepath.Join(tmpDir, "pii.txt")
	os.WriteFile(piiFile, []byte("Contact: test@example.com"), 0644)

	cleanFile := filepath.Join(tmpDir, "clean.txt")
	os.WriteFile(cleanFile, []byte("Just some random text"), 0644)

	binFile := filepath.Join(tmpDir, "binary.bin")
	os.WriteFile(binFile, []byte{0x00, 0x01, 0x02}, 0644)

	results := Scan(tmpDir, ScanOptions{})

	if len(results) != 1 {
		t.Errorf("Scan() returned %d results, want 1", len(results))
		return
	}

	res := results[0]
	if res.FilePath != piiFile {
		t.Errorf("Scan() result file = %s, want %s", res.FilePath, piiFile)
	}

	if len(res.Matches) != 1 {
		t.Errorf("Scan() matches count = %d, want 1", len(res.Matches))
	} else if res.Matches[0].RuleName != "Email" {
		t.Errorf("Match rule = %s, want Email", res.Matches[0].RuleName)
	}
}

func TestScanWithExclude(t *testing.T) {
	tmpDir := t.TempDir()

	keepFile := filepath.Join(tmpDir, "data.txt")
	os.WriteFile(keepFile, []byte("Email: keep@example.com"), 0644)

	os.WriteFile(filepath.Join(tmpDir, "data_test.go"), []byte("Email: excluded@example.com"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "app.log"), []byte("Email: log@example.com"), 0644)

	results := Scan(tmpDir, ScanOptions{
		ExcludePatterns: []string{"*_test.go", "*.log"},
	})

	if len(results) != 1 {
		t.Errorf("Scan() with excludes returned %d results, want 1", len(results))
		return
	}

	if results[0].FilePath != keepFile {
		t.Errorf("Scan() result file = %s, want %s", results[0].FilePath, keepFile)
	}
}

func TestScanDefaultIgnores(t *testing.T) {
	tmpDir := t.TempDir()

	normalFile := filepath.Join(tmpDir, "source.txt")
	os.WriteFile(normalFile, []byte("Email: normal@example.com"), 0644)

	nmDir := filepath.Join(tmpDir, "node_modules")
	os.MkdirAll(nmDir, 0755)
	os.WriteFile(filepath.Join(nmDir, "package.txt"), []byte("Email: npm@example.com"), 0644)

	gitDir := filepath.Join(tmpDir, ".git")
	os.MkdirAll(gitDir, 0755)
	os.WriteFile(filepath.Join(gitDir, "config"), []byte("Email: git@example.com"), 0644)

	results := Scan(tmpDir, ScanOptions{})

	if len(results) != 1 {
		t.Errorf("Scan() returned %d results, want 1", len(results))
		return
	}

	if results[0].FilePath != normalFile {
		t.Errorf("Scan() result file = %s, want %s", results[0].FilePath, normalFile)
	}
}

func TestScanMaxDepth(t *testing.T) {
	tmpDir := t.TempDir()

	rootFile := filepath.Join(tmpDir, "root.txt")
	os.WriteFile(rootFile, []byte("Email: root@example.com"), 0644)

	subDir := filepath.Join(tmpDir, "sub")
	os.MkdirAll(subDir, 0755)
	os.WriteFile(filepath.Join(subDir, "sub.txt"), []byte("Email: sub@example.com"), 0644)

	deepDir := filepath.Join(subDir, "deep")
	os.MkdirAll(deepDir, 0755)
	os.WriteFile(filepath.Join(deepDir, "deep.txt"), []byte("Email: deep@example.com"), 0644)

	// MaxDepth=1 should only scan root files
	results := Scan(tmpDir, ScanOptions{MaxDepth: 1})
	if len(results) != 1 {
		t.Errorf("Scan(MaxDepth=1) returned %d results, want 1", len(results))
	}

	// MaxDepth=2 should scan root and sub files
	results = Scan(tmpDir, ScanOptions{MaxDepth: 2})
	if len(results) != 2 {
		t.Errorf("Scan(MaxDepth=2) returned %d results, want 2", len(results))
	}

	// MaxDepth=0 (unlimited) should scan all
	results = Scan(tmpDir, ScanOptions{MaxDepth: 0})
	if len(results) != 3 {
		t.Errorf("Scan(MaxDepth=0) returned %d results, want 3", len(results))
	}
}

func TestScanParallel(t *testing.T) {
	tmpDir := t.TempDir()

	// Create 20 files with PII to exercise the worker pool
	expectedFiles := 20
	for i := 0; i < expectedFiles; i++ {
		fname := filepath.Join(tmpDir, fmt.Sprintf("file_%02d.txt", i))
		os.WriteFile(fname, []byte(fmt.Sprintf("Contact: user%d@example.com", i)), 0644)
	}

	// Also create some clean files
	for i := 0; i < 10; i++ {
		fname := filepath.Join(tmpDir, fmt.Sprintf("clean_%02d.txt", i))
		os.WriteFile(fname, []byte("No sensitive data here"), 0644)
	}

	results := Scan(tmpDir, ScanOptions{})

	if len(results) != expectedFiles {
		t.Errorf("Scan() returned %d results, want %d", len(results), expectedFiles)
	}

	// Verify results are sorted by file path (deterministic output)
	for i := 1; i < len(results); i++ {
		if results[i].FilePath < results[i-1].FilePath {
			t.Errorf("Results not sorted: %s came after %s", results[i].FilePath, results[i-1].FilePath)
		}
	}

	// Run again to verify deterministic results
	results2 := Scan(tmpDir, ScanOptions{})
	if len(results2) != len(results) {
		t.Errorf("Second scan returned %d results, want %d (same as first)", len(results2), len(results))
	}
	for i := range results {
		if results[i].FilePath != results2[i].FilePath {
			t.Errorf("Non-deterministic order: run1[%d]=%s, run2[%d]=%s", i, results[i].FilePath, i, results2[i].FilePath)
		}
	}
}
