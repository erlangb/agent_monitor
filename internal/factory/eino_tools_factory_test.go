package factory

import (
	"context"
	"testing"

	"erlangb/agentmonitor/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateMCPToolClient_ServerNotFound(t *testing.T) {
	f := factoryWithServers()
	_, err := f.CreateMCPToolClient(context.Background(), "missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"missing" not found in config`)
}

func TestCreateTavilySpecializedTool_ServerNotFound(t *testing.T) {
	f := factoryWithServers()
	_, err := f.CreateTavilySpecializedTool(context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"tavily" not found in config`)
}

func TestCreateCommandMcpClient_CommandNotFound(t *testing.T) {
	f := factoryWithCommands()
	_, err := f.CreateCommandMcpClient(context.Background(), "missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `"missing" not found in config`)
}

func factoryWithServers(servers ...config.McpServerConfig) *EinoToolsFactory {
	return NewEinoToolsFactory(config.Config{
		Tools: config.ToolsConfig{McpServers: servers},
	})
}

func factoryWithCommands(cmds ...config.BinCommand) *EinoToolsFactory {
	return NewEinoToolsFactory(config.Config{
		Tools: config.ToolsConfig{Commands: cmds},
	})
}
