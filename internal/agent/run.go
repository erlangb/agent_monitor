package agent

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// RunAgentMessages runs an adk.Agent with pre-formatted messages.
func RunAgentMessages(ctx context.Context, a adk.Agent, messages []*schema.Message) (string, error) {
	runner := adk.NewRunner(ctx, adk.RunnerConfig{Agent: a})

	iter := runner.Run(ctx, messages)

	var answer string
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			return "", event.Err
		}
		if event.Output == nil || event.Output.MessageOutput == nil {
			continue
		}
		mv := event.Output.MessageOutput
		if string(mv.Role) == "assistant" && len(mv.Message.ToolCalls) == 0 {
			answer = mv.Message.Content
		}
	}
	return answer, nil
}
