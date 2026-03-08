package output

import (
	"fmt"
	"strings"

	"github.com/dcsg/archway/internal/provider"
)

type MarkdownFormatter struct{}

func (f *MarkdownFormatter) Format(result *provider.AnalyzeResponse) (string, error) {
	if result == nil {
		return "", fmt.Errorf("analysis result is nil")
	}
	b := &strings.Builder{}
	fmt.Fprintf(b, "# Archway Analysis\n\n")
	fmt.Fprintf(b, "## Project Summary\n\n")
	fmt.Fprintf(b, "- Language: `%s`\n- Packages: `%d`\n- Files: `%d`\n- Functions: `%d`\n\n",
		result.Language, result.PackageCount, result.FileCount, result.FunctionCount)

	fmt.Fprintf(b, "## Architecture\n\n")
	fmt.Fprintf(b, "Pattern: **%s** (confidence `%.2f`)\n\n", result.Architecture.Pattern, result.Architecture.Confidence)
	if len(result.Architecture.Evidence) > 0 {
		b.WriteString("Evidence:\n")
		for _, evidence := range result.Architecture.Evidence {
			fmt.Fprintf(b, "- %s\n", evidence)
		}
		b.WriteString("\n")
	}

	fmt.Fprintf(b, "## Framework\n\n")
	fmt.Fprintf(b, "Primary: **%s** `%s`\n\n", result.Framework.Name, result.Framework.Version)
	if len(result.Framework.Libraries) > 0 {
		b.WriteString("| Library | Version |\n|---|---|\n")
		for _, lib := range result.Framework.Libraries {
			fmt.Fprintf(b, "| %s | %s |\n", lib.Name, lib.Version)
		}
		b.WriteString("\n")
	}

	fmt.Fprintf(b, "## Conventions\n\n")
	fmt.Fprintf(b, "- Error handling: `%s`\n", result.Conventions.ErrorHandling.Pattern)
	fmt.Fprintf(b, "- Logging: `%s`\n", result.Conventions.Logging.Pattern)
	fmt.Fprintf(b, "- Config: `%s`\n", result.Conventions.Config.Pattern)
	fmt.Fprintf(b, "- Testing: `%s`\n\n", result.Conventions.Testing.Pattern)

	fmt.Fprintf(b, "## Dependency Issues\n\n")
	if len(result.DependencyGraph.Cycles) == 0 && len(result.Violations) == 0 {
		b.WriteString("No issues found.\n")
	} else {
		for _, cycle := range result.DependencyGraph.Cycles {
			fmt.Fprintf(b, "- Cycle: `%s`\n", strings.Join(cycle, " -> "))
		}
		for _, v := range result.Violations {
			fmt.Fprintf(b, "- %s\n", v.Message)
		}
	}

	return b.String(), nil
}
