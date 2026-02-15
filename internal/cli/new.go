package cli

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/dcsg/archway/internal/provider"
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
  archway new my-service --template go-hexagonal --module github.com/acme/orders --no-wizard
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
	language := strings.TrimSpace(opts.Language)
	if language == "" {
		language = "go"
	}
	providerImpl, err := provider.Get(language)
	if err != nil {
		return err
	}

	if !opts.NoWizard {
		if err := runNewWizard(opts); err != nil {
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
		opts.Template = "go-hexagonal"
	}
	if strings.TrimSpace(opts.OutputDir) == "" {
		opts.OutputDir = "."
	}

	request := provider.ScaffoldRequest{
		ProjectName:  opts.Name,
		ModulePath:   opts.ModulePath,
		TemplateName: opts.Template,
		OutputDir:    filepath.Clean(filepath.Join(opts.OutputDir, opts.Name)),
		Options:      map[string]string{},
	}
	for _, kv := range opts.Sets {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid --set %q (expected key=value)", kv)
		}
		request.Options[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
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

func runNewWizard(opts *newCommandOptions) error {
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
		return err
	}

	// Step 2: Language-specific fields — template, module path, etc.
	providerImpl, err := provider.Get(opts.Language)
	if err != nil {
		return err
	}
	info, err := providerImpl.GetInfo(context.Background())
	if err != nil {
		return err
	}

	templateOptions := make([]huh.Option[string], 0, len(info.Templates))
	for _, t := range info.Templates {
		templateOptions = append(templateOptions, huh.NewOption(fmt.Sprintf("%s — %s", t.Name, t.Description), t.Name))
	}
	if opts.Template == "" && len(info.Templates) > 0 {
		opts.Template = info.Templates[0].Name
	}

	langFields := []huh.Field{
		huh.NewSelect[string]().Title("Template").Value(&opts.Template).Options(templateOptions...),
	}

	// Add module path only for Go.
	if opts.Language == "go" {
		langFields = append(langFields,
			huh.NewInput().Title("Go module path").
				Description("e.g. github.com/org/service (leave empty for example.com/<name>)").
				Value(&opts.ModulePath),
		)
	}

	return huh.NewForm(huh.NewGroup(langFields...)).Run()
}
