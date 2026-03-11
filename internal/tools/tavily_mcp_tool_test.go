package tools

import (
	"context"
	"erlangb/agentmonitor/internal/client/mocks"
	"errors"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// tavilyResultText is a minimal valid Tavily search response JSON that
// TavilySearch can parse end-to-end.
const tavilyResultText = `{"results":[{"content":"fresh tuna off the Azores","score":0.95}]}`

func TestTavilySearch(t *testing.T) {
	tests := []struct {
		name         string
		query        string
		mockResult   *mcp.CallToolResult
		mockErr      error
		wantErr      bool
		wantNonEmpty bool
	}{
		{
			name:         "success",
			query:        "tuna fishing spots",
			mockResult:   mcp.NewToolResultText(tavilyResultText),
			wantNonEmpty: true,
		},
		{
			name:    "client error propagates",
			query:   "tuna fishing spots",
			mockErr: errors.New("connection refused"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mc := mocks.NewMockMcpClient(t)
			mc.EXPECT().
				CallTool(context.Background(), "tavily_search", map[string]any{"query": tt.query, "max_results": 5}).
				Return(tt.mockResult, tt.mockErr)

			tool := NewTavilyMcp(mc)
			got, err := tool.TavilySearch(context.Background(), tt.query, 5)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "tavily_search call")
			} else {
				require.NoError(t, err)
				if tt.wantNonEmpty {
					assert.NotEmpty(t, got)
				}
			}
		})
	}
}

func TestTavilyClose(t *testing.T) {
	tests := []struct {
		name      string
		nilClient bool
		closeErr  error
		wantErr   bool
	}{
		{
			name:      "nil client returns no error",
			nilClient: true,
		},
		{
			name:     "client close error propagates",
			closeErr: errors.New("close failed"),
			wantErr:  true,
		},
		{
			name: "client close success",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var subject *TavilyMcp
			if tt.nilClient {
				subject = NewTavilyMcp(nil)
			} else {
				mc := mocks.NewMockMcpClient(t)
				mc.EXPECT().Close().Return(tt.closeErr)
				subject = NewTavilyMcp(mc)
			}

			err := subject.Close()
			if tt.wantErr {
				require.EqualError(t, err, tt.closeErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
