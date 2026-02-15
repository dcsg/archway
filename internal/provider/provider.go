package provider

import (
	"context"
	"errors"
	"io/fs"
)

var ErrNotImplemented = errors.New("not yet implemented")

type LanguageProvider interface {
	Scaffold(ctx context.Context, req ScaffoldRequest) (*ScaffoldResponse, error)
	Analyze(ctx context.Context, req AnalyzeRequest) (*AnalyzeResponse, error)
	Migrate(ctx context.Context, req MigrateRequest) (*MigrateResponse, error)
	GetInfo(ctx context.Context) (*ProviderInfo, error)
	// GetTemplateFS returns the embedded filesystem for template files.
	// Used by the CLI wizard to load wizard.yaml and manifest.yaml.
	GetTemplateFS() fs.FS
}

type ScaffoldRequest struct {
	ProjectName  string            `json:"project_name"`
	ModulePath   string            `json:"module_path"`
	TemplateName string            `json:"template_name"`
	Options      map[string]string `json:"options,omitempty"`
	OutputDir    string            `json:"output_dir"`
}

type ScaffoldResponse struct {
	FilesCreated []string `json:"files_created"`
	ArchwayYAML  []byte   `json:"archway_yaml,omitempty"`
}

type AnalyzeRequest struct {
	Path       string `json:"path"`
	IncludeLLM bool   `json:"include_llm"`
}

type AnalyzeResponse struct {
	Language        string             `json:"language"`
	Architecture    ArchitectureResult `json:"architecture"`
	Framework       FrameworkResult    `json:"framework"`
	Conventions     ConventionResults  `json:"conventions"`
	DependencyGraph DependencyGraph    `json:"dependency_graph"`
	Violations      []Violation        `json:"violations,omitempty"`
	PackageCount    int                `json:"package_count"`
	FileCount       int                `json:"file_count"`
	FunctionCount   int                `json:"function_count"`
	Metadata        map[string]string  `json:"metadata,omitempty"`
	LLM             *LLMEnhancement    `json:"llm,omitempty"`
}

type LLMEnhancement struct {
	ADRs               []ADR    `json:"adrs,omitempty"`
	Invariants         []string `json:"invariants,omitempty"`
	SemanticAssessment string   `json:"semantic_assessment,omitempty"`
	Provider           string   `json:"provider,omitempty"`
	Model              string   `json:"model,omitempty"`
	TokensUsed         int      `json:"tokens_used,omitempty"`
	Warnings           []string `json:"warnings,omitempty"`
}

type ADR struct {
	Title        string `json:"title"`
	Context      string `json:"context"`
	Decision     string `json:"decision"`
	Consequences string `json:"consequences"`
}

type ArchitectureResult struct {
	Pattern    string   `json:"pattern"`
	Confidence float64  `json:"confidence"`
	Evidence   []string `json:"evidence,omitempty"`
}

type FrameworkResult struct {
	Name       string           `json:"name"`
	Version    string           `json:"version,omitempty"`
	Confidence float64          `json:"confidence"`
	Libraries  []LibraryVersion `json:"libraries,omitempty"`
}

type LibraryVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type ConventionResults struct {
	ErrorHandling ConventionFinding `json:"error_handling"`
	Logging       ConventionFinding `json:"logging"`
	Config        ConventionFinding `json:"config"`
	Testing       TestingFinding    `json:"testing"`
}

type ConventionFinding struct {
	Pattern    string   `json:"pattern"`
	Confidence float64  `json:"confidence"`
	Evidence   []string `json:"evidence,omitempty"`
}

type TestingFinding struct {
	Pattern      string   `json:"pattern"`
	Confidence   float64  `json:"confidence"`
	Evidence     []string `json:"evidence,omitempty"`
	TestFiles    int      `json:"test_files"`
	TotalGoFiles int      `json:"total_go_files"`
}

type DependencyGraph struct {
	Nodes  []PackageNode    `json:"nodes"`
	Edges  []DependencyEdge `json:"edges"`
	Cycles [][]string       `json:"cycles,omitempty"`
}

type PackageNode struct {
	Path       string `json:"path"`
	Name       string `json:"name"`
	IsInternal bool   `json:"is_internal"`
	Layer      string `json:"layer,omitempty"`
}

type DependencyEdge struct {
	From       string `json:"from"`
	To         string `json:"to"`
	ImportType string `json:"import_type,omitempty"`
}

type Violation struct {
	Rule     string `json:"rule"`
	Message  string `json:"message"`
	Source   string `json:"source"`
	Target   string `json:"target,omitempty"`
	Severity string `json:"severity"`
}

type MigrateRequest struct {
	Path     string `json:"path"`
	Strategy string `json:"strategy"`
}

type MigrateResponse struct {
	Success bool     `json:"success"`
	Changes []string `json:"changes,omitempty"`
}

type ProviderInfo struct {
	Name                   string         `json:"name"`
	Version                string         `json:"version"`
	Language               string         `json:"language"`
	SupportedArchitectures []string       `json:"supported_architectures"`
	Templates              []TemplateInfo `json:"templates"`
}

type TemplateInfo struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Variables   []VariableInfo `json:"variables,omitempty"`
}

type VariableInfo struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Default     string   `json:"default,omitempty"`
	Required    bool     `json:"required"`
	Choices     []string `json:"choices,omitempty"`
}
