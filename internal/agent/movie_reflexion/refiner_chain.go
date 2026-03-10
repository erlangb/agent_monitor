package movie_reflexion

import (
	"context"
	"erlangb/agentmonitor/internal/agent"
	"log/slog"

	appmodel "erlangb/agentmonitor/internal/model"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const refinerSystemPrompt = "You are an expert movie librarian. " +
	"Extract search parameters from the user's request. " +
	"If the user says 'the 90s', set start_year to 1990 and end_year to 1999. " +
	"Identify primary_genre, secondary_genres, is_classic (true if query mentions classic/old/vintage). " +
	"Set original_text to the identical user input verbatim. " +
	"Set query_info to a short, search-engine-optimised string that captures the core intent — mood, genre, era — " +
	"in a form suitable for a web or database search (e.g. 'surreal horror underground 1990s'). " +
	"Respond with raw JSON only. No markdown, no code fences, no explanation."

// RefinerChain extracts structured search parameters from a free-text movie query.
// Construct with NewRefinerChain. Call Invoke directly or wrap with
// compose.InvokableLambda for graph wiring.
type RefinerChain struct {
	runnable compose.Runnable[string, *appmodel.RefinedMovieQuery]
}

func refinerTemplate() *prompt.DefaultChatTemplate {
	return prompt.FromMessages(schema.GoTemplate,
		schema.SystemMessage(refinerSystemPrompt),
		schema.UserMessage("{{.query}}"),
	)
}

// NewRefinerChain compiles the refiner chain and returns a ready-to-use RefinerChain.
func NewRefinerChain(ctx context.Context, m einomodel.ToolCallingChatModel) (*RefinerChain, error) {
	tmpl := refinerTemplate()

	parser := schema.NewMessageJSONParser[*appmodel.RefinedMovieQuery](&schema.MessageJSONParseConfig{
		ParseFrom: schema.MessageParseFromContent,
	})

	chain := compose.NewChain[string, *appmodel.RefinedMovieQuery]().
		AppendLambda(compose.InvokableLambda(func(ctx context.Context, input string) ([]*schema.Message, error) {
			return tmpl.Format(ctx, map[string]any{"query": input})
		})).
		AppendChatModel(m).
		AppendLambda(compose.InvokableLambda(agent.StripMarkdownFences)).
		AppendLambda(compose.MessageParser(parser))

	runnable, err := chain.Compile(ctx)
	if err != nil {
		return nil, err
	}

	return &RefinerChain{runnable: runnable}, nil
}

// Invoke runs the refiner chain and returns the structured query parameters.
func (s *RefinerChain) Invoke(ctx context.Context, query string) (*appmodel.RefinedMovieQuery, error) {
	result, err := s.runnable.Invoke(ctx, query)
	if err != nil {
		return nil, err
	}
	slog.Info("refiner output", "start_year", result.StartYear, "genre", result.PrimaryGenre)
	return result, nil
}
