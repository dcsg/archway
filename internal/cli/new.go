package cli

import (
	"context"
	"fmt"
	"io/fs"
	"path"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/dcsg/archway/internal/provider"
	"github.com/dcsg/archway/internal/scaffold"
	_ "github.com/dcsg/archway/providers/golang"
	"github.com/spf13/cobra"
)

type newCommandOptions struct {
	Name       string
	Language   string
	Template   string
	NoWizard   bool
	ModulePath string
	OutputDir  string
	Sets       []string
}

func newNewCommand(_ *globalOptions) *cobra.Command {
	opts := &newCommandOptions{}

	cmd := &cobra.Command{
		Use:   "new [name]",
		Short: "Scaffold a new project",
		Long: `Create a new project scaffold from a template.

This command can run interactively through a wizard or non-interactively using flags.`,
		Example: `  archway new my-service
  archway new my-service --template api --module github.com/acme/orders --no-wizard
  archway new --name orders --set HasGRPC=true --set HasPostgreSQL=true`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 && strings.TrimSpace(opts.Name) == "" {
				opts.Name = args[0]
			}
			if opts.NoWizard && strings.TrimSpace(opts.Name) == "" {
				return fmt.Errorf("name is required: archway new <name> or --name <name>")
			}
			return runNew(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Project/service name")
	cmd.Flags().StringVar(&opts.Language, "language", "", "Project language (defaults to go)")
	cmd.Flags().StringVar(&opts.Template, "template", "", "Template name")
	cmd.Flags().BoolVar(&opts.NoWizard, "no-wizard", false, "Disable interactive wizard")
	cmd.Flags().StringVar(&opts.ModulePath, "module", "", "Go module path")
	cmd.Flags().StringVar(&opts.OutputDir, "output-dir", ".", "Output directory")
	cmd.Flags().StringArrayVar(&opts.Sets, "set", nil, "Template variable assignment (key=value), repeatable")

	return cmd
}

func runNew(ctx context.Context, opts *newCommandOptions) error {
	if strings.TrimSpace(opts.Language) == "" {
		opts.Language = "go"
	}

	// Wizard mode: interactive prompts for all options.
	wizardVars := map[string]interface{}{}
	if !opts.NoWizard {
		var err error
		wizardVars, err = runNewWizard(opts)
		if err != nil {
			return err
		}
	}

	if strings.TrimSpace(opts.Name) == "" {
		return fmt.Errorf("project name is required")
	}
	if strings.TrimSpace(opts.ModulePath) == "" {
		opts.ModulePath = fmt.Sprintf("example.com/%s", opts.Name)
	}
	if strings.TrimSpace(opts.Template) == "" {
		opts.Template = "api"
	}
	if strings.TrimSpace(opts.OutputDir) == "" {
		opts.OutputDir = "."
	}

	providerImpl, err := provider.Get(opts.Language)
	if err != nil {
		return err
	}

	// Build options: wizard vars first, then --set overrides on top.
	options := map[string]string{}
	for k, v := range wizardVars {
		options[k] = fmt.Sprint(v)
	}
	for _, kv := range opts.Sets {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid --set %q (expected key=value)", kv)
		}
		options[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	request := provider.ScaffoldRequest{
		ProjectName:  opts.Name,
		ModulePath:   opts.ModulePath,
		TemplateName: opts.Template,
		OutputDir:    filepath.Clean(filepath.Join(opts.OutputDir, opts.Name)),
		Options:      options,
	}

	resp, err := providerImpl.Scaffold(ctx, request)
	if err != nil {
		return err
	}
	fmt.Printf("Scaffold complete: %d files created\n", len(resp.FilesCreated))
	for _, file := range resp.FilesCreated {
		fmt.Printf("  - %s\n", file)
	}
	return nil
}

func runNewWizard(opts *newCommandOptions) (map[string]interface{}, error) {
	// Step 1: Common fields — name, output dir, language.
	languages := provider.List()
	languageOptions := make([]huh.Option[string], 0, len(languages))
	for _, lang := range languages {
		languageOptions = append(languageOptions, huh.NewOption(lang, lang))
	}
	if opts.Language == "" {
		opts.Language = "go"
	}

	commonFields := []huh.Field{
		huh.NewInput().Title("Service name").Value(&opts.Name).Validate(func(value string) error {
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("service name is required")
			}
			return nil
		}),
		huh.NewInput().Title("Output directory").
			Description("Parent directory where the project folder will be created").
			Placeholder(".").
			Value(&opts.OutputDir),
	}
	if len(languageOptions) > 1 {
		commonFields = append(commonFields,
			huh.NewSelect[string]().Title("Language").Value(&opts.Language).Options(languageOptions...),
		)
	}

	if err := huh.NewForm(huh.NewGroup(commonFields...)).Run(); err != nil {
		return nil, err
	}

	// Step 2: Load provider and attempt intent-driven wizard.
	providerImpl, err := provider.Get(opts.Language)
	if err != nil {
		return nil, err
	}
	tFS := providerImpl.GetTemplateFS()

	// Try loading the provider-level wizard.yaml for intent routing.
	providerWizardData, provWizErr := fs.ReadFile(tFS, path.Join("templates", "wizard.yaml"))
	if provWizErr != nil {
		// No provider wizard — fall back to manual template selection.
		return runLegacyTemplateWizard(opts, providerImpl)
	}

	providerWizardCfg, err := scaffold.ParseProviderWizard(providerWizardData)
	if err != nil {
		return nil, fmt.Errorf("parse provider wizard: %w", err)
	}

	// Step 3: Run provider wizard — determines intent and template.
	templateName, intentState, err := scaffold.RunProviderWizard(providerWizardCfg, nil)
	if err != nil {
		return nil, err
	}
	opts.Template = templateName

	// Step 4: Resolve fast path for template wizard step skipping.
	fastPath := providerWizardCfg.ResolveFastPath(intentState)

	// Step 5: Load the resolved template's manifest + wizard.
	templateDir := path.Join("templates", opts.Template)

	manifestData, err := fs.ReadFile(tFS, path.Join(templateDir, "manifest.yaml"))
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	manifest, err := scaffold.ParseManifest(manifestData)
	if err != nil {
		return nil, err
	}

	// Pre-seed with name and module path from step 1.
	initial := map[string]interface{}{}
	for k, v := range intentState {
		initial[k] = v
	}
	if opts.Name != "" {
		initial["ServiceName"] = opts.Name
	}
	modulePath := opts.ModulePath
	if modulePath == "" && opts.Name != "" {
		modulePath = fmt.Sprintf("github.com/example/%s", opts.Name)
	}
	if modulePath != "" {
		initial["ModulePath"] = modulePath
	}

	wizardData, err := fs.ReadFile(tFS, path.Join(templateDir, "wizard.yaml"))
	if err != nil {
		// No template wizard — return manifest defaults merged with intent state.
		defaults := manifest.Defaults()
		for k, v := range initial {
			defaults[k] = v
		}
		return defaults, nil
	}
	wizardCfg, err := scaffold.ParseWizard(wizardData)
	if err != nil {
		return nil, err
	}

	// Step 6: Run template wizard with fast-path filtering.
	var vars map[string]interface{}
	if fastPath != nil {
		vars, err = scaffold.RunWizardWithFastPath(wizardCfg, manifest, initial, fastPath)
	} else {
		vars, err = scaffold.RunWizard(wizardCfg, manifest, initial)
	}
	if err != nil {
		return nil, err
	}

	// Copy wizard results back to opts for fields it manages.
	if name, ok := vars["ServiceName"].(string); ok && name != "" {
		opts.Name = name
	}
	if mod, ok := vars["ModulePath"].(string); ok && mod != "" {
		opts.ModulePath = mod
	}

	return vars, nil
}

// runLegacyTemplateWizard falls back to manual template selection when
// no provider-level wizard.yaml exists.
func runLegacyTemplateWizard(opts *newCommandOptions, providerImpl provider.LanguageProvider) (map[string]interface{}, error) {
	info, err := providerImpl.GetInfo(context.Background())
	if err != nil {
		return nil, err
	}

	templateOptions := make([]huh.Option[string], 0, len(info.Templates))
	for _, t := range info.Templates {
		templateOptions = append(templateOptions, huh.NewOption(fmt.Sprintf("%s — %s", t.Name, t.Description), t.Name))
	}
	if opts.Template == "" && len(info.Templates) > 0 {
		opts.Template = info.Templates[0].Name
	}

	if err := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().Title("Template").Value(&opts.Template).Options(templateOptions...),
	)).Run(); err != nil {
		return nil, err
	}

	tFS := providerImpl.GetTemplateFS()
	templateDir := path.Join("templates", opts.Template)

	manifestData, err := fs.ReadFile(tFS, path.Join(templateDir, "manifest.yaml"))
	if err != nil {
		return nil, fmt.Errorf("read manifest: %w", err)
	}
	manifest, err := scaffold.ParseManifest(manifestData)
	if err != nil {
		return nil, err
	}

	wizardData, err := fs.ReadFile(tFS, path.Join(templateDir, "wizard.yaml"))
	if err != nil {
		return manifest.Defaults(), nil
	}
	wizardCfg, err := scaffold.ParseWizard(wizardData)
	if err != nil {
		return nil, err
	}

	initial := map[string]interface{}{}
	if opts.Name != "" {
		initial["ServiceName"] = opts.Name
	}
	modulePath := opts.ModulePath
	if modulePath == "" && opts.Name != "" {
		modulePath = fmt.Sprintf("github.com/example/%s", opts.Name)
	}
	if modulePath != "" {
		initial["ModulePath"] = modulePath
	}
	vars, err := scaffold.RunWizard(wizardCfg, manifest, initial)
	if err != nil {
		return nil, err
	}

	if name, ok := vars["ServiceName"].(string); ok && name != "" {
		opts.Name = name
	}
	if mod, ok := vars["ModulePath"].(string); ok && mod != "" {
		opts.ModulePath = mod
	}

	return vars, nil
}
