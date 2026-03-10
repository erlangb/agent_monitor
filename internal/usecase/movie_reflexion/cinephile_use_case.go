package movie_reflexion

import (
	"context"
	"erlangb/agentmonitor/internal/agent/movie_reflexion"
	"erlangb/agentmonitor/internal/factory"
	appmodel "erlangb/agentmonitor/internal/model"
	"erlangb/agentmonitor/internal/usecase"

	"github.com/bytedance/sonic"
)

// CinephileUseCase invokes the cinephile agent directly, without the clerk/curator loop.
type CinephileUseCase struct {
	usecase.BaseEinoUseCase
	node *movie_reflexion.CinephileAgent
}

// NewCinephileUseCase creates the cinephile agent and returns a CinephileUseCase.
func NewCinephileUseCase(ctx context.Context, modelFactory *factory.EinoChatModelFactory, opts ...usecase.Option) (usecase.UseCase, error) {
	chatModel, err := modelFactory.CreateOpenAI(ctx, "gpt-4.1-mini", 0.7, 2048)
	if err != nil {
		return nil, err
	}

	node, err := movie_reflexion.NewCinephileAgent(ctx, chatModel)
	if err != nil {
		return nil, err
	}

	u := &CinephileUseCase{node: node}
	for _, o := range opts {
		o.Apply(u)
	}
	return u, nil
}

func (u *CinephileUseCase) Run(ctx context.Context, input string) (string, error) {
	ctx = u.PrepareRun(ctx, u.Name())

	state, err := u.node.Invoke(ctx, &appmodel.FindMoviesState{
		UserQuery: appmodel.RefinedMovieQuery{OriginalText: input},
	})
	if err != nil {
		return "", err
	}

	out, _ := sonic.MarshalString(state.CurrentDraft)
	return out, nil
}

func (u *CinephileUseCase) Name() string { return "cinephile" }

func (u *CinephileUseCase) ExampleInput() string {
	return "Suggest obscure sci-fi films from 70s Eastern Europe"
}

func (u *CinephileUseCase) Description() string {
	return "Cinephile — underground cinema oracle. Suggests obscure films with no external search."
}
