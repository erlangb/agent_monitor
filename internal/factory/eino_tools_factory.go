package factory

import (
	"context"
	"erlangb/agentmonitor/internal/client"
	"erlangb/agentmonitor/internal/model"
	"erlangb/agentmonitor/internal/tools"
	"fmt"

	etool "github.com/cloudwego/eino/components/tool"
	toolutils "github.com/cloudwego/eino/components/tool/utils"

	"erlangb/agentmonitor/internal/config"
)

const (
	// ToolSourceTavily is the config name for the Tavily MCP server.
	ToolSourceTavily = "tavily"
)

// EinoToolsFactory creates named client.McpClient or specialized Eino tools.
type EinoToolsFactory struct {
	cfg config.Config
}

// NewEinoToolsFactory returns a factory backed by the given config.
func NewEinoToolsFactory(cfg config.Config) *EinoToolsFactory {
	return &EinoToolsFactory{cfg: cfg}
}

// CreateMCPToolClient returns a client.McpClient for a given server name from config.
func (f *EinoToolsFactory) CreateMCPToolClient(ctx context.Context, serverName string) (client.McpClient, error) {
	srv, ok := f.findServer(serverName)
	if !ok {
		return nil, fmt.Errorf("mcp server %q not found in config", serverName)
	}
	return f.createEinoMcpClient(ctx, srv)
}

// CreateCommandMcpClient connects to the named stdio MCP binary from config and
// returns all tools it advertises as generic Eino tools.
func (f *EinoToolsFactory) CreateCommandMcpClient(ctx context.Context, commandName string) (client.McpClient, error) {
	cmd, ok := f.findCommand(commandName)
	if !ok {
		return nil, fmt.Errorf("mcp command %q not found in config", commandName)
	}

	mcpClient, err := client.NewClient(ctx, client.WithCommandBin(cmd.CmdPath))
	if err != nil {
		return nil, fmt.Errorf("connect mcp command %q: %w", commandName, err)
	}

	return mcpClient, nil
}

// CreateTavilyRawEinoTools creates the Tavily MCP client and returns its tools
// as raw Eino tools (full Tavily response). Eino tool creation is handled here.
func (f *EinoToolsFactory) CreateTavilyRawEinoTools(ctx context.Context) ([]etool.BaseTool, error) {
	mc, err := f.createTavilyMcpClient(ctx)
	if err != nil {
		return nil, err
	}
	return mc.GetEinoTools(ctx, []string{"tavily_search"})
}

// CreateTavilyParsedEinoTools creates the Tavily MCP client, wraps it with domain
// parsing logic (content+score filter), and returns an Eino tool via InferTool.
func (f *EinoToolsFactory) CreateTavilyParsedEinoTools(ctx context.Context) ([]etool.BaseTool, error) {
	mc, err := f.createTavilyMcpClient(ctx)
	if err != nil {
		return nil, err
	}
	tavilyMcp := tools.NewTavilyMcp(mc)
	t, err := toolutils.InferTool[model.TavilyQuery, string](
		"tavily_search",
		"Search on tavily",
		func(ctx context.Context, input model.TavilyQuery) (string, error) {
			return tavilyMcp.TavilySearch(ctx, input.Query, input.MaxResults)
		},
	)
	if err != nil {
		return nil, err
	}
	return []etool.BaseTool{t}, nil
}

// createTavilyMcpClient is a convenience helper for Tavily-specific methods.
func (f *EinoToolsFactory) createTavilyMcpClient(ctx context.Context) (*client.Client, error) {
	srv, ok := f.findServer(ToolSourceTavily)
	if !ok {
		return nil, fmt.Errorf("mcp server %q not found in config", ToolSourceTavily)
	}
	return f.createEinoMcpClient(ctx, srv)
}

func (f *EinoToolsFactory) createEinoMcpClient(ctx context.Context, srv config.McpServerConfig) (*client.Client, error) {
	opts := []client.Option{
		client.WithURL(srv.URL),
		client.WithTransport(srv.Transport),
		client.WithHeaders(srv.Headers),
	}
	if srv.APIKey != "" {
		opts = append(opts, client.WithBearerAuth(srv.APIKey))
	}

	einoMcpClient, err := client.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("connect mcp server %q: %w", srv.Name, err)
	}
	return einoMcpClient, nil
}

func (f *EinoToolsFactory) findServer(name string) (config.McpServerConfig, bool) {
	for _, srv := range f.cfg.Tools.McpServers {
		if srv.Name == name {
			return srv, true
		}
	}
	return config.McpServerConfig{}, false
}

func (f *EinoToolsFactory) findCommand(name string) (config.BinCommand, bool) {
	for _, cmd := range f.cfg.Tools.Commands {
		if cmd.Name == name {
			return cmd, true
		}
	}
	return config.BinCommand{}, false
}
