package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/dcsg/archway/internal/config"
	"github.com/dcsg/archway/internal/llm"
	"github.com/spf13/cobra"
)

type configureLLMOptions struct {
	Provider string
	Model    string
	APIKey   string
	BaseURL  string
	NoTest   bool
}

func newConfigureCommand(_ *globalOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure",
		Short: "Configure Archway settings",
		Long:  "Manage Archway configuration, including LLM provider setup.",
	}

	cmd.AddCommand(newConfigureLLMCommand())
	return cmd
}

func newConfigureLLMCommand() *cobra.Command {
	opts := &configureLLMOptions{}
	cmd := &cobra.Command{
		Use:   "llm",
		Short: "Configure LLM provider",
		Long:  "Configure OpenAI-compatible LLM provider settings for semantic analysis features.",
		Example: `  archway configure llm
  archway configure llm --provider ollama --model llama3.2
  archway configure llm --provider openai --api-key $OPENAI_API_KEY --model gpt-4o-mini`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return runConfigureLLM(opts)
		},
	}

	cmd.Flags().StringVar(&opts.Provider, "provider", "", "Provider: ollama|openai|groq|together|custom")
	cmd.Flags().StringVar(&opts.Model, "model", "", "Model name")
	cmd.Flags().StringVar(&opts.APIKey, "api-key", "", "API key (for cloud providers)")
	cmd.Flags().StringVar(&opts.BaseURL, "base-url", "", "OpenAI-compatible base URL")
	cmd.Flags().BoolVar(&opts.NoTest, "no-test", false, "Skip test completion call")

	return cmd
}

func runConfigureLLM(opts *configureLLMOptions) error {
	cfg, err := config.Load("")
	if err != nil {
		return err
	}

	if opts.Provider == "" {
		if err := runConfigureWizard(opts); err != nil {
			return err
		}
	}

	providerName := strings.ToLower(strings.TrimSpace(opts.Provider))
	if providerName == "" {
		return fmt.Errorf("provider is required")
	}

	if opts.Model == "" {
		switch providerName {
		case "ollama":
			opts.Model = "llama3.2"
		default:
			opts.Model = "gpt-4o-mini"
		}
	}

	if opts.BaseURL == "" {
		switch providerName {
		case "ollama":
			opts.BaseURL = "http://localhost:11434/v1"
		case "openai":
			opts.BaseURL = "https://api.openai.com/v1"
		case "groq":
			opts.BaseURL = "https://api.groq.com/openai/v1"
		case "together":
			opts.BaseURL = "https://api.together.xyz/v1"
		}
	}

	if providerName == "ollama" && opts.APIKey == "" {
		opts.APIKey = "ollama"
	}
	if providerName != "ollama" && strings.TrimSpace(opts.APIKey) == "" {
		return fmt.Errorf("--api-key is required for %s", providerName)
	}

	providerImpl := llm.NewOpenAIProvider(opts.APIKey, opts.BaseURL, opts.Model)
	if !opts.NoTest {
		if _, err := providerImpl.Complete(context.Background(), llm.CompletionRequest{
			SystemPrompt: "You are a connectivity check.",
			UserPrompt:   "Reply with: ok",
			MaxTokens:    10,
		}); err != nil {
			return fmt.Errorf("provider test failed: %w", err)
		}
	}

	cfg.LLM.Provider = providerName
	cfg.LLM.Model = opts.Model
	cfg.LLM.APIKey = opts.APIKey
	cfg.LLM.BaseURL = opts.BaseURL
	if err := config.Save("", cfg); err != nil {
		return err
	}
	fmt.Printf("LLM configured: provider=%s model=%s\n", providerName, opts.Model)
	return nil
}

func runConfigureWizard(opts *configureLLMOptions) error {
	providerName := "ollama"
	model := ""
	apiKey := ""
	baseURL := ""

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().Title("Provider").Value(&providerName).Options(
				huh.NewOption("ollama", "ollama"),
				huh.NewOption("openai", "openai"),
				huh.NewOption("groq", "groq"),
				huh.NewOption("together", "together"),
				huh.NewOption("custom", "custom"),
			),
			huh.NewInput().Title("Model").Value(&model),
			huh.NewInput().Title("API key (if required)").Value(&apiKey),
			huh.NewInput().Title("Base URL (optional)").Value(&baseURL),
		),
	)
	if err := form.Run(); err != nil {
		return err
	}

	opts.Provider = providerName
	opts.Model = strings.TrimSpace(model)
	opts.APIKey = strings.TrimSpace(apiKey)
	opts.BaseURL = strings.TrimSpace(baseURL)
	return nil
}
