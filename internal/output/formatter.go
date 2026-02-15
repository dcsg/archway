package output

import (
	"fmt"

	"github.com/dcsg/archway/internal/provider"
)

type Formatter interface {
	Format(result *provider.AnalyzeResponse) (string, error)
}

func NewFormatter(format string, noColor bool) (Formatter, error) {
	switch format {
	case "", "terminal":
		return &TerminalFormatter{NoColor: noColor}, nil
	case "json":
		return &JSONFormatter{}, nil
	case "markdown":
		return &MarkdownFormatter{}, nil
	default:
		return nil, fmt.Errorf("unsupported format %q", format)
	}
}
