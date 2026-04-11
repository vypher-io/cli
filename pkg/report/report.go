package report

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/vypher-io/cli/pkg/engine"
	"github.com/fatih/color"
)

// FileResult holds the scan results for a single file
type FileResult struct {
	FilePath string         `json:"file_path"`
	Matches  []engine.Match `json:"matches"`
}

// Print outputs the results in the specified format
func Print(results []FileResult, format string) {
	switch format {
	case "json":
		printJSON(results)
	case "sarif":
		printSARIF(results)
	default:
		printConsole(results)
	}
}

func printJSON(results []FileResult) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(results); err != nil {
		fmt.Printf("Error encoding format JSON: %v\n", err)
	}
}

// --- SARIF v2.1.0 Output ---

type sarifDocument struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []sarifRun `json:"runs"`
}

type sarifRun struct {
	Tool    sarifTool     `json:"tool"`
	Results []sarifResult `json:"results"`
}

type sarifTool struct {
	Driver sarifDriver `json:"driver"`
}

type sarifDriver struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type sarifResult struct {
	RuleID    string          `json:"ruleId"`
	Level     string          `json:"level"`
	Message   sarifMessage    `json:"message"`
	Locations []sarifLocation `json:"locations"`
}

type sarifMessage struct {
	Text string `json:"text"`
}

type sarifLocation struct {
	PhysicalLocation sarifPhysicalLocation `json:"physicalLocation"`
}

type sarifPhysicalLocation struct {
	ArtifactLocation sarifArtifactLocation `json:"artifactLocation"`
	Region           sarifRegion           `json:"region"`
}

type sarifArtifactLocation struct {
	URI string `json:"uri"`
}

type sarifRegion struct {
	CharOffset int `json:"charOffset"`
	CharLength int `json:"charLength"`
}

func printSARIF(results []FileResult) {
	var sarifResults []sarifResult
	for _, fr := range results {
		for _, m := range fr.Matches {
			sarifResults = append(sarifResults, sarifResult{
				RuleID: m.RuleName,
				Level:  "warning",
				Message: sarifMessage{
					Text: fmt.Sprintf("Detected %s: %s", m.RuleName, maskContent(m.Content)),
				},
				Locations: []sarifLocation{
					{
						PhysicalLocation: sarifPhysicalLocation{
							ArtifactLocation: sarifArtifactLocation{
								URI: fr.FilePath,
							},
							Region: sarifRegion{
								CharOffset: m.Index,
								CharLength: len(m.Content),
							},
						},
					},
				},
			})
		}
	}

	doc := sarifDocument{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs: []sarifRun{
			{
				Tool: sarifTool{
					Driver: sarifDriver{
						Name:    "vypher",
						Version: "0.1.0",
					},
				},
				Results: sarifResults,
			},
		},
	}

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(doc); err != nil {
		fmt.Printf("Error encoding SARIF: %v\n", err)
	}
}

func printConsole(results []FileResult) {
	red := color.New(color.FgRed).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	totalFiles := len(results)
	totalMatches := 0

	for _, res := range results {
		totalMatches += len(res.Matches)
	}

	fmt.Printf("\nScan Complete.\n")
	fmt.Printf("Scanned %d files with findings.\n", totalFiles)
	fmt.Printf("Total %s found: %d\n\n", red("Issues"), totalMatches)

	for _, res := range results {
		if len(res.Matches) > 0 {
			fmt.Printf("%s: %s\n", green("File"), res.FilePath)
			for _, match := range res.Matches {
				maskedContent := maskContent(match.Content)
				fmt.Printf("  - [%s] %s (Index: %d)\n", yellow(match.RuleName), maskedContent, match.Index)
			}
			fmt.Println()
		}
	}
}

func maskContent(content string) string {
	if len(content) <= 4 {
		return "****"
	}
	return content[:2] + "****" + content[len(content)-2:]
}
