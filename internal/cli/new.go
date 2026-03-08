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
	Name         string
	Language     string
	Template     string
	Architecture string
	Capabilities string
	NoWizard     bool
	ModulePath   string
	OutputDir    string
	Sets         []string
}

func newNewCommand(_ *globalOptions) *cobra.Command {
	opts := &newCommandOptions{}

	cmd := &cobra.Command{
		Use:   "new [name]",
		Short: "Scaffold a new project",
		Long: `Create a new project scaffold by composing architecture + capabilities.

This command runs an interactive wizard by default, or can be used non-interactively with flags.`,
		Example: `  archway new my-service
  archway new my-service --arch hexagonal --cap http-api,mysql,observability --no-wizard
  archway new my-service --template cli --no-wizard`,
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
	cmd.Flags().StringVar(&opts.Template, "template", "", "Legacy template name (use --arch instead)")
	cmd.Flags().StringVar(&opts.Architecture, "arch", "", "Architecture pattern (hexagonal, flat)")
	cmd.Flags().StringVar(&opts.Capabilities, "cap", "", "Capabilities (comma-separated: http-api,mysql,redis)")
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

	wizardVars := map[string]interface{}{}
	if !opts.NoWizard {
		var err error
		wizardVars, err = runCompositionWizard(opts)
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
	if strings.TrimSpace(opts.OutputDir) == "" {
		opts.OutputDir = "."
	}

	// Resolve template from architecture flag.
	template := opts.Template
	if template == "" && opts.Architecture != "" {
		template = opts.Architecture
	}
	if template == "" {
		template = "api"
	}

	providerImpl, err := provider.Get(opts.Language)
	if err != nil {
		return err
	}

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

	// Pass capabilities to the provider.
	if opts.Capabilities != "" {
		options["capabilities"] = opts.Capabilities
	}

	request := provider.ScaffoldRequest{
		ProjectName:  opts.Name,
		ModulePath:   opts.ModulePath,
		TemplateName: template,
		OutputDir:    filepath.Clean(filepath.Join(opts.OutputDir, opts.Name)),
		Options:      options,
	}

	resp, err := providerImpl.Scaffold(ctx, request)
	if err != nil {
		return err
	}

	fmt.Printf("\nScaffold complete: %d files created\n", len(resp.FilesCreated))
	for _, file := range resp.FilesCreated {
		fmt.Printf("  %s\n", file)
	}

	// Print equivalent non-interactive command.
	printEquivalentCommand(opts)

	fmt.Println("\nNext steps:")
	fmt.Printf("  cd %s\n", filepath.Join(opts.OutputDir, opts.Name))
	fmt.Printf("  go run ./cmd/%s\n", opts.Name)
	fmt.Println("\nTip: Run 'keel:init' to set up AI-powered development guardrails.")

	return nil
}

func printEquivalentCommand(opts *newCommandOptions) {
	parts := []string{"archway new", opts.Name}
	if opts.Architecture != "" {
		parts = append(parts, "--arch", opts.Architecture)
	} else if opts.Template != "" && opts.Template != "api" {
		parts = append(parts, "--template", opts.Template)
	}
	if opts.Capabilities != "" {
		parts = append(parts, "--cap", opts.Capabilities)
	}
	if opts.ModulePath != "" {
		parts = append(parts, "--module", opts.ModulePath)
	}
	parts = append(parts, "--no-wizard")
	fmt.Printf("\nEquivalent command:\n  %s\n", strings.Join(parts, " "))
}

// runCompositionWizard runs the interactive wizard for architecture + capability selection.
func runCompositionWizard(opts *newCommandOptions) (map[string]interface{}, error) {
	// Step 1: Project basics.
	languages := provider.List()
	if opts.Language == "" {
		opts.Language = "go"
	}

	fields := []huh.Field{
		huh.NewInput().Title("Service name").Value(&opts.Name).Validate(func(value string) error {
			if strings.TrimSpace(value) == "" {
				return fmt.Errorf("service name is required")
			}
			return nil
		}),
		huh.NewInput().Title("Go module path").
			Description("e.g. github.com/org/my-service").
			Value(&opts.ModulePath),
		huh.NewInput().Title("Output directory").
			Placeholder(".").
			Value(&opts.OutputDir),
	}
	if len(languages) > 1 {
		langOpts := make([]huh.Option[string], 0, len(languages))
		for _, lang := range languages {
			langOpts = append(langOpts, huh.NewOption(lang, lang))
		}
		fields = append(fields, huh.NewSelect[string]().Title("Language").Value(&opts.Language).Options(langOpts...))
	}

	if err := huh.NewForm(huh.NewGroup(fields...)).Run(); err != nil {
		return nil, err
	}

	// Set default module path.
	if strings.TrimSpace(opts.ModulePath) == "" && opts.Name != "" {
		opts.ModulePath = fmt.Sprintf("github.com/example/%s", opts.Name)
	}

	providerImpl, err := provider.Get(opts.Language)
	if err != nil {
		return nil, err
	}
	tFS := providerImpl.GetTemplateFS()

	// Step 2: Architecture selection.
	architectures, err := listArchitectures(tFS)
	if err != nil {
		return nil, err
	}

	archOpts := make([]huh.Option[string], 0, len(architectures))
	for _, a := range architectures {
		archOpts = append(archOpts, huh.NewOption(a.label, a.name))
	}
	if opts.Architecture == "" && len(architectures) > 0 {
		opts.Architecture = architectures[0].name
	}

	if err := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("Architecture pattern").
			Options(archOpts...).
			Value(&opts.Architecture),
	)).Run(); err != nil {
		return nil, err
	}

	// Step 3: Capability selection (multi-select).
	capabilities, err := listCapabilities(tFS)
	if err != nil {
		return nil, err
	}

	capOpts := make([]huh.Option[string], 0, len(capabilities))
	for _, c := range capabilities {
		capOpts = append(capOpts, huh.NewOption(
			fmt.Sprintf("%-18s %s", c.name, c.description),
			c.name,
		))
	}

	var selectedCaps []string
	if err := huh.NewForm(huh.NewGroup(
		huh.NewMultiSelect[string]().
			Title("What does your service need? (space to toggle)").
			Options(capOpts...).
			Value(&selectedCaps),
	)).Run(); err != nil {
		return nil, err
	}

	// Step 4: Suggestions.
	suggestions := scaffold.ComputeSuggestions(selectedCaps)
	if len(suggestions) > 0 {
		suggOpts := make([]huh.Option[string], 0, len(suggestions))
		for _, s := range suggestions {
			suggOpts = append(suggOpts, huh.NewOption(
				fmt.Sprintf("%-18s %s", s.Capability, s.Reason),
				s.Capability,
			))
		}

		var acceptedSuggestions []string
		if err := huh.NewForm(huh.NewGroup(
			huh.NewMultiSelect[string]().
				Title("Based on your selections, you might also want:").
				Options(suggOpts...).
				Value(&acceptedSuggestions),
		)).Run(); err != nil {
			return nil, err
		}
		selectedCaps = append(selectedCaps, acceptedSuggestions...)
	}

	opts.Capabilities = strings.Join(selectedCaps, ",")

	// Step 5: Confirmation.
	fmt.Printf("\nReady to scaffold:\n")
	fmt.Printf("  Project:        %s\n", opts.Name)
	fmt.Printf("  Module:         %s\n", opts.ModulePath)
	fmt.Printf("  Architecture:   %s\n", opts.Architecture)
	if len(selectedCaps) > 0 {
		fmt.Printf("  Capabilities:   %s\n", strings.Join(selectedCaps, ", "))
	}
	fmt.Printf("  Output:         %s\n", filepath.Join(opts.OutputDir, opts.Name))

	var confirmed bool
	if err := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().Title("Proceed?").Value(&confirmed),
	)).Run(); err != nil {
		return nil, err
	}
	if !confirmed {
		return nil, fmt.Errorf("cancelled")
	}

	return map[string]interface{}{
		"ServiceName": opts.Name,
		"ModulePath":  opts.ModulePath,
	}, nil
}

type archEntry struct {
	name        string
	label       string
	description string
}

func listArchitectures(tFS fs.FS) ([]archEntry, error) {
	entries, err := fs.ReadDir(tFS, path.Join("templates", "architectures"))
	if err != nil {
		return nil, fmt.Errorf("list architectures: %w", err)
	}
	var archs []archEntry
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		data, err := fs.ReadFile(tFS, path.Join("templates", "architectures", entry.Name(), "manifest.yaml"))
		if err != nil {
			continue
		}
		m, err := scaffold.ParseManifest(data)
		if err != nil {
			continue
		}
		archs = append(archs, archEntry{
			name:        entry.Name(),
			label:       fmt.Sprintf("%s — %s", m.Name, m.Description),
			description: m.Description,
		})
	}
	return archs, nil
}

type capEntry struct {
	name        string
	description string
}

func listCapabilities(tFS fs.FS) ([]capEntry, error) {
	entries, err := fs.ReadDir(tFS, path.Join("templates", "capabilities"))
	if err != nil {
		return nil, fmt.Errorf("list capabilities: %w", err)
	}
	var caps []capEntry
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		data, err := fs.ReadFile(tFS, path.Join("templates", "capabilities", entry.Name(), "capability.yaml"))
		if err != nil {
			continue
		}
		cm, err := scaffold.ParseCapabilityManifest(data)
		if err != nil {
			continue
		}
		caps = append(caps, capEntry{
			name:        cm.Name,
			description: cm.Description,
		})
	}
	return caps, nil
}
