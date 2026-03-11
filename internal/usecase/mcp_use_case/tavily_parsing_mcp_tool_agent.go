package mcp_use_case

import (
	"context"
	"erlangb/agentmonitor/internal/agent/travel_assistant"
	"erlangb/agentmonitor/internal/factory"
	"erlangb/agentmonitor/internal/usecase"
)

type TavilyParsedUseCase struct {
	usecase.BaseEinoUseCase
	travelAssistant *travel_assistant.TravelAssistant
}

func NewTavilyParsedTravelAssistant(ctx context.Context, modelFactory *factory.EinoChatModelFactory, toolsFactory *factory.EinoToolsFactory, opts ...usecase.Option) (*TavilyParsedUseCase, error) {
	model, err := modelFactory.CreateOpenAI(ctx, "gpt-4.1", 0.1, 2048)
	if err != nil {
		return nil, err
	}

	tools, err := toolsFactory.CreateTavilyParsedEinoTools(ctx)
	if err != nil {
		return nil, err
	}

	travelAssistant, err := travel_assistant.NewTravelAssistant(ctx, model, tools)

	if err != nil {
		return nil, err
	}

	u := &TavilyParsedUseCase{travelAssistant: travelAssistant}
	for _, o := range opts {
		o.Apply(u)
	}
	return u, nil
}

func (u *TavilyParsedUseCase) Run(ctx context.Context, input string) (string, error) {
	ctx = u.PrepareRun(ctx, u.Name())
	return u.travelAssistant.Invoke(ctx, input)
}

func (u *TavilyParsedUseCase) Name() string { return "tavily_parsed" }

func (u *TavilyParsedUseCase) ExampleInput() string {
	return "Suggest 10 places to visit in Rome"
}

func (u *TavilyParsedUseCase) Description() string {
	return "Tavily Parsed — same travel agent but Tavily responses are pre-filtered to content+score before reaching the LLM" +
		"\nThis use case is useful for debugging and understanding the Tavily response wrapper in action."
}
