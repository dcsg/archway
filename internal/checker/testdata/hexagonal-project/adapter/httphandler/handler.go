package httphandler

import "context"

type Handler struct{}

// HandleRequest uses context.Background() in a handler — should trigger context_background_in_handler.
func (h *Handler) HandleRequest() {
	ctx := context.Background()
	_ = ctx

	// Naked goroutine — should trigger naked_goroutine.
	go func() {
		println("background work")
	}()
}
