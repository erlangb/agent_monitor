package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	toolmcp "github.com/cloudwego/eino-ext/components/tool/mcp"
	etool "github.com/cloudwego/eino/components/tool"
	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	transportSSEClient            = "sse"
	transportStreamableHTTPClient = "streamablehttp"
)

// McpClient is the interface consumers depend on for MCP interactions.
type McpClient interface {
	CallTool(ctx context.Context, name string, args map[string]any) (*mcp.CallToolResult, error)
	Close() error
}

// options holds the configuration for building a Client.
type options struct {
	url        string
	commandBin string
	transport  string
	headers    map[string]string
}

// Option configures a Client.
type Option func(*options)

// WithURL sets the server URL.
func WithURL(url string) Option {
	return func(o *options) { o.url = url }
}

// WithCommandBin sets the binary path for stdio transport.
// When set, it takes precedence over URL-based transports.
func WithCommandBin(bin string) Option {
	return func(o *options) { o.commandBin = bin }
}

// WithTransport sets the transport type ("sse" or "streamablehttp").
// Defaults to "sse" when not set.
func WithTransport(t string) Option {
	return func(o *options) { o.transport = strings.ToLower(strings.TrimSpace(t)) }
}

// WithHeaders sets HTTP headers for SSE and StreamableHTTP transports.
// Values may reference $ENV_VARs — they are expanded at dial time.
func WithHeaders(h map[string]string) Option {
	return func(o *options) { o.headers = h }
}

// WithBearerAuth injects an Authorization: Bearer header.
func WithBearerAuth(token string) Option {
	return func(o *options) {
		if token == "" {
			return
		}
		if o.headers == nil {
			o.headers = make(map[string]string)
		}
		o.headers["Authorization"] = "Bearer " + token
	}
}

// validate ensures exactly one of commandBin or url is set.
func (o *options) validate() error {
	if o.commandBin != "" && o.url != "" {
		return fmt.Errorf("mcp: commandBin and url are mutually exclusive")
	}
	if o.commandBin == "" && o.url == "" {
		return fmt.Errorf("mcp: either commandBin or url must be set")
	}
	return nil
}

// Client wraps an mcp-go connection.
type Client struct {
	inner *mcpclient.Client
}

func (c *Client) CallTool(ctx context.Context, name string, args map[string]any) (*mcp.CallToolResult, error) {
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args
	return c.inner.CallTool(ctx, req)
}

// NewClient builds, connects, and initializes an MCP Client.
func NewClient(ctx context.Context, opts ...Option) (*Client, error) {
	o := &options{transport: transportSSEClient}
	for _, opt := range opts {
		opt(o)
	}

	if err := o.validate(); err != nil {
		return nil, err
	}

	raw, err := dial(ctx, o)
	if err != nil {
		return nil, err
	}

	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = "2024-11-05"
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    "mcp_client_agent",
		Version: "0.1.0",
	}
	if _, err := raw.Initialize(ctx, initReq); err != nil {
		_ = raw.Close()
		return nil, fmt.Errorf("mcp initialize: %w", err)
	}

	return &Client{inner: raw}, nil
}

// Close shuts down the underlying MCP connection.
func (c *Client) Close() error {
	return c.inner.Close()
}

// dial creates and starts the raw mcp-go client based on options.
// Precondition: o.validate() has already been called.
func dial(ctx context.Context, o *options) (*mcpclient.Client, error) {
	if o.commandBin != "" {
		return mcpclient.NewStdioMCPClient(o.commandBin, []string{})
	}

	headers := expandHeaders(o.headers)

	var (
		cli *mcpclient.Client
		err error
	)
	switch o.transport {
	case transportStreamableHTTPClient:
		cli, err = mcpclient.NewStreamableHttpClient(o.url, transport.WithHTTPHeaders(headers))
	default:
		cli, err = mcpclient.NewSSEMCPClient(o.url, transport.WithHeaders(headers))
	}
	if err != nil {
		return nil, err
	}
	if err := cli.Start(ctx); err != nil {
		_ = cli.Close()
		return nil, err
	}
	return cli, nil
}

// resultHandler handle tool result
func (c *Client) resultHandler(ctx context.Context, name string, result *mcp.CallToolResult) (*mcp.CallToolResult, error) {
	if result == nil {
		return result, errors.New("mcp error nil result")
	}
	if result.IsError {
		slog.Info("mcp-go: tool-call-result:", slog.String("name", name), slog.Bool("error", result.IsError))
		return result, errors.New("mcp error")
	}

	return result, nil
}

// GetEinoTools returns the named tools as Eino BaseTool instances.
func (c *Client) GetEinoTools(ctx context.Context, toolNames []string) ([]etool.BaseTool, error) {
	return toolmcp.GetTools(ctx, &toolmcp.Config{
		Cli:                   c.inner,
		ToolNameList:          toolNames,
		ToolCallResultHandler: c.resultHandler,
	})
}

// expandHeaders expands $ENV_VAR references in header values.
func expandHeaders(raw map[string]string) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	out := make(map[string]string, len(raw))
	for k, v := range raw {
		out[k] = os.ExpandEnv(v)
	}
	return out
}
