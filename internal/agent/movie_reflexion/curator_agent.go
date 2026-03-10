package movie_reflexion

import (
	"context"
	"erlangb/agentmonitor/internal/agent"
	"fmt"

	appmodel "erlangb/agentmonitor/internal/model"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const FinalSystemPrompt = `# ROLE: The Curator (Final Arbitrator)
You are a precise, no-nonsense film curator. You have received a list of movie suggestions and feedback from a fact-checker.
Your job is to produce the definitive, best-possible movie list given what you know.

# RESPONSIBILITIES
- Review the current draft and final verdict from fact-checker
- Keep movies that are factually correct and match the user's request
- Remove movies that were flagged as wrong or irrelevant
- Prioritise quality and accuracy over novelty
- If no movies match the user's request, answer with an empty list
- Never add a new movie; don't use your internal knowledge

# OUTPUT FORMAT (strict JSON)
{
  "movies": [
    {
      "title": "string",
      "year": number,
      "reason": "string — why this matches the vibe"
    }
  ]
}`

// CuratorChain is the terminal node of the find-movies graph.
// It receives the final draft and the fact-checker's verdict, then produces
// the curated movie list. Construct with NewCuratorChain.
type CuratorChain struct {
	runnable compose.Runnable[*appmodel.FindMoviesState, *appmodel.CinephileResponse]
}

func curatorTemplate() *prompt.DefaultChatTemplate {
	return prompt.FromMessages(schema.GoTemplate,
		schema.UserMessage(`User wants: "{{.original_text}}"{{if .last_summary}}
Fact-checker verdict: {{.last_summary}}{{end}}
Current movie draft:
{{range .movies}}- "{{.Title}}" ({{.Year}}) — {{.Reason}}
{{end}}`),
	)
}

// NewCuratorChain compiles the curator chain and returns a ready-to-use CuratorChain.
func NewCuratorChain(ctx context.Context, m einomodel.ToolCallingChatModel) (*CuratorChain, error) {
	tmpl := curatorTemplate()

	parser := schema.NewMessageJSONParser[*appmodel.CinephileResponse](&schema.MessageJSONParseConfig{
		ParseFrom: schema.MessageParseFromContent,
	})

	runnable, err := compose.NewChain[*appmodel.FindMoviesState, *appmodel.CinephileResponse]().
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, state *appmodel.FindMoviesState) ([]*schema.Message, error) {
			msgs, err := tmpl.Format(ctx, map[string]any{
				"original_text": state.UserQuery.OriginalText,
				"last_summary":  state.LastSummary,
				"movies":        state.CurrentDraft,
			})
			if err != nil {
				return nil, err
			}
			return append([]*schema.Message{schema.SystemMessage(FinalSystemPrompt)}, msgs...), nil
		})).
		AppendChatModel(m).
		AppendLambda(compose.InvokableLambda(agent.StripMarkdownFences)).
		AppendLambda(compose.MessageParser(parser)).
		Compile(ctx)
	if err != nil {
		return nil, err
	}

	return &CuratorChain{runnable: runnable}, nil
}

func (s *CuratorChain) Invoke(ctx context.Context, state *appmodel.FindMoviesState) (*appmodel.FindMoviesState, error) {
	response, err := s.runnable.Invoke(ctx, state)
	if err != nil {
		return nil, fmt.Errorf("curator run: %w", err)
	}
	state.CurrentDraft = response.Movies
	return state, nil
}
