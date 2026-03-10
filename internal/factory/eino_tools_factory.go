package factory

import (
	"context"
	"erlangb/agentmonitor/internal/client"
	"fmt"

	"erlangb/agentmonitor/internal/config"
	"erlangb/agentmonitor/internal/tools"
)

const (
	// ToolSourceTavily is the config name for the Tavily MCP server.
	ToolSourceTavily = "tavily"
)

// EinoToolsFactory creates named client.EinoMcpClient or specialized tool application tool TavilyMCPTool
type EinoToolsFactory struct {
	cfg config.Config
}

// NewEinoToolsFactory returns a factory backed by the given config.
func NewEinoToolsFactory(cfg config.Config) *EinoToolsFactory {
	return &EinoToolsFactory{cfg: cfg}
}

// CreateMCPToolClient CreateMCPTools
// client.EinoMcpClient for a given server name from config
// EinoTools can be invoked from client
func (f *EinoToolsFactory) CreateMCPToolClient(ctx context.Context, serverName string) (client.EinoMcpClient, error) {
	srv, ok := f.findServer(serverName)
	if !ok {
		return nil, fmt.Errorf("mcp server %q not found in config", serverName)
	}
	return f.createEinoMcpClient(ctx, srv)
}

// CreateTavilySpecializedTool connects to the Tavily MCP server and returns specialized TavilyMCPTool
// typed-input tools via TavilyMCPTool instead of the raw generic tool list.
func (f *EinoToolsFactory) CreateTavilySpecializedTool(ctx context.Context) (*tools.TavilyMCPTool, error) {
	srv, ok := f.findServer(ToolSourceTavily)
	if !ok {
		return nil, fmt.Errorf("mcp server %q not found in config", ToolSourceTavily)
	}

	einoClient, err := f.createEinoMcpClient(ctx, srv)

	if err != nil {
		return nil, err
	}

	return tools.NewTavilyMCPTool(einoClient), nil
}

// CreateCommandMcpClient connects to the named stdio MCP binary from config and
// returns all tools it advertises as generic Eino tools.
func (f *EinoToolsFactory) CreateCommandMcpClient(ctx context.Context, commandName string) (client.EinoMcpClient, error) {
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
