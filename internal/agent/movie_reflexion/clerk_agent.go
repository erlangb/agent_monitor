package movie_reflexion

import (
	"context"
	"erlangb/agentmonitor/internal/agent"
	"erlangb/agentmonitor/internal/agent/einoagent"
	"fmt"

	appmodel "erlangb/agentmonitor/internal/model"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/adk"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	etool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const ClerkSystemPrompt = `# ROLE: The Clerk (Movie Fact-Checker)
You are a pedantic, data-driven movie database clerk. You have no taste — only facts.

# RESPONSIBILITIES
- Verify that each suggested movie actually exists using your search tool
- For each movie: call tavily_search with max_results=3
- Check release years, directors, and basic facts to match the user's request
- Determine if the movies match the user's request
- Be literal and objective: if a movie doesn't exist, say so

# DECISION
After verifying all movies:
- If ALL movies exist AND match the request well: set isSatisfied to true
- If ANY movie is wrong or doesn't fit: set isSatisfied to false and explain what to fix


# SUMMARY
Make sure the summary IN OUTPUT mentions all movies titles so the user can understand from it what to keep and what needs to be removed.

# SUMMARY
Make sure the release years requested by user are satisfied precisely.

# OUTPUT FORMAT (strict JSON)
{
  "critiques": ["string — specific issue found, or empty if all good"],
  "isSatisfied": boolean,
  "summary": "string — one sentence verdict"
}`

// ClerkAgent is a fact-checking node.
// It uses an adk.Agent internally to support the tool loop (tavily search).
// Construct with NewClerkAgent, call Invoke directly.
type ClerkAgent struct {
	agent adk.Agent
	tmpl  prompt.ChatTemplate
}

// NewClerkAgent builds an adk.Agent with the Tavily search tool wired in and returns a ClerkAgent.
func NewClerkAgent(ctx context.Context, m einomodel.ToolCallingChatModel, tools []etool.BaseTool) (*ClerkAgent, error) {
	a, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "clerk_structured",
		Description: "fact-checks movie suggestions against structured query constraints",
		Instruction: ClerkSystemPrompt,
		Model:       m,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{Tools: tools},
		},
		MaxIterations: 4,

		ModelRetryConfig: &agent.DefaultRetry,
	})
	if err != nil {
		return nil, err
	}

	return &ClerkAgent{agent: a, tmpl: clerkTemplate()}, nil
}

func clerkTemplate() *prompt.DefaultChatTemplate {
	return prompt.FromMessages(schema.GoTemplate,
		schema.UserMessage(`User wants: "{{.query.OriginalText}}"
{{if .query.PrimaryGenre}}Primary genre: {{.query.PrimaryGenre}}
{{end -}}{{if or .query.StartYear .query.EndYear}}Year range: {{.query.StartYear}}–{{.query.EndYear}}
{{end -}}{{if .query.IsClassic}}Style: classic films only
{{end -}}{{if .query.QueryInfo}}Search context: {{.query.QueryInfo}}
{{end}}Movies to verify:
{{range .movies}}- "{{.Title}}" ({{.Year}}) — {{.Reason}}
{{end}}`),
	)
}

func (s *ClerkAgent) Invoke(ctx context.Context, state *appmodel.FindMoviesState) (*appmodel.FindMoviesState, error) {
	msgs, err := s.tmpl.Format(ctx, map[string]any{
		"query":  state.UserQuery,
		"movies": state.CurrentDraft,
	})
	if err != nil {
		return nil, fmt.Errorf("clerk template: %w", err)
	}

	answer, err := einoagent.RunAgentMessages(ctx, s.agent, msgs)
	if err != nil {
		return nil, fmt.Errorf("clerk run: %w", err)
	}

	critique, err := ParseClerkResponse(answer)
	if err != nil {
		state.CritiqueHistory = append(state.CritiqueHistory, answer)
		return state, nil
	}

	state.CritiqueHistory = append(state.CritiqueHistory, critique.Summary)
	state.IsSatisfied = critique.IsSatisfied
	state.LastSummary = critique.Summary
	return state, nil
}

// ParseClerkResponse unmarshals the LLM output into a ClerkResponse.
func ParseClerkResponse(raw string) (appmodel.ClerkResponse, error) {
	var resp appmodel.ClerkResponse
	if err := sonic.UnmarshalString(raw, &resp); err != nil {
		return appmodel.ClerkResponse{}, fmt.Errorf("parse clerk: %w", err)
	}
	return resp, nil
}
