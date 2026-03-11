package mcp_use_case

import (
	"context"
	"erlangb/agentmonitor/internal/agent/travel_assistant"
	"erlangb/agentmonitor/internal/factory"
	"erlangb/agentmonitor/internal/usecase"
)

type TavilyFullInteracionUseCase struct {
	usecase.BaseEinoUseCase
	travelAssistant *travel_assistant.TravelAssistant
}

func NewTavilyFullInteractionTravelAssistant(ctx context.Context, modelFactory *factory.EinoChatModelFactory, toolsFactory *factory.EinoToolsFactory, opts ...usecase.Option) (*TavilyFullInteracionUseCase, error) {
	model, err := modelFactory.CreateOpenAI(ctx, "gpt-4.1", 0.1, 2048)
	if err != nil {
		return nil, err
	}

	tools, err := toolsFactory.CreateTavilyRawEinoTools(ctx)
	if err != nil {
		return nil, err
	}

	travelAssistant, err := travel_assistant.NewTravelAssistant(ctx, model, tools)

	if err != nil {
		return nil, err
	}

	u := &TavilyFullInteracionUseCase{travelAssistant: travelAssistant}

	for _, o := range opts {
		o.Apply(u)
	}

	return u, nil
}

func (u *TavilyFullInteracionUseCase) Run(ctx context.Context, input string) (string, error) {
	ctx = u.PrepareRun(ctx, u.Name())
	return u.travelAssistant.Invoke(ctx, input)
}

func (u *TavilyFullInteracionUseCase) Name() string { return "tavily_raw" }

func (u *TavilyFullInteracionUseCase) ExampleInput() string {
	return "Suggest 10 places to visit in Rome"
}

func (u *TavilyFullInteracionUseCase) Description() string {
	return "Tavily Raw — same travel agent but the full Tavily response (title, url, content, score, raw_content…) is passed directly to the LLM" +
		"\nThis use case is useful for debugging and understanding the raw Tavily response."
}
