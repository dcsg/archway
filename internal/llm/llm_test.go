package llm

import (
	"context"
	"os"
	"testing"

	"github.com/dcsg/archway/internal/config"
	"github.com/dcsg/archway/internal/provider"
)

type mockProvider struct {
	response string
}

func (m mockProvider) Complete(_ context.Context, _ CompletionRequest) (*CompletionResponse, error) {
	return &CompletionResponse{Content: m.response, TokensUsed: 42}, nil
}

func (m mockProvider) Available() bool { return true }

func TestNoopProvider(t *testing.T) {
	_, err := NoopProvider{}.Complete(context.Background(), CompletionRequest{})
	if err != ErrNoLLM {
		t.Fatalf("expected ErrNoLLM, got %v", err)
	}
}

func TestDetectProviderFromConfig(t *testing.T) {
	t.Setenv("ARCHWAY_SKIP_OLLAMA", "1")
	cfg := &config.AppConfig{LLM: config.LLMConfig{Provider: "openai", APIKey: "x", Model: "gpt-4o-mini", BaseURL: "https://example.com/v1"}}
	p, info, err := DetectProviderWithInfo(cfg)
	if err != nil {
		t.Fatalf("DetectProviderWithInfo() error = %v", err)
	}
	if !p.Available() {
		t.Fatal("provider should be available")
	}
	if info.Provider != "openai" {
		t.Fatalf("provider = %q, want openai", info.Provider)
	}
}

func TestDetectProviderFromEnv(t *testing.T) {
	t.Setenv("ARCHWAY_SKIP_OLLAMA", "1")
	t.Setenv("OPENAI_API_KEY", "secret")
	t.Setenv("ARCHWAY_LLM_MODEL", "gpt-test")
	p, info, err := DetectProviderWithInfo(nil)
	if err != nil {
		t.Fatalf("DetectProviderWithInfo() error = %v", err)
	}
	if !p.Available() {
		t.Fatal("provider should be available")
	}
	if info.Provider != "env" {
		t.Fatalf("provider = %q, want env", info.Provider)
	}
}

func TestOpenAIProviderCompleteErrorPath(t *testing.T) {
	p := NewOpenAIProvider("key", "http://127.0.0.1:1/v1", "test")
	if _, err := p.Complete(context.Background(), CompletionRequest{SystemPrompt: "s", UserPrompt: "u"}); err == nil {
		t.Fatal("expected network error for unreachable endpoint")
	}
}

func TestLLMEnhancementFunctions(t *testing.T) {
	analysis := &provider.AnalyzeResponse{
		Architecture: provider.ArchitectureResult{Pattern: "hexagonal"},
		Framework:    provider.FrameworkResult{Name: "chi"},
		Conventions: provider.ConventionResults{
			ErrorHandling: provider.ConventionFinding{Pattern: "wrapped"},
			Logging:       provider.ConventionFinding{Pattern: "slog"},
			Config:        provider.ConventionFinding{Pattern: "koanf"},
			Testing:       provider.TestingFinding{Pattern: "table-driven"},
		},
	}
	mock := mockProvider{response: "Decision title|ctx|decision|consequence\n- invariant one\n- invariant two"}
	adrs, _, err := GenerateADRs(context.Background(), analysis, mock)
	if err != nil {
		t.Fatalf("GenerateADRs error: %v", err)
	}
	if len(adrs) == 0 {
		t.Fatal("expected ADRs")
	}
	invariants, _, err := ExtractInvariants(context.Background(), analysis, mock)
	if err != nil {
		t.Fatalf("ExtractInvariants error: %v", err)
	}
	if len(invariants) == 0 {
		t.Fatal("expected invariants")
	}
	assessment, _, err := SemanticAssessment(context.Background(), analysis, mock)
	if err != nil {
		t.Fatalf("SemanticAssessment error: %v", err)
	}
	if assessment == "" {
		t.Fatal("expected semantic assessment")
	}
}

func TestDetectProviderNoConfig(t *testing.T) {
	t.Setenv("ARCHWAY_SKIP_OLLAMA", "1")
	_ = os.Unsetenv("OPENAI_API_KEY")
	_ = os.Unsetenv("ARCHWAY_LLM_API_KEY")
	p, info, err := DetectProviderWithInfo(nil)
	if err != nil {
		t.Fatalf("DetectProviderWithInfo() error = %v", err)
	}
	if info.Provider != "none" {
		t.Fatalf("provider = %q, want none", info.Provider)
	}
	if p.Available() {
		t.Fatal("noop provider should not be available")
	}
}
