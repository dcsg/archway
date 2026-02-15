package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/dcsg/archway/internal/analyzer/detector"
	"github.com/dcsg/archway/internal/config"
	"github.com/spf13/cobra"
)

type initCommandOptions struct {
	Preset       string
	Path         string
	Architecture string
	Language     string
	Force        bool
	NoWizard     bool
}

func newInitCommand(_ *globalOptions) *cobra.Command {
	opts := &initCommandOptions{}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize archway.yaml for an existing project",
		Long:  "Generate an archway.yaml file declaring desired architecture and rules.",
		Example: `  archway init
  archway init --preset archway/go-hexagonal-strict
  archway init --language go --architecture hexagonal --no-wizard`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return runInit(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Preset, "preset", "", "Preset to extend, e.g. archway/go-hexagonal-strict")
	cmd.Flags().StringVar(&opts.Path, "path", ".", "Project path")
	cmd.Flags().StringVar(&opts.Architecture, "architecture", "", "Architecture pattern")
	cmd.Flags().StringVar(&opts.Language, "language", "", "Language")
	cmd.Flags().BoolVar(&opts.Force, "force", false, "Overwrite existing archway.yaml")
	cmd.Flags().BoolVar(&opts.NoWizard, "no-wizard", false, "Disable interactive wizard")

	return cmd
}

func runInit(opts *initCommandOptions) error {
	if opts.Path == "" {
		opts.Path = "."
	}
	archwayPath := filepath.Join(opts.Path, "archway.yaml")
	if _, err := os.Stat(archwayPath); err == nil && !opts.Force {
		return fmt.Errorf("%s already exists (use --force to overwrite)", archwayPath)
	}

	if strings.TrimSpace(opts.Language) == "" {
		if detected, _, err := detector.DetectLanguage(opts.Path); err == nil && detected != "unknown" {
			opts.Language = detected
		} else {
			opts.Language = "go"
		}
	}
	if strings.TrimSpace(opts.Architecture) == "" {
		opts.Architecture = "hexagonal"
	}

	if !opts.NoWizard && strings.TrimSpace(opts.Preset) == "" {
		if err := runInitWizard(opts); err != nil {
			return err
		}
	}

	cfg := config.DefaultArchwayConfig(opts.Language, opts.Architecture)
	if strings.TrimSpace(opts.Preset) != "" {
		cfg.Extends = []string{opts.Preset}
	}
	if err := config.SaveArchwayYAML(archwayPath, cfg); err != nil {
		return err
	}
	fmt.Printf("Generated %s\n", archwayPath)
	return nil
}

func runInitWizard(opts *initCommandOptions) error {
	architecture := opts.Architecture
	language := opts.Language
	extends := opts.Preset
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().Title("Language").Value(&language).Options(
				huh.NewOption("go", "go"),
			),
			huh.NewSelect[string]().Title("Architecture").Value(&architecture).Options(
				huh.NewOption("hexagonal", "hexagonal"),
				huh.NewOption("clean", "clean"),
				huh.NewOption("ddd", "ddd"),
				huh.NewOption("layered", "layered"),
				huh.NewOption("flat", "flat"),
			),
			huh.NewInput().Title("Preset extends (optional)").Value(&extends),
		),
	)
	if err := form.Run(); err != nil {
		return err
	}
	opts.Language = language
	opts.Architecture = architecture
	opts.Preset = strings.TrimSpace(extends)
	return nil
}
