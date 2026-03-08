package golang

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/dcsg/archway/internal/analyzer"
	"github.com/dcsg/archway/internal/config"
	"github.com/dcsg/archway/internal/provider"
	"github.com/dcsg/archway/internal/scaffold"
)

type GoProvider struct{}

func init() {
	provider.Register("go", &GoProvider{})
}

func (p *GoProvider) Scaffold(_ context.Context, req provider.ScaffoldRequest) (*provider.ScaffoldResponse, error) {
	templateName := strings.TrimSpace(req.TemplateName)
	if templateName == "" {
		templateName = "api"
	}
	if req.OutputDir == "" {
		req.OutputDir = "."
	}
	if err := os.MkdirAll(req.OutputDir, 0o755); err != nil {
		return nil, fmt.Errorf("create output dir: %w", err)
	}

	vars := map[string]interface{}{}
	for k, v := range req.Options {
		vars[k] = v
	}
	if req.ProjectName != "" {
		vars["ServiceName"] = req.ProjectName
	}
	if req.ModulePath != "" {
		vars["ModulePath"] = req.ModulePath
	}

	// Map legacy template names to architectures.
	archMap := map[string]string{"api": "hexagonal", "cli": "flat", "worker": "hexagonal"}
	architecture := archMap[templateName]
	if architecture == "" {
		architecture = templateName
	}

	renderer := scaffold.NewRenderer(templatesFS)

	// Parse capabilities from options (comma-separated).
	var capabilities []string
	if capStr := req.Options["capabilities"]; capStr != "" {
		for _, c := range strings.Split(capStr, ",") {
			c = strings.TrimSpace(c)
			if c != "" {
				capabilities = append(capabilities, c)
			}
		}
	}

	var renderResult *scaffold.RenderResult
	if len(capabilities) > 0 {
		// Composition mode: architecture + capabilities.
		plan, err := scaffold.ComposeProject(templatesFS, architecture, capabilities, vars)
		if err != nil {
			return nil, fmt.Errorf("compose project: %w", err)
		}
		renderResult, err = renderer.RenderComposition(plan, req.OutputDir)
		if err != nil {
			return nil, err
		}
	} else {
		// Legacy mode: single template directory.
		archDir := path.Join("templates", "architectures", architecture)
		var err error
		renderResult, err = renderer.RenderTemplate(archDir, req.OutputDir, vars)
		if err != nil {
			return nil, err
		}
	}

	archDir := path.Join("templates", "architectures", architecture)
	manifest, err := loadManifest(archDir)
	if err == nil {
		hooks := manifest.Hooks
		if len(hooks) == 0 {
			hooks = scaffold.DefaultGoHooks()
		}
		if strings.EqualFold(req.Options["skip_hooks"], "true") {
			hooks = nil
		}
		if err := scaffold.RunPostScaffoldHooks(req.OutputDir, hooks, vars); err != nil {
			return nil, err
		}
	}

	archwayCfg := config.DefaultArchwayConfig("go", architecture)
	if len(capabilities) > 0 {
		archwayCfg.Capabilities = capabilities
	}
	archwayPath := filepath.Join(req.OutputDir, "archway.yaml")
	if err := config.SaveArchwayYAML(archwayPath, archwayCfg); err != nil {
		return nil, err
	}
	archwayBytes, _ := os.ReadFile(archwayPath)

	files := append([]string{}, renderResult.FilesCreated...)
	files = append(files, archwayPath)

	// Generate project matrix doc.
	matrixPath, err := generateProjectMatrix(req.OutputDir, architecture, capabilities, archwayCfg)
	if err == nil && matrixPath != "" {
		files = append(files, matrixPath)
	}

	return &provider.ScaffoldResponse{FilesCreated: files, ArchwayYAML: archwayBytes}, nil
}

func (p *GoProvider) Analyze(ctx context.Context, req provider.AnalyzeRequest) (*provider.AnalyzeResponse, error) {
	path := strings.TrimSpace(req.Path)
	if path == "" {
		path = "."
	}
	a := analyzer.New(path)
	if err := a.LoadPackages(""); err != nil {
		return nil, err
	}
	return a.Analyze(ctx)
}

func (p *GoProvider) Migrate(_ context.Context, _ provider.MigrateRequest) (*provider.MigrateResponse, error) {
	return nil, provider.ErrNotImplemented
}

func (p *GoProvider) GetInfo(_ context.Context) (*provider.ProviderInfo, error) {
	templates, err := listTemplates()
	if err != nil {
		return nil, err
	}
	return &provider.ProviderInfo{
		Name:                   "archway-go-provider",
		Version:                "v1",
		Language:               "go",
		SupportedArchitectures: []string{"hexagonal", "clean", "ddd", "layered", "flat"},
		Templates:              templates,
	}, nil
}

func (p *GoProvider) GetTemplateFS() fs.FS {
	return templatesFS
}

func loadManifest(templateDir string) (*scaffold.Manifest, error) {
	data, err := fs.ReadFile(templatesFS, path.Join(templateDir, "manifest.yaml"))
	if err != nil {
		return nil, err
	}
	return scaffold.ParseManifest(data)
}

func listTemplates() ([]provider.TemplateInfo, error) {
	entries, err := fs.ReadDir(templatesFS, path.Join("templates", "architectures"))
	if err != nil {
		return nil, err
	}
	infos := []provider.TemplateInfo{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		manifest, err := loadManifest(path.Join("templates", "architectures", entry.Name()))
		if err != nil {
			continue
		}
		vars := make([]provider.VariableInfo, 0, len(manifest.Variables))
		for _, v := range manifest.Variables {
			vars = append(vars, provider.VariableInfo{
				Name:        v.Name,
				Type:        v.Type,
				Description: v.Description,
				Default:     v.Default,
				Required:    v.Required,
				Choices:     v.Choices,
			})
		}
		infos = append(infos, provider.TemplateInfo{
			Name:        manifest.Name,
			Description: manifest.Description,
			Variables:   vars,
		})
	}
	return infos, nil
}
