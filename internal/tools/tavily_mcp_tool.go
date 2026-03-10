package tools

import (
	"context"
	tool_mcp "erlangb/agentmonitor/internal/client"
	"erlangb/agentmonitor/internal/model"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"
)

// TavilyMCPTool wraps an EinoMcpClient to expose Tavily search as an Eino tool.
type TavilyMCPTool struct {
	client tool_mcp.EinoMcpClient
}

// NewTavilyMCPTool wraps an EinoMcpClient and returns a TavilyMCPTool.
func NewTavilyMCPTool(c tool_mcp.EinoMcpClient) *TavilyMCPTool {
	return &TavilyMCPTool{client: c}
}

func (t *TavilyMCPTool) Close() error {
	if t.client == nil {
		return nil
	}
	return t.client.Close()
}

func (t *TavilyMCPTool) TavilySearch(ctx context.Context, query string) ([]byte, error) {
	result, err := t.client.CallTool(ctx, "tavily_search", map[string]any{"query": query})
	if err != nil {
		return nil, fmt.Errorf("tavily_search call: %w", err)
	}
	return result.MarshalJSON()
}

func (t *TavilyMCPTool) asTavilySearchTool() (tool.BaseTool, error) {
	return toolutils.InferTool[model.TavilyQuery, string](
		"tavily_search",
		"Search on tavily",
		func(ctx context.Context, input model.TavilyQuery) (string, error) {
			cont, err := t.TavilySearch(ctx, input.Query)
			return string(cont), err
		},
	)
}

// GetTools returns all Eino tools provided by this MCP wrapper.
func (t *TavilyMCPTool) GetTools(ctx context.Context) ([]tool.BaseTool, error) {
	// More tools from client can be added here.

	tavilyTool, err := t.asTavilySearchTool()
	return []tool.BaseTool{tavilyTool}, err
}
