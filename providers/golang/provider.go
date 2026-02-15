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
		templateName = "go-hexagonal"
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

	renderer := scaffold.NewRenderer(templatesFS)
	templateDir := path.Join("templates", templateName)
	renderResult, err := renderer.RenderTemplate(templateDir, req.OutputDir, vars)
	if err != nil {
		return nil, err
	}

	manifest, err := loadManifest(templateDir)
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

	architecture := "layered"
	if strings.Contains(templateName, "hexagonal") {
		architecture = "hexagonal"
	}
	archwayCfg := config.DefaultArchwayConfig("go", architecture)
	archwayPath := filepath.Join(req.OutputDir, "archway.yaml")
	if err := config.SaveArchwayYAML(archwayPath, archwayCfg); err != nil {
		return nil, err
	}
	archwayBytes, _ := os.ReadFile(archwayPath)

	files := append([]string{}, renderResult.FilesCreated...)
	files = append(files, archwayPath)

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

func loadManifest(templateDir string) (*scaffold.Manifest, error) {
	data, err := fs.ReadFile(templatesFS, path.Join(templateDir, "manifest.yaml"))
	if err != nil {
		return nil, err
	}
	return scaffold.ParseManifest(data)
}

func listTemplates() ([]provider.TemplateInfo, error) {
	entries, err := fs.ReadDir(templatesFS, "templates")
	if err != nil {
		return nil, err
	}
	infos := []provider.TemplateInfo{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		manifest, err := loadManifest(path.Join("templates", entry.Name()))
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
