package factory

import (
	"context"
	"testing"

	"erlangb/agentmonitor/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateOpenAI_ReturnsModel(t *testing.T) {
	f := chatModelFactory()
	m, err := f.CreateOpenAI(context.Background(), "gpt-4o", 0.7, 512)
	require.NoError(t, err)
	assert.NotNil(t, m)
}

func TestCreateOpenAI_ZeroMaxTokens(t *testing.T) {
	f := chatModelFactory()
	m, err := f.CreateOpenAI(context.Background(), "gpt-4o", 0.7, 0)
	require.NoError(t, err)
	assert.NotNil(t, m)
}
func chatModelFactory() *EinoChatModelFactory {
	return NewChatModelFactory(config.Config{
		Models: map[string]config.ModelEntry{
			OPENAI: {APIKey: "test-key", ModelID: "gpt-4o"},
		},
	})
}
