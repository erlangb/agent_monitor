package tools

import (
	"context"
	tool_mcp "erlangb/agentmonitor/internal/client"
	"erlangb/agentmonitor/internal/model"
	"fmt"

	"github.com/bytedance/sonic"
)

// TavilyMcp wraps a McpClient to expose Tavily search as a domain tool.
type TavilyMcp struct {
	client tool_mcp.McpClient
}

// NewTavilyMcp wraps a McpClient and returns a TavilyMcp.
func NewTavilyMcp(c tool_mcp.McpClient) *TavilyMcp {
	return &TavilyMcp{client: c}
}

func (t *TavilyMcp) Close() error {
	if t.client == nil {
		return nil
	}
	return t.client.Close()
}

// TavilySearch calls the Tavily MCP tool and returns a JSON string containing only
// content+score per result (Tolerant Reader). The MCP envelope is unwrapped and the
// Tavily payload is re-encoded in one pass — no intermediate struct round-trip.
func (t *TavilyMcp) TavilySearch(ctx context.Context, query string, maxResults int) (string, error) {
	result, err := t.client.CallTool(ctx, "tavily_search", map[string]any{"query": query, "max_results": maxResults})
	if err != nil {
		return "", fmt.Errorf("tavily_search call: %w", err)
	}

	mcpRaw, err := result.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("tavily_search marshal: %w", err)
	}

	var envelope struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := sonic.Unmarshal(mcpRaw, &envelope); err != nil {
		return "", fmt.Errorf("tavily_search unmarshal envelope: %w", err)
	}
	if len(envelope.Content) == 0 {
		return "", fmt.Errorf("tavily_search: empty content")
	}

	var resp model.TavilySearchResponse
	if err := sonic.UnmarshalString(envelope.Content[0].Text, &resp); err != nil {
		return "", fmt.Errorf("tavily_search unmarshal response: %w", err)
	}

	out, err := sonic.Marshal(resp.Results)
	if err != nil {
		return "", fmt.Errorf("tavily_search marshal results: %w", err)
	}
	return string(out), nil
}

