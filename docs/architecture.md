# AgentMonitor — How It's Built

AgentMonitor is a Go CLI for running AI agent pipelines. You pick a use case from a menu, type a prompt, and get back a result. Under the hood, multiple LLM calls and tool invocations happen in a structured pipeline before anything reaches you.

---

## The big picture

The code is split into four concerns, top to bottom:

**UI** — two interchangeable frontends: a bubbletea TUI and a plain terminal. Both implement the same `Runner` interface. They know nothing about agents or LLMs.

**Use cases** — each use case implements `UseCase.Run(ctx, input) → string`. It owns the full pipeline for one task: creates models, wires agents, runs the graph, serializes the output. The use case is the seam between the UI and the AI framework.

**Agent nodes** — the actual LLM logic. Each node is a Go struct wrapping a compiled Eino chain or adk.Agent. Nodes take a typed state, call a model, mutate the state, and return it.

**Infrastructure** — factories (model creation, MCP client creation), config (koanf, env vars expanded once at load), and the MCP client wrapper for external tool servers.

---

## The main use case: FindMovies

The most complete pipeline is `FindMoviesUseCase`. It takes a free-text movie request and returns a curated, fact-checked list of underground films. Four agent nodes run in sequence:

```
User input (string)
    │
    ▼
RefinerChain          — extracts structured search params from free text
                        e.g. genre, year range, is_classic flag
    │
    ▼
 ┌──────────────────────────────────────┐
 │           reflexion loop             │
 │                                      │
 │  CinephileAgent  — suggests films    │
 │       │                              │
 │       ▼                              │
 │  ClerkAgent      — fact-checks them  │
 │       │           via Tavily search  │
 │       │                              │
 │       ├── not satisfied → back to Cinephile
 │       └── satisfied or max retries reached
 │                                      │
 └──────────────────────────────────────┘
    │
    ▼
CuratorChain          — prunes bad movies, finalises the list
    │
    ▼
JSON output
```

The refiner runs once before the graph. The cinephile–clerk loop runs inside an Eino `compose.Graph` with a branch condition. The curator is the terminal node.

---

## How agent nodes are structured

Every node follows the same pattern:

```go
type CinephileAgent struct {
    runnable compose.Runnable[*FindMoviesState, *CinephileResponse]
}

func NewCinephileAgent(ctx context.Context, m ToolCallingChatModel) (*CinephileAgent, error) {
    // build prompt template + compile chain → store as runnable
}

func (a *CinephileAgent) Invoke(ctx context.Context, state *FindMoviesState) (*FindMoviesState, error) {
    // call runnable, mutate state, return
}
```

`Invoke` is domain logic — testable directly with a fake model, no framework setup needed. The compiled runnable is just an implementation detail inside the struct.

Nodes that need tool calls (like the Clerk) use `adk.Agent` instead of a chain, which handles the multi-turn model→tool→model loop internally.

---

## State threading

All nodes in the FindMovies graph share a single `FindMoviesState` struct passed by pointer:

```go
type FindMoviesState struct {
    UserQuery       RefinedMovieQuery  // set by refiner, read by all nodes
    CurrentDraft    []Movie            // written by cinephile, pruned by curator
    CritiqueHistory []string           // appended by clerk on each iteration
    IsSatisfied     bool               // clerk sets this to exit the loop
    RetryCount      int
    MaxRetries      int
}
```

The refiner produces a `RefinedMovieQuery` with structured fields (`PrimaryGenre`, `StartYear`, `QueryInfo`, …). This is extracted once from the raw user string and never re-parsed. Every downstream node reads from it directly.

---

## Testing approach

Nodes are unit-tested by passing a hand-written fake model. Rather than using a mock framework (mockery, gomock) to stub the LLM interface, we opted for a simple fake struct — the interface is small enough that a mock would only add noise:

```go
type fakeChatModel struct{ response string }

func (f *fakeChatModel) Generate(_ context.Context, _ []*schema.Message, _ ...model.Option) (*schema.Message, error) {
    return schema.AssistantMessage(f.response, nil), nil
}
```

This satisfies `ToolCallingChatModel` (3 methods). No API keys, no MCP server, no network — just `node.Invoke(ctx, state)` with controlled model output. The clerk's `adk.Agent` works the same way: the fake model returns a clean JSON response with no tool calls, so the tool loop exits on the first iteration.

---

## What Eino provides

[Eino](https://github.com/cloudwego/eino) (CloudWeGo) is the AI orchestration framework. It provides:

- `compose.Chain` — linear pipeline of typed steps, compiled to a `Runnable`
- `compose.Graph` — stateful graph with edges and branch conditions
- `adk.Agent` — tool-calling loop (model → tools → model) with configurable max iterations
- `prompt.FromMessages` — typed prompt templates
- Callback hooks — used here for token/cost tracking via agentmeter

The rest of the stack is standard Go: `koanf` for config, `bubbletea` for the TUI, `sonic` for JSON, `testify` for tests.