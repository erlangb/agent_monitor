package factory

import (
	"context"
	"erlangb/agentmonitor/internal/config"
	"fmt"

	openaimodel "github.com/cloudwego/eino-ext/components/model/openai"
	einomodel "github.com/cloudwego/eino/components/model"
)

const OPENAI = "openai"

// EinoChatModelFactory creates ToolCallingChatModel instances keyed by agent name.
// Config is read once at construction; individual model instances are created
// on demand via For.
type EinoChatModelFactory struct {
	cfg config.Config
}

// NewChatModelFactory returns a new EinoChatModelFactory backed by the given config.
func NewChatModelFactory(cfg config.Config) *EinoChatModelFactory {
	return &EinoChatModelFactory{cfg: cfg}
}

// CreateOpenAI creates an OpenAI ToolCallingChatModel with the given model ID, temperature, and max tokens.
// Pass maxTokens ≤ 0 to use the API default.
func (f *EinoChatModelFactory) CreateOpenAI(ctx context.Context, model string, temp float32, maxTokens int) (einomodel.ToolCallingChatModel, error) {
	modelCfg := &openaimodel.ChatModelConfig{
		APIKey:      f.cfg.Models[OPENAI].APIKey,
		Model:       model,
		Temperature: &temp,
	}

	if maxTokens > 0 {
		modelCfg.MaxCompletionTokens = &maxTokens
	}

	m, err := openaimodel.NewChatModel(ctx, modelCfg)

	if err != nil {
		return nil, fmt.Errorf("create chat model %q: %w", model, err)
	}
	return m, nil
}
