package movie_reflexion

import (
	"context"
	"testing"

	appmodel "erlangb/agentmonitor/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseClerkResponse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(t *testing.T, got *appmodel.ClerkResponse)
	}{
		{
			name:  "satisfied with no critiques",
			input: `{"critiques":[],"isSatisfied":true,"summary":"All good"}`,
			check: func(t *testing.T, got *appmodel.ClerkResponse) {
				assert.True(t, got.IsSatisfied)
				assert.Equal(t, "All good", got.Summary)
				assert.Empty(t, got.Critiques)
			},
		},
		{
			name:  "not satisfied with critiques",
			input: `{"critiques":["wrong year","director mismatch"],"isSatisfied":false,"summary":"Issues found"}`,
			check: func(t *testing.T, got *appmodel.ClerkResponse) {
				assert.False(t, got.IsSatisfied)
				assert.Equal(t, "Issues found", got.Summary)
				assert.Equal(t, []string{"wrong year", "director mismatch"}, got.Critiques)
			},
		},
		{
			name:    "invalid json returns error",
			input:   "not json at all",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseClerkResponse(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			tt.check(t, &got)
		})
	}
}

func TestClerkAgent_Invoke(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		modelResponse string
		initialState  *appmodel.FindMoviesState
		check         func(t *testing.T, got *appmodel.FindMoviesState)
	}{
		{
			name:          "satisfied: isSatisfied set, summary and critique appended",
			modelResponse: `{"critiques":[],"isSatisfied":true,"summary":"All movies verified"}`,
			initialState: &appmodel.FindMoviesState{
				UserQuery:    appmodel.RefinedMovieQuery{OriginalText: "surreal horror"},
				CurrentDraft: []appmodel.Movie{{Title: "Eraserhead", Year: 1977, Reason: "surreal"}},
			},
			check: func(t *testing.T, got *appmodel.FindMoviesState) {
				assert.True(t, got.IsSatisfied)
				assert.Equal(t, "All movies verified", got.LastSummary)
				require.Len(t, got.CritiqueHistory, 1)
				assert.Equal(t, "All movies verified", got.CritiqueHistory[0])
			},
		},
		{
			name:          "not satisfied: isSatisfied false, summary stored",
			modelResponse: `{"critiques":["wrong year"],"isSatisfied":false,"summary":"Year mismatch"}`,
			initialState: &appmodel.FindMoviesState{
				UserQuery:    appmodel.RefinedMovieQuery{OriginalText: "surreal horror"},
				CurrentDraft: []appmodel.Movie{{Title: "Eraserhead", Year: 1977, Reason: "surreal"}},
			},
			check: func(t *testing.T, got *appmodel.FindMoviesState) {
				assert.False(t, got.IsSatisfied)
				assert.Equal(t, "Year mismatch", got.LastSummary)
				require.Len(t, got.CritiqueHistory, 1)
				assert.Equal(t, "Year mismatch", got.CritiqueHistory[0])
			},
		},
		{
			name:          "parse error: raw answer appended to critique, no error returned",
			modelResponse: "not json",
			initialState: &appmodel.FindMoviesState{
				CurrentDraft: []appmodel.Movie{{Title: "Eraserhead", Year: 1977}},
			},
			check: func(t *testing.T, got *appmodel.FindMoviesState) {
				require.Len(t, got.CritiqueHistory, 1)
				assert.Equal(t, "not json", got.CritiqueHistory[0])
				assert.False(t, got.IsSatisfied)
				assert.Empty(t, got.LastSummary)
			},
		},
		{
			name:          "empty draft: no panic, parse error path",
			modelResponse: "bad",
			initialState:  &appmodel.FindMoviesState{},
			check: func(t *testing.T, got *appmodel.FindMoviesState) {
				assert.Len(t, got.CritiqueHistory, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clerk, err := NewClerkAgent(ctx, &fakeChatModel{response: tt.modelResponse}, nil)
			require.NoError(t, err)

			got, err := clerk.Invoke(ctx, tt.initialState)
			require.NoError(t, err)
			tt.check(t, got)
		})
	}
}