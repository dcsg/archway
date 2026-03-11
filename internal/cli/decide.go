package cli

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dcsg/archway/internal/config"
	"github.com/dcsg/archway/internal/guide"
	"github.com/spf13/cobra"
)

func newDecideCommand(_ *globalOptions) *cobra.Command {
	var list bool

	cmd := &cobra.Command{
		Use:   "decide [topic]",
		Short: "Resolve architecture decisions",
		Long:  `Interactive CLI for resolving architecture decision gates.`,
		Example: `  archway decide                          # Interactive
  archway decide authentication-strategy  # Specific topic
  archway decide --list                   # Show all decisions`,
		RunE: func(_ *cobra.Command, args []string) error {
			if list {
				return runDecideList()
			}
			topic := ""
			if len(args) > 0 {
				topic = args[0]
			}
			return runDecide(topic)
		},
	}

	cmd.Flags().BoolVar(&list, "list", false, "List all decisions with status")
	return cmd
}

func runDecideList() error {
	cfgPath, cfg, err := loadDecideConfig()
	if err != nil {
		return err
	}

	decisions := cfg.Decisions
	if len(decisions) == 0 {
		decisions = guide.AutoPopulateDecisions(cfg.Architecture, cfg.Capabilities)
		cfg.Decisions = decisions
		if saveErr := config.SaveArchwayYAML(cfgPath, cfg); saveErr != nil {
			return fmt.Errorf("save archway.yaml: %w", saveErr)
		}
		fmt.Println("Auto-populated decisions from project config.")
	}

	byTier := guide.DecisionsByTier(decisions)
	for tier := 1; tier <= 2; tier++ {
		tierDecisions, ok := byTier[tier]
		if !ok {
			continue
		}
		fmt.Printf("\nTier %d:\n", tier)
		for _, d := range tierDecisions {
			if d.Status == "decided" {
				fmt.Printf("  [v] %s: %s\n", d.Topic, d.Choice)
			} else {
				fmt.Printf("  [x] %s: UNDECIDED\n", d.Topic)
			}
		}
	}

	undecided := guide.UndecidedDecisions(decisions)
	if len(undecided) > 0 {
		fmt.Printf("\n%d undecided. Run `archway decide` to resolve.\n", len(undecided))
	} else {
		fmt.Println("\nAll decisions resolved.")
	}
	return nil
}

func runDecide(topic string) error {
	cfgPath, cfg, err := loadDecideConfig()
	if err != nil {
		return err
	}

	if len(cfg.Decisions) == 0 {
		cfg.Decisions = guide.AutoPopulateDecisions(cfg.Architecture, cfg.Capabilities)
		if saveErr := config.SaveArchwayYAML(cfgPath, cfg); saveErr != nil {
			return fmt.Errorf("save archway.yaml: %w", saveErr)
		}
		fmt.Println("Auto-populated decisions from project config.")
	}

	scanner := bufio.NewScanner(os.Stdin)

	if topic == "" {
		undecided := guide.UndecidedDecisions(cfg.Decisions)
		if len(undecided) == 0 {
			fmt.Println("All decisions are resolved.")
			return nil
		}

		fmt.Println("Undecided topics:")
		for i, d := range undecided {
			tmpl, _ := guide.FindDecisionTemplate(d.Topic)
			fmt.Printf("  %d. %s — %s\n", i+1, d.Topic, tmpl.Question)
		}

		fmt.Print("\nSelect topic number: ")
		if !scanner.Scan() {
			return fmt.Errorf("no input")
		}
		idx, parseErr := strconv.Atoi(strings.TrimSpace(scanner.Text()))
		if parseErr != nil || idx < 1 || idx > len(undecided) {
			return fmt.Errorf("invalid selection")
		}
		topic = undecided[idx-1].Topic
	}

	tmpl, found := guide.FindDecisionTemplate(topic)
	if !found {
		return fmt.Errorf("unknown decision topic: %q", topic)
	}

	fmt.Printf("\n%s\n", tmpl.Question)
	fmt.Println("Options:")
	for i, opt := range tmpl.Options {
		fmt.Printf("  %d. %s\n", i+1, opt)
	}

	fmt.Print("\nSelect option number: ")
	if !scanner.Scan() {
		return fmt.Errorf("no input")
	}
	idx, parseErr := strconv.Atoi(strings.TrimSpace(scanner.Text()))
	if parseErr != nil || idx < 1 || idx > len(tmpl.Options) {
		return fmt.Errorf("invalid selection")
	}
	choice := tmpl.Options[idx-1]

	fmt.Print("Rationale (optional, press Enter to skip): ")
	rationale := ""
	if scanner.Scan() {
		rationale = strings.TrimSpace(scanner.Text())
	}

	updated, resolveErr := guide.ResolveDecision(cfg.Decisions, topic, choice, rationale, "cli")
	if resolveErr != nil {
		return resolveErr
	}
	cfg.Decisions = updated

	if saveErr := config.SaveArchwayYAML(cfgPath, cfg); saveErr != nil {
		return fmt.Errorf("save archway.yaml: %w", saveErr)
	}

	fmt.Printf("Decided: %s = %s\n", topic, choice)
	return nil
}

func loadDecideConfig() (string, *config.ArchwayConfig, error) {
	cfgPath, err := config.FindArchwayYAML(".")
	if err != nil {
		return "", nil, fmt.Errorf("no archway.yaml found: %w", err)
	}

	cfg, err := config.LoadArchwayYAML(cfgPath)
	if err != nil {
		return "", nil, fmt.Errorf("load archway.yaml: %w", err)
	}

	return cfgPath, cfg, nil
}
