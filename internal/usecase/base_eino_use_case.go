package usecase

import (
	"context"

	"github.com/cloudwego/eino/callbacks"
)

// BaseEinoUseCase provides Eino callback wiring. Embed by value in concrete use case structs.
type BaseEinoUseCase struct {
	callbackHandlers []callbacks.Handler
}

func (b *BaseEinoUseCase) enableCallbacks(handlers ...callbacks.Handler) {
	b.callbackHandlers = handlers
}

// PrepareRun initialises the Eino callback context for the current run.
func (b *BaseEinoUseCase) PrepareRun(ctx context.Context, name string) context.Context {
	return callbacks.InitCallbacks(ctx, &callbacks.RunInfo{Name: name}, b.callbackHandlers...)
}

// Option mutates a UseCase before Run begins.
type Option interface {
	Apply(uc UseCase)
}

// WithCallBackHandlers passes Eino callback handlers into the use case at construction time.
type WithCallBackHandlers struct {
	Handlers []callbacks.Handler
}

func (o WithCallBackHandlers) Apply(uc UseCase) {
	if e, ok := uc.(interface{ enableCallbacks(...callbacks.Handler) }); ok {
		e.enableCallbacks(o.Handlers...)
	}
}
