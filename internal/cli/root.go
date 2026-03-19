package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	// Register language providers via init() so they're available via provider.Get().
	_ "github.com/dcsg/archway/providers/golang"
	_ "github.com/dcsg/archway/providers/typescript"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type globalOptions struct {
	NoColor bool
	Output  string
}

func Execute() error {
	return newRootCommand().Execute()
}

func newRootCommand() *cobra.Command {
	opts := &globalOptions{}

	cmd := &cobra.Command{
		Use:   "archway",
		Short: "Architecture-aware service composer and enforcer",
		Long: `Archway composes production-ready services from architecture patterns and capability modules.

Use archway new to scaffold projects, archway check to validate architecture rules, and archway analyze to understand existing codebases.`,
		SilenceUsage:  true,
		SilenceErrors: true,
	}

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
		newAddCommand(opts),
		newInitCommand(opts),
		newAnalyzeCommand(opts),
		newCheckCommand(opts),
		newDiffCommand(opts),
		newGuideCommand(opts),
		newDecideCommand(opts),
		newVersionCommand(),
		newSetupCommand(),
	)

	return cmd
}
