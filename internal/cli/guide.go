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
	var target string

	cmd := &cobra.Command{
		Use:   "guide",
		Short: "Generate AI agent architecture instructions",
		Long: `Generate architecture guidance files for AI coding agents.

Reads archway.yaml from the current directory and generates instruction files
for Claude Code, Cursor, GitHub Copilot, and Windsurf.`,
		Example: `  archway guide
  archway guide --target claude
  archway guide --target cursor`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runGuide(target)
		},
	}

	cmd.Flags().StringVar(&target, "target", "all", "Output target: all, claude, cursor, copilot, windsurf")
	_ = cmd.RegisterFlagCompletionFunc("target", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"all", "claude", "cursor", "copilot", "windsurf"}, cobra.ShellCompDirectiveNoFileComp
	})

	return cmd
}

func runGuide(target string) error {
	cfgPath, err := config.FindArchwayYAML(".")
	if err != nil {
		return fmt.Errorf("no archway.yaml found in current directory or parents: %w", err)
	}

	cfg, err := config.LoadArchwayYAML(cfgPath)
	if err != nil {
		return fmt.Errorf("load archway.yaml: %w", err)
	}

	projectDir := "."

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
