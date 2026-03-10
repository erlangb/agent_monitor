package application

import (
	"context"
	"erlangb/agentmonitor/internal/config"
	"erlangb/agentmonitor/internal/factory"
	"erlangb/agentmonitor/internal/usecase"
	"erlangb/agentmonitor/internal/usecase/movie_reflexion"
	"erlangb/agentmonitor/internal/usecase/simple"
	"log/slog"

	"github.com/cloudwego/eino/callbacks"
	"github.com/erlangb/agentmeter"
	einometer "github.com/erlangb/agentmeter/adapters/eino"
	"github.com/erlangb/agentmeter/pricing"
)

// Application is the top-level object that owns factories, use cases, and the UI runner.
type Application struct {
	modelFactory *factory.EinoChatModelFactory
	toolsFactory *factory.EinoToolsFactory

	useCases   []usecase.UseCase
	agentMeter *agentmeter.Meter

	runner Runner
}

// NewApplication constructs the Application, initialises all use cases, and wires agentmeter callbacks.
func NewApplication(ctx context.Context, cfg config.Config, runner Runner) (*Application, error) {
	modelFactory := factory.NewChatModelFactory(cfg)
	toolsFactory := factory.NewEinoToolsFactory(cfg)

	costFn := pricing.WithDefaultPricing()
	agentMeter := agentmeter.New(costFn)

	useCases, err := loadUseCases(ctx, modelFactory, toolsFactory, agentMeter)

	if err != nil {
		return nil, err
	}

	return &Application{
		modelFactory: modelFactory,
		toolsFactory: toolsFactory,
		useCases:     useCases,
		agentMeter:   agentMeter,
		runner:       runner,
	}, nil
}

// Runner is the UI abstraction. Implementations drive the interaction loop (TUI or terminal).
type Runner interface {
	Run(ctx context.Context, useCases []usecase.UseCase, meter *agentmeter.Meter) error
	Close()
}

func (app *Application) Run(ctx context.Context) error {
	return app.runner.Run(ctx, app.useCases, app.agentMeter)
}

func (app *Application) Close() {
	slog.Default().Info("closing application")
	app.runner.Close()
}

func loadUseCases(ctx context.Context, modelFactory *factory.EinoChatModelFactory, toolsFactory *factory.EinoToolsFactory, meter *agentmeter.Meter) ([]usecase.UseCase, error) {
	agentMeterHandler := einometer.NewAgentMeterHandler(meter)
	// Attach agentMeter to eino use cases via callback handlers.
	opts := usecase.WithCallBackHandlers{
		Handlers: []callbacks.Handler{agentMeterHandler},
	}

	cinephileUseCase, err := movie_reflexion.NewCinephileUseCase(ctx, modelFactory, opts)
	if err != nil {
		return nil, err
	}

	findMoviesUseCase, err := movie_reflexion.NewFindMoviesUseCase(ctx, modelFactory, toolsFactory, opts)
	if err != nil {
		return nil, err
	}

	simpleLLmUseCase := simple.NewSimpleAgentLLMNoTools(modelFactory, opts)
	refinerQueryUseCase := movie_reflexion.NewRefinerQueryUseCase(modelFactory, opts)

	return []usecase.UseCase{
		simpleLLmUseCase,
		cinephileUseCase,
		findMoviesUseCase,
		refinerQueryUseCase,
	}, nil
}
