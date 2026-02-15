package output

import (
	"fmt"
	"strings"

	"github.com/dcsg/archway/internal/provider"
)

type TerminalFormatter struct {
	NoColor bool
}

func (f *TerminalFormatter) Format(result *provider.AnalyzeResponse) (string, error) {
	if result == nil {
		return "", fmt.Errorf("analysis result is nil")
	}
	b := &strings.Builder{}
	writeHeading(b, "Project Summary", !f.NoColor)
	fmt.Fprintf(b, "Language: %s\nPackages: %d\nFiles: %d\nFunctions: %d\n\n",
		result.Language, result.PackageCount, result.FileCount, result.FunctionCount)

	writeHeading(b, "Architecture", !f.NoColor)
	fmt.Fprintf(b, "Pattern: %s (confidence %.2f)\n", result.Architecture.Pattern, result.Architecture.Confidence)
	for _, evidence := range result.Architecture.Evidence {
		fmt.Fprintf(b, "  • %s\n", evidence)
	}
	b.WriteString("\n")

	writeHeading(b, "Framework", !f.NoColor)
	fmt.Fprintf(b, "Primary: %s %s (confidence %.2f)\n", result.Framework.Name, result.Framework.Version, result.Framework.Confidence)
	if len(result.Framework.Libraries) > 0 {
		b.WriteString("Libraries:\n")
		for _, lib := range result.Framework.Libraries {
			fmt.Fprintf(b, "  • %s %s\n", lib.Name, lib.Version)
		}
	}
	b.WriteString("\n")

	writeHeading(b, "Conventions", !f.NoColor)
	fmt.Fprintf(b, "Error handling: %s\n", result.Conventions.ErrorHandling.Pattern)
	fmt.Fprintf(b, "Logging: %s\n", result.Conventions.Logging.Pattern)
	fmt.Fprintf(b, "Config: %s\n", result.Conventions.Config.Pattern)
	fmt.Fprintf(b, "Testing: %s (test files %d/%d)\n\n",
		result.Conventions.Testing.Pattern,
		result.Conventions.Testing.TestFiles,
		result.Conventions.Testing.TotalGoFiles,
	)

	writeHeading(b, "Dependency Issues", !f.NoColor)
	if len(result.Violations) == 0 && len(result.DependencyGraph.Cycles) == 0 {
		b.WriteString("No dependency issues found.\n")
	} else {
		for _, cycle := range result.DependencyGraph.Cycles {
			fmt.Fprintf(b, "Cycle: %s\n", strings.Join(cycle, " -> "))
		}
		for _, violation := range result.Violations {
			fmt.Fprintf(b, "[%s] %s (%s -> %s)\n", strings.ToUpper(violation.Severity), violation.Message, violation.Source, violation.Target)
		}
	}
	return b.String(), nil
}

func writeHeading(b *strings.Builder, text string, color bool) {
	header := fmt.Sprintf("┌─ %s", text)
	if color {
		header = "\033[1m" + header + "\033[0m"
	}
	b.WriteString(header)
	b.WriteString("\n")
}
