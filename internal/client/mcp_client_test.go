package client

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOptionsValidate(t *testing.T) {
	tests := []struct {
		name       string
		commandBin string
		url        string
		wantErr    string
	}{
		{
			name:    "neither set",
			wantErr: "mcp: either commandBin or url must be set",
		},
		{
			name:       "both set",
			commandBin: "/usr/bin/osmmcp",
			url:        "http://localhost:8080",
			wantErr:    "mcp: commandBin and url are mutually exclusive",
		},
		{
			name:       "only commandBin",
			commandBin: "/usr/bin/osmmcp",
		},
		{
			name: "only url",
			url:  "http://localhost:8080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &options{commandBin: tt.commandBin, url: tt.url}
			err := o.validate()
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestExpandHeaders(t *testing.T) {
	t.Setenv("TEST_API_KEY", "secret123")

	tests := []struct {
		name string
		in   map[string]string
		want map[string]string
	}{
		{
			name: "nil input",
			in:   nil,
			want: nil,
		},
		{
			name: "empty map",
			in:   map[string]string{},
			want: nil,
		},
		{
			name: "plain values",
			in:   map[string]string{"X-App": "fishing"},
			want: map[string]string{"X-App": "fishing"},
		},
		{
			name: "env var expansion",
			in:   map[string]string{"Authorization": "Bearer $TEST_API_KEY"},
			want: map[string]string{"Authorization": "Bearer secret123"},
		},
		{
			name: "mixed plain and env var",
			in:   map[string]string{"X-App": "fishing", "Authorization": "Bearer $TEST_API_KEY"},
			want: map[string]string{"X-App": "fishing", "Authorization": "Bearer secret123"},
		},
		{
			name: "unknown env var expands to empty",
			in:   map[string]string{"X-Token": "$UNSET_VAR_XYZ"},
			want: map[string]string{"X-Token": os.ExpandEnv("$UNSET_VAR_XYZ")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandHeaders(tt.in)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOptionBuilders(t *testing.T) {
	tests := []struct {
		name string
		opts []Option
		want options
	}{
		{
			name: "WithURL",
			opts: []Option{WithURL("http://localhost:9090")},
			want: options{url: "http://localhost:9090", transport: transportSSEClient},
		},
		{
			name: "WithCommandBin",
			opts: []Option{WithCommandBin("/usr/local/bin/osmmcp")},
			want: options{commandBin: "/usr/local/bin/osmmcp", transport: transportSSEClient},
		},
		{
			name: "WithTransport streamablehttp",
			opts: []Option{WithURL("http://localhost:9090"), WithTransport("streamablehttp")},
			want: options{url: "http://localhost:9090", transport: transportStreamableHTTPClient},
		},
		{
			name: "WithTransport trims and lowercases",
			opts: []Option{WithURL("http://localhost:9090"), WithTransport("  SSE  ")},
			want: options{url: "http://localhost:9090", transport: transportSSEClient},
		},
		{
			name: "WithHeaders",
			opts: []Option{WithURL("http://localhost:9090"), WithHeaders(map[string]string{"X-App": "test"})},
			want: options{url: "http://localhost:9090", transport: transportSSEClient, headers: map[string]string{"X-App": "test"}},
		},
		{
			name: "WithBearerAuth sets Authorization header",
			opts: []Option{WithURL("http://localhost:9090"), WithBearerAuth("my-token")},
			want: options{url: "http://localhost:9090", transport: transportSSEClient, headers: map[string]string{"Authorization": "Bearer my-token"}},
		},
		{
			name: "WithBearerAuth empty token is no-op",
			opts: []Option{WithURL("http://localhost:9090"), WithBearerAuth("")},
			want: options{url: "http://localhost:9090", transport: transportSSEClient},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := &options{transport: transportSSEClient}
			for _, opt := range tt.opts {
				opt(o)
			}
			assert.Equal(t, tt.want, *o)
		})
	}
}

func TestNewClient_ValidationErrors(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		wantErr string
	}{
		{
			name:    "no options",
			opts:    nil,
			wantErr: "mcp: either commandBin or url must be set",
		},
		{
			name:    "both commandBin and url",
			opts:    []Option{WithCommandBin("/bin/foo"), WithURL("http://localhost:8080")},
			wantErr: "mcp: commandBin and url are mutually exclusive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(context.Background(), tt.opts...)
			require.EqualError(t, err, tt.wantErr)
		})
	}
}
