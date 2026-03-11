package movie_reflexion

import (
	"context"

	appmodel "erlangb/agentmonitor/internal/model"

	"github.com/cloudwego/eino/compose"
)

// FindMoviesPipeline
//
//	flowchart TD
//	    START --> refiner["refiner (query input)"]
//	    refiner --> cinephile
//	    cinephile --> clerk
//	    clerk -->|not satisfied & retries left| cinephile
//	    clerk -->|satisfied OR max retries| curator
//	    curator --> END
type FindMoviesPipeline = compose.Runnable[*appmodel.FindMoviesState, *appmodel.FindMoviesState]

// NewFindMoviesPipeline wires cinephile, clerk, and curator into the find-movies graph
// and compiles it into a runnable pipeline.
func NewFindMoviesPipeline(
	ctx context.Context,
	cinephile *CinephileAgent,
	clerk *ClerkAgent,
	curator *CuratorChain,
) (FindMoviesPipeline, error) {
	g := compose.NewGraph[*appmodel.FindMoviesState, *appmodel.FindMoviesState]()

	if err := g.AddLambdaNode("cinephile", compose.InvokableLambda(cinephile.Invoke), compose.WithNodeName("cinephile")); err != nil {
		return nil, err
	}
	if err := g.AddLambdaNode("clerk", compose.InvokableLambda(clerk.Invoke), compose.WithNodeName("clerk")); err != nil {
		return nil, err
	}
	if err := g.AddLambdaNode("curator", compose.InvokableLambda(curator.Invoke), compose.WithNodeName("curator")); err != nil {
		return nil, err
	}

	if err := g.AddEdge(compose.START, "cinephile"); err != nil {
		return nil, err
	}
	if err := g.AddEdge("cinephile", "clerk"); err != nil {
		return nil, err
	}
	if err := g.AddEdge("curator", compose.END); err != nil {
		return nil, err
	}

	branch := compose.NewGraphBranch(
		func(ctx context.Context, state *appmodel.FindMoviesState) (string, error) {
			if state.IsSatisfied || state.RetryCount >= state.MaxRetries {
				return "curator", nil
			}
			return "cinephile", nil
		},
		map[string]bool{"curator": true, "cinephile": true},
	)
	if err := g.AddBranch("clerk", branch); err != nil {
		return nil, err
	}

	return g.Compile(ctx, compose.WithGraphName("find_movies_graph"))
}
