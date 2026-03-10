package usecase

import (
	"context"
)

// UseCase is the single entry point for executing an agent query.
type UseCase interface {
	Run(ctx context.Context, input string) (string, error)
	Name() string
	// ExampleInput returns a ready-to-paste sample query displayed after selection.
	ExampleInput() string
	// Description returns a short human-readable summary shown in the selection menu.
	Description() string
}
