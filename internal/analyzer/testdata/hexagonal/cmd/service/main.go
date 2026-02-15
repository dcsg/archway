package main

import (
	"fmt"
	"log/slog"

	httpadapter "example.com/hexagonal/internal/adapter/http"
)

func main() {
	_ = httpadapter.Handler()
	slog.Info("started")
	fmt.Errorf("wrapped: %w", fmt.Errorf("boom"))
}
