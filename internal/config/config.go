// Package config provides configuration loading via Koanf (YAML + env).
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env/v2"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

const (
	envPrefix = "APP_ENV_"
	envDelim  = "."
)

// AppConfig holds application identity (name, env, version).
type AppConfig struct {
	Name    string `koanf:"name"`
	Env     string `koanf:"env"`
	Version string `koanf:"version"`
}

// ModelEntry defines a named LLM model configuration.
// The map key is the model name (e.g. "openai").
type ModelEntry struct {
	Provider string `koanf:"provider"`
	APIKey   string `koanf:"api_key"`
	ModelID  string `koanf:"model_id"`
}

// UseCaseEntry holds per-use-case model overrides.
type UseCaseEntry struct {
	Model       string  `koanf:"model"`
	ModelID     string  `koanf:"model_id"`
	Temperature float64 `koanf:"temperature"`
	MaxTokens   int     `koanf:"max_tokens"`
}

// McpServerConfig holds a single MCP server entry.
// Transport: "sse" (default) or "streamableHttp". Empty is treated as sse.
// Headers are sent with every request; values support $ENV_VAR expansion.
type McpServerConfig struct {
	Name      string            `koanf:"name"`
	URL       string            `koanf:"url"`
	Transport string            `koanf:"transport"`
	APIKey    string            `koanf:"api_key"`
	Headers   map[string]string `koanf:"headers"`
}

// ToolsConfig holds generic tool/MCP configuration. Open-ended for future MCP integration.
type ToolsConfig struct {
	McpServers []McpServerConfig `koanf:"mcp_servers"`
	Commands   []BinCommand      `koanf:"commands"`
}

// BinCommand holds configuration for the local OSM geocoding resolver.
// BinCommand is the path to the osmmcp binary; leave empty to disable geo enrichment.
type BinCommand struct {
	Name    string `koanf:"name"`
	CmdPath string `koanf:"cmd_path"`
}

// Config is the root application configuration.
type Config struct {
	App      AppConfig               `koanf:"app"`
	Models   map[string]ModelEntry   `koanf:"models"`
	UseCases map[string]UseCaseEntry `koanf:"use_cases"`
	Tools    ToolsConfig             `koanf:"tools"`
}

// LoadConfig loads configuration from the given YAML path and overrides from env.
//
// Env vars must use prefix APP_ENV_ and double-underscore (__) as the nesting
// separator. Single underscores are preserved as part of key names.
//
// Examples:
//
//	APP_ENV_APP__NAME                      → app.name
//	APP_ENV_MODELS__OPENAI__API_KEY        → models.openai.api_key
func LoadConfig(path string) (*Config, error) {
	k := koanf.New(".")

	// 1. Load YAML file as base.
	if path != "" {
		if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
			return nil, fmt.Errorf("load config file %q: %w", path, err)
		}
	}

	// 2. Override with environment variables (APP_ENV_*).
	envProvider := env.Provider(envDelim, env.Opt{
		Prefix: envPrefix,
		TransformFunc: func(k, v string) (string, any) {
			return envKeyTransform(k), v
		},
	})
	if err := k.Load(envProvider, nil); err != nil {
		return nil, fmt.Errorf("load env config: %w", err)
	}

	var cfg Config
	if err := k.Unmarshal("", &cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	cfg.expandEnvVars()
	return &cfg, nil
}

// expandEnvVars resolves $ENV_VAR references in config fields that support them.
// Called once after unmarshal so the rest of the app works with clean values.
func (c *Config) expandEnvVars() {
	for name, m := range c.Models {
		m.APIKey = os.ExpandEnv(m.APIKey)
		c.Models[name] = m
	}
	for i, srv := range c.Tools.McpServers {
		c.Tools.McpServers[i].URL = os.ExpandEnv(srv.URL)
		c.Tools.McpServers[i].APIKey = os.ExpandEnv(srv.APIKey)
	}
}

// envKeyTransform strips APP_ENV_ prefix, lowercases, and converts double
// underscores to dots. Single underscores are kept intact so that key names
//
//	APP_ENV_APP__NAME                → app.name
//	APP_ENV_MODELS__OPENAI__API_KEY  → models.openai.api_key
func envKeyTransform(s string) string {
	s = strings.TrimPrefix(s, envPrefix)
	if s == "" {
		return ""
	}
	s = strings.ToLower(s)
	return strings.ReplaceAll(s, "__", ".")
}
