package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/dcsg/archway/internal/provider"
)

func GenerateADRs(ctx context.Context, analysis *provider.AnalyzeResponse, llmProvider LLMProvider) ([]provider.ADR, int, error) {
	if llmProvider == nil || !llmProvider.Available() {
		return nil, 0, ErrNoLLM
	}
	prompt := fmt.Sprintf("Architecture=%s, Framework=%s, Conventions=%s/%s/%s. Return 2 ADR bullets in format: Title|Context|Decision|Consequences.",
		analysis.Architecture.Pattern,
		analysis.Framework.Name,
		analysis.Conventions.ErrorHandling.Pattern,
		analysis.Conventions.Logging.Pattern,
		analysis.Conventions.Config.Pattern,
	)
	resp, err := llmProvider.Complete(ctx, CompletionRequest{
		SystemPrompt: "You generate concise architecture decision records.",
		UserPrompt:   prompt,
		MaxTokens:    500,
	})
	if err != nil {
		return nil, 0, err
	}
	adrs := []provider.ADR{}
	for _, line := range strings.Split(resp.Content, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "-"))
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}
		adrs = append(adrs, provider.ADR{
			Title:        strings.TrimSpace(parts[0]),
			Context:      strings.TrimSpace(parts[1]),
			Decision:     strings.TrimSpace(parts[2]),
			Consequences: strings.TrimSpace(parts[3]),
		})
	}
	return adrs, resp.TokensUsed, nil
}

func ExtractInvariants(ctx context.Context, analysis *provider.AnalyzeResponse, llmProvider LLMProvider) ([]string, int, error) {
	if llmProvider == nil || !llmProvider.Available() {
		return nil, 0, ErrNoLLM
	}
	prompt := fmt.Sprintf("Given testing pattern %s and architecture %s, list up to 5 invariants as bullet points.",
		analysis.Conventions.Testing.Pattern, analysis.Architecture.Pattern)
	resp, err := llmProvider.Complete(ctx, CompletionRequest{
		SystemPrompt: "Return concise invariant bullets.",
		UserPrompt:   prompt,
		MaxTokens:    400,
	})
	if err != nil {
		return nil, 0, err
	}
	invariants := []string{}
	for _, line := range strings.Split(resp.Content, "\n") {
		line = strings.TrimSpace(strings.TrimPrefix(line, "-"))
		if line != "" {
			invariants = append(invariants, line)
		}
	}
	return invariants, resp.TokensUsed, nil
}

func SemanticAssessment(ctx context.Context, analysis *provider.AnalyzeResponse, llmProvider LLMProvider) (string, int, error) {
	if llmProvider == nil || !llmProvider.Available() {
		return "", 0, ErrNoLLM
	}
	prompt := fmt.Sprintf("Assess architecture quality for a %s codebase using %s. Keep under 120 words.",
		analysis.Architecture.Pattern,
		analysis.Framework.Name,
	)
	resp, err := llmProvider.Complete(ctx, CompletionRequest{
		SystemPrompt: "Provide actionable architecture assessment.",
		UserPrompt:   prompt,
		MaxTokens:    300,
	})
	if err != nil {
		return "", 0, err
	}
	return strings.TrimSpace(resp.Content), resp.TokensUsed, nil
}
