package llm

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dcsg/archway/internal/config"
)

type DetectionInfo struct {
	Provider string
	Model    string
	BaseURL  string
}

func DetectProvider(cfg *config.AppConfig) (LLMProvider, error) {
	provider, _, err := DetectProviderWithInfo(cfg)
	return provider, err
}

func DetectProviderWithInfo(cfg *config.AppConfig) (LLMProvider, DetectionInfo, error) {
	if os.Getenv("ARCHWAY_SKIP_OLLAMA") != "1" && available("http://localhost:11434/api/tags") {
		model := "llama3.2"
		if cfg != nil && cfg.LLM.Model != "" {
			model = cfg.LLM.Model
		}
		return NewOpenAIProvider("ollama", "http://localhost:11434/v1", model), DetectionInfo{
			Provider: "ollama",
			Model:    model,
			BaseURL:  "http://localhost:11434/v1",
		}, nil
	}

	apiKey := strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	if apiKey == "" {
		apiKey = strings.TrimSpace(os.Getenv("ARCHWAY_LLM_API_KEY"))
	}
	if apiKey != "" {
		model := strings.TrimSpace(os.Getenv("ARCHWAY_LLM_MODEL"))
		if model == "" {
			model = "gpt-4o-mini"
		}
		baseURL := strings.TrimSpace(os.Getenv("ARCHWAY_LLM_BASE_URL"))
		return NewOpenAIProvider(apiKey, baseURL, model), DetectionInfo{Provider: "env", Model: model, BaseURL: baseURL}, nil
	}

	if cfg != nil && strings.TrimSpace(cfg.LLM.Provider) != "" {
		if strings.TrimSpace(cfg.LLM.APIKey) == "" && cfg.LLM.Provider != "ollama" {
			return nil, DetectionInfo{}, fmt.Errorf("llm provider %q configured without API key", cfg.LLM.Provider)
		}
		apiKey = cfg.LLM.APIKey
		if cfg.LLM.Provider == "ollama" && apiKey == "" {
			apiKey = "ollama"
		}
		return NewOpenAIProvider(apiKey, cfg.LLM.BaseURL, cfg.LLM.Model), DetectionInfo{
			Provider: cfg.LLM.Provider,
			Model:    cfg.LLM.Model,
			BaseURL:  cfg.LLM.BaseURL,
		}, nil
	}

	return NoopProvider{}, DetectionInfo{Provider: "none"}, nil
}

func available(url string) bool {
	client := &http.Client{Timeout: 800 * time.Millisecond}
	resp, err := client.Get(url) //nolint:noctx // short health check
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 300
}
