package guide

import (
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStripTemplateDirectives_RemovesBlockDirectives(t *testing.T) {
	input := `package handler
{{ if .HasHTTP }}
import "net/http"
{{ end }}

func Hello() {}
`
	result := StripTemplateDirectives(input)
	assert.NotContains(t, result, "{{ if")
	assert.NotContains(t, result, "{{ end")
	assert.Contains(t, result, `import "net/http"`)
	assert.Contains(t, result, "func Hello() {}")
}

func TestStripTemplateDirectives_ReplacesInlineVars(t *testing.T) {
	input := `package {{ .ServiceName }}

const name = "{{ .ModulePath }}"
`
	result := StripTemplateDirectives(input)
	assert.Contains(t, result, "<ServiceName>")
	assert.Contains(t, result, "<ModulePath>")
	assert.NotContains(t, result, "{{")
}

func TestStripTemplateDirectives_RemovesRange(t *testing.T) {
	input := `{{ range .Items }}
item
{{ end }}
`
	result := StripTemplateDirectives(input)
	assert.NotContains(t, result, "{{ range")
	assert.Contains(t, result, "item")
}

func TestStripTemplateDirectives_CollapsesBlankLines(t *testing.T) {
	input := "line1\n\n\n\n\nline2\n"
	result := StripTemplateDirectives(input)
	// Should have at most 2 consecutive newlines.
	assert.NotContains(t, result, "\n\n\n")
	assert.Contains(t, result, "line1")
	assert.Contains(t, result, "line2")
}

func TestExtractPatterns_NilFS(t *testing.T) {
	result := ExtractPatterns(nil, []string{"http-api"})
	assert.Empty(t, result)
}

func TestExtractPatterns_NoCaps(t *testing.T) {
	fs := fstest.MapFS{}
	result := ExtractPatterns(fs, nil)
	assert.Empty(t, result)
}

func TestExtractPatterns_UnknownCapability(t *testing.T) {
	fs := fstest.MapFS{}
	result := ExtractPatterns(fs, []string{"unknown-thing"})
	assert.Empty(t, result)
}

func TestExtractPatterns_ExtractsHTTPPattern(t *testing.T) {
	tmplContent := `package httphandler

{{ if .HasAuth }}
import "auth"
{{ end }}

// Handler implements HTTP handlers.
type Handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{logger: logger}
}
`
	fs := fstest.MapFS{
		"templates/capabilities/http-api/files/adapter/httphandler/handler.go.tmpl": &fstest.MapFile{
			Data: []byte(tmplContent),
		},
	}

	result := ExtractPatterns(fs, []string{"http-api"})
	require.NotEmpty(t, result)
	assert.Contains(t, result, "## Patterns")
	assert.Contains(t, result, "HTTP Handler Pattern")
	assert.Contains(t, result, "type Handler struct")
	assert.Contains(t, result, "func NewHandler")
	assert.NotContains(t, result, "{{ if")
	assert.Contains(t, result, "```go")
}

func TestExtractPatterns_MultipleCapabilities(t *testing.T) {
	fs := fstest.MapFS{
		"templates/capabilities/http-api/files/adapter/httphandler/handler.go.tmpl": &fstest.MapFile{
			Data: []byte("package httphandler\n\ntype Handler struct{}\n"),
		},
		"templates/capabilities/mysql/files/adapter/mysqlrepo/connection.go.tmpl": &fstest.MapFile{
			Data: []byte("package mysqlrepo\n\nfunc NewConnection() {}\n"),
		},
	}

	result := ExtractPatterns(fs, []string{"http-api", "mysql"})
	assert.Contains(t, result, "HTTP Handler Pattern")
	assert.Contains(t, result, "MySQL Repository Pattern")
}

func TestExtractPatterns_BFFGateway(t *testing.T) {
	tmplContent := `package bffgateway

// Gateway aggregates responses from multiple backend services.
type Gateway struct {
	clients map[string]ServiceClient
}

func NewGateway(clients map[string]ServiceClient) *Gateway {
	return &Gateway{clients: clients}
}
`
	fs := fstest.MapFS{
		"templates/capabilities/bff/files/adapter/bffgateway/gateway.go.tmpl": &fstest.MapFile{
			Data: []byte(tmplContent),
		},
	}

	result := ExtractPatterns(fs, []string{"bff"})
	require.NotEmpty(t, result)
	assert.Contains(t, result, "BFF Gateway Pattern")
	assert.Contains(t, result, "type Gateway struct")
	assert.Contains(t, result, "func NewGateway")
	assert.Contains(t, result, "```go")
}

func TestCapabilityToTemplateMapping(t *testing.T) {
	// Verify all mapped capabilities have labels.
	for cap := range capabilityTemplateMap {
		label, ok := patternLabel[cap]
		assert.True(t, ok, "missing label for capability %q", cap)
		assert.NotEmpty(t, label)
	}
}
