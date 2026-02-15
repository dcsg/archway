package output

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/dcsg/archway/internal/provider"
)

func sampleResult() *provider.AnalyzeResponse {
	return &provider.AnalyzeResponse{
		Language:      "go",
		PackageCount:  5,
		FileCount:     12,
		FunctionCount: 20,
		Architecture: provider.ArchitectureResult{
			Pattern:    "hexagonal",
			Confidence: 0.89,
			Evidence:   []string{"found domain"},
		},
		Framework: provider.FrameworkResult{Name: "chi", Confidence: 0.95},
		Conventions: provider.ConventionResults{
			ErrorHandling: provider.ConventionFinding{Pattern: "wrapped"},
			Logging:       provider.ConventionFinding{Pattern: "slog/structured"},
			Config:        provider.ConventionFinding{Pattern: "koanf"},
			Testing:       provider.TestingFinding{Pattern: "table-driven", TestFiles: 3, TotalGoFiles: 12},
		},
	}
}

func TestFormatters(t *testing.T) {
	result := sampleResult()

	terminal, err := NewFormatter("terminal", true)
	if err != nil {
		t.Fatalf("NewFormatter terminal: %v", err)
	}
	out, err := terminal.Format(result)
	if err != nil {
		t.Fatalf("terminal format: %v", err)
	}
	if !strings.Contains(out, "Project Summary") {
		t.Fatalf("terminal output missing section: %s", out)
	}

	jsonFmt, err := NewFormatter("json", true)
	if err != nil {
		t.Fatalf("NewFormatter json: %v", err)
	}
	jsonOut, err := jsonFmt.Format(result)
	if err != nil {
		t.Fatalf("json format: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonOut), &parsed); err != nil {
		t.Fatalf("invalid json: %v", err)
	}

	mdFmt, err := NewFormatter("markdown", true)
	if err != nil {
		t.Fatalf("NewFormatter markdown: %v", err)
	}
	mdOut, err := mdFmt.Format(result)
	if err != nil {
		t.Fatalf("markdown format: %v", err)
	}
	if !strings.Contains(mdOut, "## Architecture") {
		t.Fatalf("markdown output missing architecture section: %s", mdOut)
	}
}
