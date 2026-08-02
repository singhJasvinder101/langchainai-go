# Graphs

`graph` is a minimal, general-purpose cyclic state-machine executor — the LangGraph-equivalent piece of this library. It has no dependency on [`llm`](llm.md), [`agent`](agents.md), or [`chain`](chains.md) — node functions are free to call into those packages themselves, which keeps `graph` reusable for non-LLM workflows too.

## Core concepts

```go
type State map[string]any
type NodeFunc func(ctx context.Context, state State) (State, error)
type Router func(state State) string
```

- **State** is a plain map passed between nodes and mutated (or replaced) at each step.
- **NodeFunc** is a unit of work: read `state`, do something (call an LLM, run an [agent](agents.md), hit an API), return the updated `state`.
- **Router** decides which node runs next, based on the current state — this is how conditional branching works.
- **`graph.End`** is the sentinel destination that stops execution.

## Building a graph

```go
import "github.com/singhJasvinder101/agentic-go/graph"

g := graph.New()
g.AddNode("classify", classifyNode)
g.AddNode("handle_billing", billingNode)
g.AddNode("handle_support", supportNode)

g.AddConditionalEdge("classify", func(s graph.State) string {
	return s["category"].(string) // "billing" or "support"
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

final, err := runnable.Run(ctx, graph.State{"input": "why was I charged twice?"})
if err != nil {
	log.Fatal(err)
}
fmt.Println(final["result"])
```

`Compile` catches configuration mistakes (typo'd node names, missing entry point) before you ever run the graph, instead of failing mid-execution.

## Cycles

Because edges (including conditional ones) can point back to an already-visited node, graphs can loop — e.g. a "retry until valid" or "keep refining" pattern:

```go
g.AddNode("draft", draftNode)
g.AddNode("critique", critiqueNode)

g.AddConditionalEdge("critique", func(s graph.State) string {
	if s["approved"].(bool) {
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

A node function is just a `func(ctx, State) (State, error)` — call anything from inside it:

```go
func summarizeNode(ctx context.Context, s graph.State) (graph.State, error) {
	text, _ := s["text"].(string)
	out, err := summarizeChain.Run(ctx, map[string]any{"Text": text})
	if err != nil {
		return nil, err
	}
	s["summary"] = out
	return s, nil
}

func supportAgentNode(ctx context.Context, s graph.State) (graph.State, error) {
	question, _ := s["input"].(string)
	result, err := supportAgent.Run(ctx, question)
	if err != nil {
		return nil, err
	}
	s["result"] = result.Text
	return s, nil
}
```

This is how you build multi-agent or multi-step LLM workflows: each step is a node, each decision point is a conditional edge, and `graph.State` carries data between them — the same shape as LangGraph's `StateGraph`, minus checkpointing/persistence and parallel branch execution (not yet implemented — see the root [README](../README.md#roadmap)).

## See also

- [Agents](agents.md) and [Chains](chains.md) — the most common building blocks for node functions
- [LLM providers](llm.md) — for nodes that call a model directly
