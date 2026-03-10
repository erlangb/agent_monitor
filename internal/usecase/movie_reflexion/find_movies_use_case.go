package movie_reflexion

import (
	"context"
	"erlangb/agentmonitor/internal/agent/movie_reflexion"
	"erlangb/agentmonitor/internal/factory"
	appmodel "erlangb/agentmonitor/internal/model"
	"erlangb/agentmonitor/internal/usecase"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/eino/compose"
	"github.com/mark3labs/mcp-go/mcp"
)

// FindMoviesUseCase orchestrates the full pipeline: refine → cinephile → clerk loop → curator.
type FindMoviesUseCase struct {
	usecase.BaseEinoUseCase
	refiner  *movie_reflexion.RefinerChain
	runnable compose.Runnable[*appmodel.FindMoviesState, *appmodel.FindMoviesState]
}

// NewFindMoviesUseCase builds and wires all pipeline nodes, connects to Tavily MCP, and returns a FindMoviesUseCase.
func NewFindMoviesUseCase(ctx context.Context, modelFactory *factory.EinoChatModelFactory, toolsFactory *factory.EinoToolsFactory, opts ...usecase.Option) (usecase.UseCase, error) {
	refinerModel, err := modelFactory.CreateOpenAI(ctx, "gpt-4.1-mini", 0.2, 1024)
	if err != nil {
		return nil, err
	}

	cinephileModel, err := modelFactory.CreateOpenAI(ctx, "gpt-4.1-mini", 0.7, 2048)
	if err != nil {
		return nil, err
	}

	clerkModel, err := modelFactory.CreateOpenAI(ctx, "gpt-4.1-mini", 0.2, 2048)
	if err != nil {
		return nil, err
	}

	curatorModel, err := modelFactory.CreateOpenAI(ctx, "gpt-4.1-mini", 0.1, 2048)
	if err != nil {
		return nil, err
	}

	tavilyClient, err := toolsFactory.CreateMCPToolClient(ctx, factory.ToolSourceTavily)
	if err != nil {
		return nil, err
	}

	clerkTools, err := tavilyClient.EinoTools(ctx, []mcp.Tool{{Name: "tavily_search"}})
	if err != nil {
		return nil, err
	}

	refinerNode, err := movie_reflexion.NewRefinerChain(ctx, refinerModel)
	if err != nil {
		return nil, err
	}

	cinephileNode, err := movie_reflexion.NewCinephileAgent(ctx, cinephileModel)
	if err != nil {
		return nil, err
	}

	clerkNode, err := movie_reflexion.NewClerkAgent(ctx, clerkModel, clerkTools)
	if err != nil {
		return nil, err
	}

	curatorNode, err := movie_reflexion.NewCuratorChain(ctx, curatorModel)
	if err != nil {
		return nil, err
	}

	runnable, err := movie_reflexion.NewFindMoviesPipeline(ctx, cinephileNode, clerkNode, curatorNode)
	if err != nil {
		return nil, err
	}

	u := &FindMoviesUseCase{refiner: refinerNode, runnable: runnable}
	for _, o := range opts {
		o.Apply(u)
	}
	return u, nil
}

func (u *FindMoviesUseCase) Run(ctx context.Context, input string) (string, error) {
	ctx = u.PrepareRun(ctx, u.Name())

	refined, err := u.refiner.Invoke(ctx, input)
	if err != nil {
		return "", err
	}

	final, err := u.runnable.Invoke(ctx, &appmodel.FindMoviesState{
		UserQuery:  *refined,
		MaxRetries: 3,
	})
	if err != nil {
		return "", err
	}

	out, _ := sonic.MarshalString(final.CurrentDraft)
	return out, nil
}

func (u *FindMoviesUseCase) Name() string { return "find-movies" }

func (u *FindMoviesUseCase) ExampleInput() string {
	return "I want obscure horror films from the 90s with a surreal vibe"
}

func (u *FindMoviesUseCase) Description() string {
	return "Find Movies — refines query, suggests underground films, fact-checks in a loop until satisfied."
}
