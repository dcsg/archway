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
		Use:   "new",
		Short: "Scaffold a new project",
		Long: `Create a new project scaffold from a template.

This command can run interactively through a wizard or non-interactively using flags.`,
		Example: `  archway new
  archway new --name orders --language go --template go-hexagonal --module github.com/acme/orders --no-wizard
  archway new --name orders --set Transport=http --set DataStore=postgres`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if opts.NoWizard && strings.TrimSpace(opts.Name) == "" {
				return fmt.Errorf("--name is required when --no-wizard is set")
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
		OutputDir:    filepath.Join(opts.OutputDir, opts.Name),
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
	template := opts.Template
	if template == "" {
		template = "go-hexagonal"
	}
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Service name").Value(&opts.Name).Validate(func(value string) error {
				if strings.TrimSpace(value) == "" {
					return fmt.Errorf("service name is required")
				}
				return nil
			}),
			huh.NewInput().Title("Module path").Value(&opts.ModulePath),
			huh.NewSelect[string]().Title("Template").Value(&template).Options(
				huh.NewOption("go-hexagonal", "go-hexagonal"),
				huh.NewOption("go-minimal", "go-minimal"),
			),
		),
	)
	if err := form.Run(); err != nil {
		return err
	}
	opts.Template = template
	return nil
}
