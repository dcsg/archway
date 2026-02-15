package llm

import (
	"context"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIProvider struct {
	client *openai.Client
	model  string
}

func NewOpenAIProvider(apiKey, baseURL, model string) *OpenAIProvider {
	if strings.TrimSpace(apiKey) == "" {
		apiKey = ""
	}
	cfg := openai.DefaultConfig(apiKey)
	if strings.TrimSpace(baseURL) != "" {
		cfg.BaseURL = strings.TrimRight(baseURL, "/")
	}
	if strings.TrimSpace(model) == "" {
		model = openai.GPT4oMini
	}
	return &OpenAIProvider{
		client: openai.NewClientWithConfig(cfg),
		model:  model,
	}
}

func (p *OpenAIProvider) Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error) {
	if req.MaxTokens <= 0 {
		req.MaxTokens = 800
	}
	resp, err := p.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: p.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: req.SystemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: req.UserPrompt},
		},
		MaxTokens: req.MaxTokens,
	})
	if err != nil {
		return nil, err
	}
	content := ""
	if len(resp.Choices) > 0 {
		content = strings.TrimSpace(resp.Choices[0].Message.Content)
	}
	return &CompletionResponse{
		Content:    content,
		TokensUsed: resp.Usage.TotalTokens,
	}, nil
}

func (p *OpenAIProvider) Available() bool {
	return p != nil && p.client != nil
}
