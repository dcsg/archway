package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type globalOptions struct {
	ConfigPath string
	Verbose    bool
	NoColor    bool
	Output     string
}

func Execute() error {
	return newRootCommand().Execute()
}

func newRootCommand() *cobra.Command {
	opts := &globalOptions{}

	cmd := &cobra.Command{
		Use:   "archway",
		Short: "Terraform for Code Architecture",
		Long: `Archway scaffolds and analyzes code architecture with a Terraform-style workflow.

Use archway new to scaffold projects and archway analyze to understand existing codebases.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVar(&opts.ConfigPath, "config", "", "Path to config file")
	cmd.PersistentFlags().BoolVar(&opts.Verbose, "verbose", false, "Enable verbose logging")
	cmd.PersistentFlags().BoolVar(&opts.NoColor, "no-color", false, "Disable colored output")
	cmd.PersistentFlags().StringVarP(&opts.Output, "output", "o", "terminal", "Output format: terminal|json|markdown")

	_ = cmd.RegisterFlagCompletionFunc("output", func(_ *cobra.Command, _ []string, _ string) ([]string, cobra.ShellCompDirective) {
		return []string{"terminal", "json", "markdown"}, cobra.ShellCompDirectiveNoFileComp
	})

	cmd.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		switch opts.Output {
		case "terminal", "json", "markdown":
			return nil
		default:
			return fmt.Errorf("invalid --output value %q (expected terminal|json|markdown)", opts.Output)
		}
	}

	cmd.AddCommand(
		newNewCommand(opts),
		newInitCommand(opts),
		newAnalyzeCommand(opts),
		newConfigureCommand(opts),
		newVersionCommand(),
		newMCPCommand(opts),
	)

	return cmd
}

func setVersionInfo(v, c, d string) {
	if v != "" {
		version = v
	}
	if c != "" {
		commit = c
	}
	if d != "" {
		date = d
	}
}
