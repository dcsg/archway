package llm

import "context"

type LLMProvider interface {
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	Available() bool
}

type CompletionRequest struct {
	SystemPrompt string `json:"system_prompt"`
	UserPrompt   string `json:"user_prompt"`
	MaxTokens    int    `json:"max_tokens"`
}

type CompletionResponse struct {
	Content    string `json:"content"`
	TokensUsed int    `json:"tokens_used"`
}
