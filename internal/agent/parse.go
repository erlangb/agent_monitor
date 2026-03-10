package agent

import (
	"context"
	"strings"

	"github.com/cloudwego/eino/schema"
)

// StripMarkdownFences removes ```json ... ``` or ``` ... ``` wrappers that some
// models add around JSON output, returning a plain-content message for the parser.
func StripMarkdownFences(_ context.Context, msg *schema.Message) (*schema.Message, error) {
	if msg != nil {
		cleanJSON := strings.TrimSpace(msg.Content)
		cleanJSON = strings.TrimPrefix(cleanJSON, "```json")
		cleanJSON = strings.TrimSuffix(cleanJSON, "```")
		return &schema.Message{Role: msg.Role, Content: cleanJSON}, nil
	}

	return nil, nil
}

func MessageParseFromContent[T any]() schema.MessageParser[T] {
	return schema.NewMessageJSONParser[T](&schema.MessageJSONParseConfig{
		ParseFrom: schema.MessageParseFromContent,
	})
}
