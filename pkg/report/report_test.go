package report

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/vypher-io/cli/pkg/engine"
)

func TestMaskContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "Short string",
			content: "123",
			want:    "****",
		},
		{
			name:    "Exact 4 chars",
			content: "1234",
			want:    "****",
		},
		{
			name:    "Long string",
			content: "123456789",
			want:    "12****89",
		},
		{
			name:    "Credit Card",
			content: "4111222233334444",
			want:    "41****44",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maskContent(tt.content); got != tt.want {
				t.Errorf("maskContent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSARIFOutput(t *testing.T) {
	results := []FileResult{
		{
			FilePath: "/tmp/test.txt",
			Matches: []engine.Match{
				{
					RuleName: "SSN",
					Content:  "123-45-6789",
					Index:    10,
				},
			},
		},
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	printSARIF(results)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	buf.ReadFrom(r)
	output := buf.String()

	// Parse as JSON and verify SARIF structure
	var doc sarifDocument
	if err := json.Unmarshal([]byte(output), &doc); err != nil {
		t.Fatalf("SARIF output is not valid JSON: %v", err)
	}

	if doc.Version != "2.1.0" {
		t.Errorf("SARIF version = %s, want 2.1.0", doc.Version)
	}

	if !strings.Contains(doc.Schema, "sarif-schema") {
		t.Errorf("SARIF $schema does not reference sarif-schema: %s", doc.Schema)
	}

	if len(doc.Runs) != 1 {
		t.Fatalf("SARIF runs count = %d, want 1", len(doc.Runs))
	}

	if doc.Runs[0].Tool.Driver.Name != "vypher" {
		t.Errorf("SARIF tool name = %s, want vypher", doc.Runs[0].Tool.Driver.Name)
	}

	if len(doc.Runs[0].Results) != 1 {
		t.Fatalf("SARIF results count = %d, want 1", len(doc.Runs[0].Results))
	}

	result := doc.Runs[0].Results[0]
	if result.RuleID != "SSN" {
		t.Errorf("SARIF result ruleId = %s, want SSN", result.RuleID)
	}

	if result.Level != "warning" {
		t.Errorf("SARIF result level = %s, want warning", result.Level)
	}
}

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

func TestPrintJSON(t *testing.T) {
	results := []FileResult{
		{
			FilePath: "/tmp/data.txt",
			Matches: []engine.Match{
				{RuleName: "Email", Content: "test@example.com", Index: 5},
			},
		},
	}

	output := captureStdout(func() { printJSON(results) })

	var decoded []FileResult
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("printJSON output is not valid JSON: %v", err)
	}

	if len(decoded) != 1 {
		t.Fatalf("printJSON decoded length = %d, want 1", len(decoded))
	}
	if decoded[0].FilePath != "/tmp/data.txt" {
		t.Errorf("printJSON FilePath = %s, want /tmp/data.txt", decoded[0].FilePath)
	}
	if len(decoded[0].Matches) != 1 || decoded[0].Matches[0].RuleName != "Email" {
		t.Errorf("printJSON match RuleName = %v, want Email", decoded[0].Matches)
	}
}

func TestPrintJSONEmpty(t *testing.T) {
	output := captureStdout(func() { printJSON(nil) })
	output = strings.TrimSpace(output)
	if output != "null" {
		t.Errorf("printJSON(nil) = %q, want \"null\"", output)
	}
}

func TestPrintConsole(t *testing.T) {
	results := []FileResult{
		{
			FilePath: "/tmp/report.txt",
			Matches: []engine.Match{
				{RuleName: "SSN", Content: "123-45-6789", Index: 0},
			},
		},
	}

	output := captureStdout(func() { printConsole(results) })

	if !strings.Contains(output, "Scan Complete") {
		t.Errorf("printConsole output missing 'Scan Complete': %s", output)
	}
	if !strings.Contains(output, "/tmp/report.txt") {
		t.Errorf("printConsole output missing file path: %s", output)
	}
	if !strings.Contains(output, "SSN") {
		t.Errorf("printConsole output missing rule name 'SSN': %s", output)
	}
}

func TestPrintConsoleEmpty(t *testing.T) {
	output := captureStdout(func() { printConsole(nil) })
	if !strings.Contains(output, "Scan Complete") {
		t.Errorf("printConsole(nil) output missing 'Scan Complete': %s", output)
	}
	if !strings.Contains(output, "0") {
		t.Errorf("printConsole(nil) output missing zero count: %s", output)
	}
}

func TestPrintDispatch(t *testing.T) {
	results := []FileResult{
		{FilePath: "/tmp/test.txt", Matches: []engine.Match{{RuleName: "Email", Content: "a@b.com", Index: 0}}},
	}

	// console (default)
	consoleOut := captureStdout(func() { Print(results, "console") })
	if !strings.Contains(consoleOut, "Scan Complete") {
		t.Errorf("Print(console) missing 'Scan Complete'")
	}

	// json
	jsonOut := captureStdout(func() { Print(results, "json") })
	var decoded []FileResult
	if err := json.Unmarshal([]byte(jsonOut), &decoded); err != nil {
		t.Errorf("Print(json) is not valid JSON: %v", err)
	}

	// sarif
	sarifOut := captureStdout(func() { Print(results, "sarif") })
	var doc sarifDocument
	if err := json.Unmarshal([]byte(sarifOut), &doc); err != nil {
		t.Errorf("Print(sarif) is not valid JSON: %v", err)
	}

	// unknown format falls back to console
	defaultOut := captureStdout(func() { Print(results, "unknown") })
	if !strings.Contains(defaultOut, "Scan Complete") {
		t.Errorf("Print(unknown) should fall back to console output")
	}
}
