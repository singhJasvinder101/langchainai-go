# Graphs

`graph` is a minimal, general-purpose cyclic state-machine executor — the LangGraph-equivalent piece of this library. It has no dependency on [`llm`](llm.md), [`agent`](agents.md), or [`chain`](chains.md) — node functions are free to call into those packages themselves, which keeps `graph` reusable for non-LLM workflows too.

State is generic — you define your own state struct, and the compiler checks every field access. There is no `map[string]any` and no `s["field"].(T)` type assertion anywhere; a typo'd key or wrong type is a build error, not a runtime panic.

## Core concepts

```go
type NodeFunc[S any] func(ctx context.Context, state S) (S, error)
type Router[S any] func(state S) string
```

- **S** is your own state type — a plain struct, chosen when you call `graph.New[S]()`. Fields are read/written as `s.Field`.
- **NodeFunc[S]** is a unit of work: read `state`, do something (call an LLM, run an [agent](agents.md), hit an API), return the updated `state`.
- **Router[S]** decides which node runs next, based on the current state — this is how conditional branching works.
- **`graph.End`** is the sentinel destination that stops execution.

## Building a graph

```go
import "github.com/singhJasvinder101/agentic-go/graph"

type SupportState struct {
	Input    string
	Category string
	Result   string
}

g := graph.New[SupportState]()
g.AddNode("classify", classifyNode)
g.AddNode("handle_billing", billingNode)
g.AddNode("handle_support", supportNode)

g.AddConditionalEdge("classify", func(s SupportState) string {
	return s.Category // "billing" or "support" — a compile-time-checked field, not a map lookup
}, map[string]string{
	"billing": "handle_billing",
	"support": "handle_support",
})
g.AddEdge("handle_billing", graph.End)
g.AddEdge("handle_support", graph.End)
g.SetEntryPoint("classify")
```

- **`AddNode(name, fn)`** registers a node. Registering the same name twice returns `graph.ErrNodeExists`.
- **`AddEdge(from, to)`** is a static, unconditional edge. `to` may be `graph.End`.
- **`AddConditionalEdge(from, router, routes)`** routes to whichever node the `routes` map has for `router`'s return value.
- **`SetEntryPoint(name)`** sets where execution starts.
- A node with no outgoing edge implicitly ends the run there — you don't have to `AddEdge(node, graph.End)` for terminal nodes, though doing so is more explicit and self-documenting.

## Compiling and running

```go
runnable, err := g.Compile() // validates: entry point set + exists, every edge target exists or is End
if err != nil {
	log.Fatal(err)
}

final, err := runnable.Run(ctx, SupportState{Input: "why was I charged twice?"})
if err != nil {
	log.Fatal(err)
}
fmt.Println(final.Result)
```

`Compile` catches configuration mistakes (typo'd node names, missing entry point) before you ever run the graph, instead of failing mid-execution. `Run` (and every `NodeFunc`) is typed to `SupportState` specifically — passing the wrong state type, or misspelling `final.Result`, fails at `go build`, not at runtime.

## Cycles

Because edges (including conditional ones) can point back to an already-visited node, graphs can loop — e.g. a "retry until valid" or "keep refining" pattern:

```go
type DraftState struct {
	Draft    string
	Approved bool
}

g := graph.New[DraftState]()
g.AddNode("draft", draftNode)
g.AddNode("critique", critiqueNode)

g.AddConditionalEdge("critique", func(s DraftState) string {
	if s.Approved {
		return "done"
	}
	return "retry"
}, map[string]string{
	"done":  graph.End,
	"retry": "draft",
})
g.AddEdge("draft", "critique")
g.SetEntryPoint("draft")
```

`Run` guards against an unterminated cycle with a default 100-step limit, returning an error if exceeded. Use `RunWithLimit(ctx, state, maxSteps)` to set your own:

```go
final, err := runnable.RunWithLimit(ctx, initialState, 20)
```

## Composing with agents and chains

A node function is just a `func(ctx, S) (S, error)` — call anything from inside it:

```go
type PipelineState struct {
	Text    string
	Summary string
	Result  string
}

func summarizeNode(ctx context.Context, s PipelineState) (PipelineState, error) {
	out, err := summarizeChain.Run(ctx, map[string]any{"Text": s.Text})
	if err != nil {
		return s, err
	}
	s.Summary = out
	return s, nil
}

func supportAgentNode(ctx context.Context, s PipelineState) (PipelineState, error) {
	result, err := supportAgent.Run(ctx, s.Text)
	if err != nil {
		return s, err
	}
	s.Result = result.Text
	return s, nil
}
```

This is how you build multi-agent or multi-step LLM workflows: each step is a node, each decision point is a conditional edge, and your state struct carries data between them, with the compiler enforcing every node agrees on its shape — the same idea as LangGraph's `StateGraph`, minus checkpointing/persistence and parallel branch execution (not yet implemented — see the root [README](../README.md#roadmap)).

## See also

- [Agents](agents.md) and [Chains](chains.md) — the most common building blocks for node functions
- [LLM providers](llm.md) — for nodes that call a model directly
