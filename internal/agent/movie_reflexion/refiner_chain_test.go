package movie_reflexion

import (
	"context"
	"testing"

	appmodel "erlangb/agentmonitor/internal/model"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func surreal90sJSON() string {
	return `{
		"primary_genre":  "horror",
		"secondary_genres": ["surreal"],
		"start_year":     1990,
		"end_year":       1999,
		"is_classic":     false,
		"original_text":  "surreal horror films from the 90s",
		"query_info":     "surreal horror underground 1990s"
	}`
}

func classicNoirFencedJSON() string {
	return "```json\n" + `{
		"primary_genre":    "noir",
		"secondary_genres": [],
		"start_year":       0,
		"end_year":         0,
		"is_classic":       true,
		"original_text":    "old classic noir",
		"query_info":       "classic noir"
	}` + "\n```"
}

func TestRefinerChain_Invoke(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		modelResponse string
		query         string
		wantErr       bool
		check         func(t *testing.T, got *appmodel.RefinedMovieQuery)
	}{
		{
			name:          "full structured response",
			query:         "surreal horror films from the 90s",
			modelResponse: surreal90sJSON(),
			check: func(t *testing.T, got *appmodel.RefinedMovieQuery) {
				assert.Equal(t, "horror", got.PrimaryGenre)
				assert.Equal(t, []string{"surreal"}, got.Secondary)
				assert.Equal(t, 1990, got.StartYear)
				assert.Equal(t, 1999, got.EndYear)
				assert.False(t, got.IsClassic)
				assert.Equal(t, "surreal horror films from the 90s", got.OriginalText)
				assert.Equal(t, "surreal horror underground 1990s", got.QueryInfo)
			},
		},
		{
			name:          "model wraps json in markdown fences",
			query:         "old classic noir",
			modelResponse: classicNoirFencedJSON(),
			check: func(t *testing.T, got *appmodel.RefinedMovieQuery) {
				assert.Equal(t, "noir", got.PrimaryGenre)
				assert.True(t, got.IsClassic)
				assert.Equal(t, "old classic noir", got.OriginalText)
			},
		},
		{
			name:          "invalid json propagates error",
			query:         "something",
			modelResponse: "not json",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain, err := NewRefinerChain(ctx, &fakeChatModel{response: tt.modelResponse})
			require.NoError(t, err)

			got, err := chain.Invoke(ctx, tt.query)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			tt.check(t, got)
		})
	}
}