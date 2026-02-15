package llm

import (
	"context"
	"errors"
)

var ErrNoLLM = errors.New("no llm provider configured")

type NoopProvider struct{}

func (NoopProvider) Complete(_ context.Context, _ CompletionRequest) (*CompletionResponse, error) {
	return nil, ErrNoLLM
}

func (NoopProvider) Available() bool {
	return false
}
