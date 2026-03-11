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

func TestFormatters_NilResult(t *testing.T) {
	formatters := []struct {
		name   string
		format string
	}{
		{"terminal", "terminal"},
		{"json", "json"},
		{"markdown", "markdown"},
	}
	for _, tc := range formatters {
		t.Run(tc.name, func(t *testing.T) {
			f, err := NewFormatter(tc.format, true)
			if err != nil {
				t.Fatalf("NewFormatter(%q): %v", tc.format, err)
			}
			_, err = f.Format(nil)
			if err == nil {
				t.Fatalf("%s formatter should return error for nil result", tc.name)
			}
		})
	}
}

func TestFormatters_EmptyResult(t *testing.T) {
	empty := &provider.AnalyzeResponse{}
	formatters := []string{"terminal", "json", "markdown"}
	for _, format := range formatters {
		t.Run(format, func(t *testing.T) {
			f, err := NewFormatter(format, true)
			if err != nil {
				t.Fatalf("NewFormatter(%q): %v", format, err)
			}
			out, err := f.Format(empty)
			if err != nil {
				t.Fatalf("%s format error on empty result: %v", format, err)
			}
			if out == "" {
				t.Fatalf("%s format returned empty string for empty result", format)
			}
		})
	}
}

func TestTerminalFormatter_ColorEnabled(t *testing.T) {
	f := &TerminalFormatter{NoColor: false}
	out, err := f.Format(sampleResult())
	if err != nil {
		t.Fatalf("format error: %v", err)
	}
	if !strings.Contains(out, "\033[1m") {
		t.Fatalf("expected ANSI bold codes when NoColor=false, got: %s", out)
	}
}

func TestTerminalFormatter_ViolationsAndCycles(t *testing.T) {
	result := sampleResult()
	result.DependencyGraph.Cycles = [][]string{{"a", "b", "a"}}
	result.Violations = []provider.Violation{
		{Rule: "dependency", Message: "domain must not depend on adapters", Source: "domain/order", Target: "adapter/http", Severity: "error"},
	}
	f := &TerminalFormatter{NoColor: true}
	out, err := f.Format(result)
	if err != nil {
		t.Fatalf("format error: %v", err)
	}
	if !strings.Contains(out, "Cycle:") {
		t.Fatalf("expected cycle output, got: %s", out)
	}
	if !strings.Contains(out, "[ERROR]") {
		t.Fatalf("expected violation output, got: %s", out)
	}
}

func TestMarkdownFormatter_WithLibraries(t *testing.T) {
	result := sampleResult()
	result.Framework.Libraries = []provider.LibraryVersion{
		{Name: "slog", Version: "1.0.0"},
		{Name: "chi", Version: "5.0.0"},
	}
	f := &MarkdownFormatter{}
	out, err := f.Format(result)
	if err != nil {
		t.Fatalf("format error: %v", err)
	}
	if !strings.Contains(out, "| slog | 1.0.0 |") {
		t.Fatalf("expected library table row, got: %s", out)
	}
	if !strings.Contains(out, "| Library | Version |") {
		t.Fatalf("expected table header, got: %s", out)
	}
}

func TestJSONFormatter_ValidJSON(t *testing.T) {
	f := &JSONFormatter{}
	out, err := f.Format(sampleResult())
	if err != nil {
		t.Fatalf("format error: %v", err)
	}
	if !json.Valid([]byte(out)) {
		t.Fatalf("JSON output is not valid: %s", out)
	}
}

func TestNewFormatter_UnsupportedFormat(t *testing.T) {
	_, err := NewFormatter("xml", false)
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewFormatter_EmptyDefaultsToTerminal(t *testing.T) {
	f, err := NewFormatter("", false)
	if err != nil {
		t.Fatalf("NewFormatter empty: %v", err)
	}
	if _, ok := f.(*TerminalFormatter); !ok {
		t.Fatalf("expected TerminalFormatter, got %T", f)
	}
}
