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

const CinephileSystemPrompt = `# ROLE: The Cinephile (Movie Suggester)
1. A good film recommendation surfaces what the user would never find on their own.
2. You never suggest anything with >100k IMDb ratings unless the user explicitly asks for something mainstream.

# I. PREVIOUS CRITICS
If there are previous critics to fix check them and reuse accepted movies on the output.

# I. HOW YOU RESPOND
- If the user gives you a genre, decade, or mood — dig into the real underground, not the acclaimed underground (no Bergman, no Tarkovsky unless specifically asked).
- For each give: title, year, one-line reason to why you picked this movie based on user request

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

// CinephileAgent is a graph-ready node that suggests underground movies.
// Construct with NewCinephileAgent. Call Invoke directly or wrap with
// compose.InvokableLambda for graph wiring.
type CinephileAgent struct {
	runnableFind compose.Runnable[*appmodel.FindMoviesState, *appmodel.CinephileResponse]
}

// NewCinephileAgent compiles the cinephile chain and returns a ready-to-use CinephileAgent.
func NewCinephileAgent(ctx context.Context, m einomodel.ToolCallingChatModel) (*CinephileAgent, error) {
	tmpl := cinephileTemplate()

	parser := agent.MessageParseFromContent[*appmodel.CinephileResponse]()

	runnableFind, err := compose.NewChain[*appmodel.FindMoviesState, *appmodel.CinephileResponse]().
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, state *appmodel.FindMoviesState) ([]*schema.Message, error) {
			userMsgs, err := tmpl.Format(ctx, map[string]any{
				"query":       state.UserQuery,
				"critiques":   state.CritiqueHistory,
				"retry_count": state.RetryCount,
			})
			if err != nil {
				return nil, err
			}
			return append([]*schema.Message{schema.SystemMessage(CinephileSystemPrompt)}, userMsgs...), nil
		})).
		AppendChatModel(m).
		AppendLambda(compose.InvokableLambda(agent.StripMarkdownFences)).
		AppendLambda(compose.MessageParser(parser)).
		Compile(ctx)
	if err != nil {
		return nil, err
	}

	return &CinephileAgent{runnableFind: runnableFind}, nil
}

func cinephileTemplate() *prompt.DefaultChatTemplate {
	return prompt.FromMessages(schema.GoTemplate,
		schema.UserMessage(`User wants: "{{.query.OriginalText}}"
{{if .query.PrimaryGenre}}Primary genre: {{.query.PrimaryGenre}}
{{end -}}{{if .query.Secondary}}Secondary genres: {{range $i, $g := .query.Secondary}}{{if $i}}, {{end}}{{$g}}{{end}}
{{end -}}{{if or .query.StartYear .query.EndYear}}Year range: {{.query.StartYear}}–{{.query.EndYear}}
{{end -}}{{if .query.IsClassic}}Style: classic films only
{{end}}[Query] {{.query.QueryInfo}}
{{if and .retry_count .critiques}}Previous critiques to fix:
{{range .critiques}}- {{.}}
{{end}}{{end}}Provide 3 movie suggestions in strict JSON format.`),
	)
}

// Invoke runs the cinephile chain, populates state.CurrentDraft, and increments RetryCount.
func (s *CinephileAgent) Invoke(ctx context.Context, state *appmodel.FindMoviesState) (*appmodel.FindMoviesState, error) {
	response, err := s.runnableFind.Invoke(ctx, state)
	if err != nil {
		state.RetryCount++
		state.CritiqueHistory = append(state.CritiqueHistory, fmt.Sprintf("Generator error: %v", err))
		return state, nil
	}
	state.CurrentDraft = response.Movies
	state.RetryCount++
	return state, nil
}
