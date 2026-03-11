package cli

import (
	"fmt"

	"github.com/dcsg/archway/internal/config"
	"github.com/dcsg/archway/internal/guide"
	"github.com/dcsg/archway/internal/provider"
	"github.com/spf13/cobra"

	// Ensure Go provider is registered.
	_ "github.com/dcsg/archway/providers/golang"
)

func newGuideCommand(_ *globalOptions) *cobra.Command {
	var (
		target      string
		catalogOnly bool
	)

	cmd := &cobra.Command{
		Use:   "guide",
		Short: "Generate AI agent architecture instructions",
		Long: `Generate architecture guidance files for AI coding agents.

Reads archway.yaml from the current directory and generates instruction files
for Claude Code, Cursor, GitHub Copilot, and Windsurf.

Use --catalog-only to generate only the capability catalog without requiring
an archway.yaml configuration file.`,
		Example: `  archway guide
  archway guide --target claude
  archway guide --target cursor
  archway guide --catalog-only`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runGuide(target, catalogOnly)
		},
	}

	cmd.Flags().StringVar(&target, "target", "all", "Output target: all, claude, cursor, copilot, windsurf")
	cmd.Flags().BoolVar(&catalogOnly, "catalog-only", false, "Generate only the capability catalog (no archway.yaml required)")
	_ = cmd.RegisterFlagCompletionFunc("target", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"all", "claude", "cursor", "copilot", "windsurf"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func runGuide(target string, catalogOnly bool) error {
	projectDir := "."

	if catalogOnly {
		// Catalog-only mode: no archway.yaml required.
		p, provErr := provider.Get("go")
		opts := guide.GenerateOptions{
			ProjectDir:  projectDir,
			Target:      target,
			CatalogOnly: true,
		}
		if provErr == nil {
			opts.TemplateFS = p.GetTemplateFS()
		}
		if err := guide.Generate(opts); err != nil {
			return err
		}
		fmt.Printf("Guide generated (catalog-only) for target: %s\n", target)
		return nil
	}

	cfgPath, err := config.FindArchwayYAML(".")
	if err != nil {
		return fmt.Errorf("no archway.yaml found in current directory or parents: %w", err)
	}

	cfg, err := config.LoadArchwayYAML(cfgPath)
	if err != nil {
		return fmt.Errorf("load archway.yaml: %w", err)
	}

	// Look up the language provider to get the template FS for pattern extraction.
	p, provErr := provider.Get(cfg.Language)
	if provErr == nil {
		if err := guide.GenerateFromConfig(projectDir, cfg, target, p.GetTemplateFS()); err != nil {
			return err
		}
	} else {
		if err := guide.GenerateFromConfig(projectDir, cfg, target); err != nil {
			return err
		}
	}

	fmt.Printf("Guide generated for target: %s\n", target)
	return nil
}
