package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	gomcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/dcsg/archway/internal/analyzer/detector"
	"github.com/dcsg/archway/internal/output"
	"github.com/dcsg/archway/internal/provider"
	_ "github.com/dcsg/archway/providers/golang"
)

type analyzeInput struct {
	Path         string `json:"path" jsonschema:"Path to the project to analyze"`
	OutputFormat string `json:"output_format" jsonschema:"Output format: json or markdown"`
}

type detectArchInput struct {
	Path string `json:"path" jsonschema:"Path to the project"`
}

type listTemplatesInput struct {
	Language string `json:"language" jsonschema:"Language to filter templates by"`
}

type scaffoldInput struct {
	Template   string            `json:"template" jsonschema:"Template name"`
	Name       string            `json:"name" jsonschema:"Project/service name"`
	ModulePath string            `json:"module_path" jsonschema:"Go module path"`
	OutputDir  string            `json:"output_dir" jsonschema:"Output directory"`
	Options    map[string]string `json:"options" jsonschema:"Template variables"`
}

func (s *Server) registerTools() {
	gomcp.AddTool(s.server, &gomcp.Tool{
		Name:        "analyze_codebase",
		Description: "Analyze a Go codebase to detect architecture, frameworks, conventions, and dependency structure.",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, in analyzeInput) (*gomcp.CallToolResult, any, error) {
		return s.handleAnalyzeCodebase(ctx, in)
	})

	gomcp.AddTool(s.server, &gomcp.Tool{
		Name:        "detect_architecture",
		Description: "Detect the architecture pattern of a Go project (hexagonal, clean, DDD, layered, flat).",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, in detectArchInput) (*gomcp.CallToolResult, any, error) {
		return s.handleDetectArchitecture(ctx, in)
	})

	gomcp.AddTool(s.server, &gomcp.Tool{
		Name:        "list_templates",
		Description: "List available scaffold templates for a language.",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, in listTemplatesInput) (*gomcp.CallToolResult, any, error) {
		return s.handleListTemplates(ctx, in)
	})

	gomcp.AddTool(s.server, &gomcp.Tool{
		Name:        "scaffold_project",
		Description: "Scaffold a new project from a template.",
	}, func(ctx context.Context, req *gomcp.CallToolRequest, in scaffoldInput) (*gomcp.CallToolResult, any, error) {
		return s.handleScaffoldProject(ctx, in)
	})
}

func (s *Server) handleAnalyzeCodebase(ctx context.Context, in analyzeInput) (*gomcp.CallToolResult, any, error) {
	path := strings.TrimSpace(in.Path)
	if path == "" {
		path = "."
	}
	language, _, err := detector.DetectLanguage(path)
	if err != nil {
		return nil, nil, err
	}
	if language == "unknown" {
		language = "go"
	}
	p, err := provider.Get(language)
	if err != nil {
		return nil, nil, err
	}
	result, err := p.Analyze(ctx, provider.AnalyzeRequest{Path: path})
	if err != nil {
		return nil, nil, err
	}
	s.setLatestAnalysis(result)

	format := strings.TrimSpace(in.OutputFormat)
	if format == "" || format == "json" {
		data, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return nil, nil, err
		}
		return &gomcp.CallToolResult{
			Content: []gomcp.Content{&gomcp.TextContent{Text: string(data)}},
		}, nil, nil
	}
	formatter, err := output.NewFormatter(format, true)
	if err != nil {
		return nil, nil, err
	}
	formatted, err := formatter.Format(result)
	if err != nil {
		return nil, nil, err
	}
	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: formatted}},
	}, nil, nil
}

func (s *Server) handleDetectArchitecture(ctx context.Context, in detectArchInput) (*gomcp.CallToolResult, any, error) {
	path := strings.TrimSpace(in.Path)
	if path == "" {
		path = "."
	}
	analyzeIn := analyzeInput{Path: path, OutputFormat: "json"}
	result, _, err := s.handleAnalyzeCodebase(ctx, analyzeIn)
	if err != nil {
		return nil, nil, err
	}

	latest := s.getLatestAnalysis()
	if latest == nil {
		return nil, nil, fmt.Errorf("analysis not available")
	}
	data, err := json.MarshalIndent(latest.Architecture, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	_ = result // suppress unused; analysis stored via side-effect
	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func (s *Server) handleListTemplates(ctx context.Context, in listTemplatesInput) (*gomcp.CallToolResult, any, error) {
	language := strings.TrimSpace(in.Language)
	if language == "" {
		language = "go"
	}
	p, err := provider.Get(language)
	if err != nil {
		return nil, nil, err
	}
	info, err := p.GetInfo(ctx)
	if err != nil {
		return nil, nil, err
	}
	data, err := json.MarshalIndent(info.Templates, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: string(data)}},
	}, nil, nil
}

func (s *Server) handleScaffoldProject(ctx context.Context, in scaffoldInput) (*gomcp.CallToolResult, any, error) {
	templateName := strings.TrimSpace(in.Template)
	if templateName == "" {
		templateName = "go-hexagonal"
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		name = "service"
	}
	modulePath := strings.TrimSpace(in.ModulePath)
	if modulePath == "" {
		modulePath = "example.com/" + name
	}
	outputDir := strings.TrimSpace(in.OutputDir)
	if outputDir == "" {
		outputDir = filepath.Join(".", name)
	}

	options := in.Options
	if options == nil {
		options = map[string]string{}
	}

	p, err := provider.Get("go")
	if err != nil {
		return nil, nil, err
	}
	resp, err := p.Scaffold(ctx, provider.ScaffoldRequest{
		ProjectName:  name,
		ModulePath:   modulePath,
		TemplateName: templateName,
		Options:      options,
		OutputDir:    outputDir,
	})
	if err != nil {
		return nil, nil, err
	}
	data, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return nil, nil, err
	}
	return &gomcp.CallToolResult{
		Content: []gomcp.Content{&gomcp.TextContent{Text: string(data)}},
	}, nil, nil
}
