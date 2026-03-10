package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExpandEnvVars(t *testing.T) {
	t.Setenv("TEST_OPENAI_KEY", "sk-test")
	t.Setenv("TEST_TAVILY_KEY", "tvly-test")
	t.Setenv("TEST_MCP_URL", "https://mcp.example.com/")

	cfg := &Config{
		Models: map[string]ModelEntry{
			"openai": {APIKey: "$TEST_OPENAI_KEY"},
		},
		Tools: ToolsConfig{
			McpServers: []McpServerConfig{
				{Name: "tavily", URL: "$TEST_MCP_URL", APIKey: "$TEST_TAVILY_KEY"},
			},
		},
	}

	cfg.expandEnvVars()

	assert.Equal(t, "sk-test", cfg.Models["openai"].APIKey)
	assert.Equal(t, "https://mcp.example.com/", cfg.Tools.McpServers[0].URL)
	assert.Equal(t, "tvly-test", cfg.Tools.McpServers[0].APIKey)
}

func TestExpandEnvVars_UnsetVarBecomesEmpty(t *testing.T) {
	cfg := &Config{
		Models: map[string]ModelEntry{
			"openai": {APIKey: "$DEFINITELY_UNSET_VAR_XYZ"},
		},
	}
	cfg.expandEnvVars()
	assert.Equal(t, "", cfg.Models["openai"].APIKey)
}
