package movie_reflexion

import (
	"context"
	"encoding/json"
	"erlangb/agentmonitor/internal/agent/movie_reflexion"
	"erlangb/agentmonitor/internal/factory"
	"erlangb/agentmonitor/internal/usecase"
	"fmt"
)

// RefinerQueryUseCase runs only the refiner chain and returns the structured query parameters as JSON.
type RefinerQueryUseCase struct {
	usecase.BaseEinoUseCase
	modelFactory *factory.EinoChatModelFactory
}

// NewRefinerQueryUseCase returns a RefinerQueryUseCase with the given model factory and options applied.
func NewRefinerQueryUseCase(modelFactory *factory.EinoChatModelFactory, opts ...usecase.Option) *RefinerQueryUseCase {
	u := &RefinerQueryUseCase{modelFactory: modelFactory}
	for _, o := range opts {
		o.Apply(u)
	}
	return u
}

func (u *RefinerQueryUseCase) Run(ctx context.Context, input string) (string, error) {
	chatModel, err := u.modelFactory.CreateOpenAI(ctx, "gpt-4.1", 0.2, 1024)
	if err != nil {
		return "", err
	}

	ctx = u.PrepareRun(ctx, u.Name())

	refinerNode, err := movie_reflexion.NewRefinerChain(ctx, chatModel)
	if err != nil {
		return "", fmt.Errorf("%s: build agent: %w", u.Name(), err)
	}

	result, err := refinerNode.Invoke(ctx, input)
	if err != nil {
		return "", fmt.Errorf("%s: invoke: %w", u.Name(), err)
	}

	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", fmt.Errorf("%s: marshal result: %w", u.Name(), err)
	}
	return string(out), nil
}

func (u *RefinerQueryUseCase) Name() string { return "refiner-query" }

func (u *RefinerQueryUseCase) ExampleInput() string {
	return "I want obscure horror films from the 90s with a surreal vibe"
}

func (u *RefinerQueryUseCase) Description() string {
	return "Refiner — extracts structured search parameters (genre, years, mood) from a free-text movie query."
}
