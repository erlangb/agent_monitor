package movie_reflexion

import (
	"context"
	"testing"

	appmodel "erlangb/agentmonitor/internal/model"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeChatModel returns a fixed response for every Generate call.
type fakeChatModel struct{ response string }

func (f *fakeChatModel) Generate(_ context.Context, _ []*schema.Message, _ ...model.Option) (*schema.Message, error) {
	return schema.AssistantMessage(f.response, nil), nil
}

func (f *fakeChatModel) Stream(_ context.Context, _ []*schema.Message, _ ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	return nil, nil
}

func (f *fakeChatModel) WithTools(_ []*schema.ToolInfo) (model.ToolCallingChatModel, error) {
	return f, nil
}

func TestCinephileAgent_Invoke(t *testing.T) {
	ctx := context.Background()

	validJSON := `{"movies":[{"title":"Eraserhead","year":1977,"reason":"surreal nightmare fuel"}]}`

	tests := []struct {
		name          string
		modelResponse string
		initialState  *appmodel.FindMoviesState
		check         func(t *testing.T, got *appmodel.FindMoviesState)
	}{
		{
			name:          "happy path: draft populated and retry incremented",
			modelResponse: validJSON,
			initialState:  &appmodel.FindMoviesState{UserQuery: appmodel.RefinedMovieQuery{OriginalText: "surreal horror"}},
			check: func(t *testing.T, got *appmodel.FindMoviesState) {
				assert.Empty(t, got.CritiqueHistory, "unexpected runnable error: %v", got.CritiqueHistory)
				require.Len(t, got.CurrentDraft, 1)
				assert.Equal(t, "Eraserhead", got.CurrentDraft[0].Title)
				assert.Equal(t, 1, got.RetryCount)
			},
		},
		{
			name:          "parse error: critique appended, retry incremented, no error returned",
			modelResponse: "not json at all",
			initialState:  &appmodel.FindMoviesState{UserQuery: appmodel.RefinedMovieQuery{OriginalText: "surreal horror"}},
			check: func(t *testing.T, got *appmodel.FindMoviesState) {
				assert.Equal(t, 1, got.RetryCount)
				assert.Len(t, got.CritiqueHistory, 1)
				assert.Empty(t, got.CurrentDraft)
			},
		},
		{
			name:          "empty state: no panic, retry incremented",
			modelResponse: validJSON,
			initialState:  &appmodel.FindMoviesState{},
			check: func(t *testing.T, got *appmodel.FindMoviesState) {
				assert.Equal(t, 1, got.RetryCount)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := NewCinephileAgent(ctx, &fakeChatModel{response: tt.modelResponse})
			require.NoError(t, err)

			got, err := node.Invoke(ctx, tt.initialState)
			require.NoError(t, err)
			tt.check(t, got)
		})
	}
}
