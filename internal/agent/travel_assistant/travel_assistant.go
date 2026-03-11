package travel_assistant

import (
	"context"
	"erlangb/agentmonitor/internal/agent"
	"erlangb/agentmonitor/internal/agent/einoagent"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type TravelAssistant struct {
	agent adk.Agent
}

func NewTravelAssistant(ctx context.Context, chatModel model.ToolCallingChatModel, tools []tool.BaseTool) (*TravelAssistant, error) {
	travelAssistant, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "travelAssistant",
		Description: "find places to visit",
		Instruction: "return 10 places to visit based on user query. Use tavily to answer the question. Return STRICT JSON with {'address':'via..', 'city': 'rome'} ",
		Model:       chatModel,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: tools, //assign tools to agent
			},
		},
		OutputKey:        "places",
		MaxIterations:    2,
		ModelRetryConfig: &agent.DefaultRetry,
	})

	return &TravelAssistant{agent: travelAssistant}, err
}

func (u *TravelAssistant) Invoke(ctx context.Context, input string) (string, error) {
	return einoagent.RunAgentMessages(ctx, u.agent, []*schema.Message{schema.UserMessage(input)})
}
