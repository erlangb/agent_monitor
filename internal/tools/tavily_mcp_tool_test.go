package tools

import (
	"context"
	"erlangb/agentmonitor/internal/client/mocks"
	"errors"
	"testing"

	"github.com/cloudwego/eino/components/tool"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
			mockResult:   &mcp.CallToolResult{},
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
			mc := mocks.NewMockEinoMcpClient(t)
			mc.EXPECT().
				CallTool(context.Background(), "tavily_search", map[string]any{"query": tt.query}).
				Return(tt.mockResult, tt.mockErr)

			tool := NewTavilyMCPTool(mc)
			got, err := tool.TavilySearch(context.Background(), tt.query)

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "tavily_search call")
			} else {
				require.NoError(t, err)
				if tt.wantNonEmpty {
					assert.NotNil(t, got)
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
			var subject *TavilyMCPTool
			if tt.nilClient {
				subject = NewTavilyMCPTool(nil)
			} else {
				mc := mocks.NewMockEinoMcpClient(t)
				mc.EXPECT().Close().Return(tt.closeErr)
				subject = NewTavilyMCPTool(mc)
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

func TestGetTools(t *testing.T) {
	mc := mocks.NewMockEinoMcpClient(t)
	subject := NewTavilyMCPTool(mc)

	tools, err := subject.GetTools(context.Background())

	require.NoError(t, err)
	require.Len(t, tools, 1)
	assert.Implements(t, (*tool.BaseTool)(nil), tools[0])
}
