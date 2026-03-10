package simple

import (
	"context"
	"erlangb/agentmonitor/internal/agent"
	"erlangb/agentmonitor/internal/factory"
	"erlangb/agentmonitor/internal/usecase"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

const ciphopherPrompt = `# ROLE: Just an hello world agent app`

// SimpleAgentLLMNoTools is a minimal use case that sends the user input straight to GPT with no tools.
type SimpleAgentLLMNoTools struct {
	usecase.BaseEinoUseCase
	modelFactory *factory.EinoChatModelFactory
}

// NewSimpleAgentLLMNoTools returns a SimpleAgentLLMNoTools with the given model factory and options applied.
func NewSimpleAgentLLMNoTools(modelFactory *factory.EinoChatModelFactory, opts ...usecase.Option) *SimpleAgentLLMNoTools {
	u := &SimpleAgentLLMNoTools{modelFactory: modelFactory}
	for _, o := range opts {
		o.Apply(u)
	}
	return u
}

func (u *SimpleAgentLLMNoTools) Run(ctx context.Context, input string) (string, error) {
	chatModel, err := u.modelFactory.CreateOpenAI(ctx, "gpt-4.1", 0.7, 2048)
	if err != nil {
		return "", err
	}

	ag, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          "simple_agent",
		Description:   "simple agent",
		Instruction:   ciphopherPrompt,
		Model:         chatModel,
		MaxIterations: 1,
	})
	if err != nil {
		return "agent error", err
	}

	ctx = u.PrepareRun(ctx, u.Name())

	answer, err := agent.RunAgentMessages(ctx, ag, []*schema.Message{schema.UserMessage(input)})
	if err != nil {
		return "", err
	}

	if answer == "" {
		return "", fmt.Errorf("%s: model returned empty answer", u.Name())
	}
	return answer, nil
}

func (u *SimpleAgentLLMNoTools) Name() string { return "simple agent" }

func (u *SimpleAgentLLMNoTools) ExampleInput() string {
	return "Hello"
}

func (u *SimpleAgentLLMNoTools) Description() string {
	return "Simple LLM/agent chat model agent"
}
