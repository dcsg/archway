package output

import (
	"encoding/json"
	"fmt"

	"github.com/dcsg/archway/internal/provider"
)

type JSONFormatter struct{}

func (f *JSONFormatter) Format(result *provider.AnalyzeResponse) (string, error) {
	if result == nil {
		return "", fmt.Errorf("analysis result is nil")
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
