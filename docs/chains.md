# Chains

`chain.Chain` is the smallest possible composable unit in this library — intentionally minimal (no LCEL-style typed I/O or streaming):

```go
type Chain interface {
	Run(ctx context.Context, vars map[string]any) (string, error)
}
```

Anything with that method signature is a `Chain` — including [`retriever.RAGChain`](retrievers.md), which satisfies it structurally without importing the `chain` package.

## Prompt

`chain.Prompt` renders a template and sends the result to an [`llm.Provider`](llm.md#the-provider-interface) as a single user turn. There are two ways to build one — reach for whichever matches how many templates you have.

### The common case: one template, no registry

`chain.NewPromptTemplate` takes the template text directly — no [`template.Registry`](templates.md) to create, no key to invent and pass around:

```go
import "github.com/singhJasvinder101/agentic-go/chain"

summarize := chain.NewPromptTemplate(provider, "Summarize the following text: {{.Text}}")
output, err := summarize.Run(ctx, map[string]any{"Text": "..."})
```

This is the whole setup. It defaults to native (Go `text/template`) syntax; set `Engine` for Jinja:

```go
summarize.Engine = template.FormatterJinja // template text would then use {{ text }} syntax
```

### Sharing templates across chains: Registry

If you have several templates that multiple chains reuse, register them once in a shared [`template.Registry`](templates.md) and build each `Prompt` from a key instead:

```go
registry := template.NewRegistry()
registry.MustRegisterTemplate("summary", template.FormatterNative, "Summarize the following text: {{.Text}}")

summarize := chain.NewPrompt(provider, registry, "summary")
output, err := summarize.Run(ctx, map[string]any{"Text": "..."})
```

Both constructors return the same `*chain.Prompt`, with the same `SystemPrompt` field to set a system message alongside the rendered template:

```go
summarize.SystemPrompt = "You are a concise technical writer."
```

## Sequential

`chain.Sequential` runs chains in order, feeding each chain's text output into the next chain's `vars` — the classic "prompt A, then feed its answer into prompt B" pipeline:

```go
translate := chain.NewPromptTemplate(provider, "Translate to French: {{.input}}")

pipeline := chain.Sequential([]chain.Chain{summarize, translate})
output, err := pipeline.Run(ctx, map[string]any{"Text": "..."})
// summarize.Run(vars) -> "a short summary"
// translate.Run(vars + {"input": "a short summary"}) -> "un résumé court"
```

By default the previous chain's output is placed under the `"input"` key. Override it with `chain.WithOutputKey`:

```go
pipeline := chain.Sequential([]chain.Chain{summarize, translate}, chain.WithOutputKey("previous_step"))
```

Each downstream chain still sees all the original `vars` too — `Sequential` only adds the output key, it never removes anything.

## Writing your own Chain

Since `Chain` is a single method, wrapping any operation — a non-LLM computation, an external API call, a [`retriever.RAGChain`](retrievers.md), or another `Sequential` pipeline — as a chain step is just implementing `Run`:

```go
type uppercase struct{}

func (uppercase) Run(ctx context.Context, vars map[string]any) (string, error) {
	text, _ := vars["input"].(string)
	return strings.ToUpper(text), nil
}

pipeline := chain.Sequential([]chain.Chain{summarize, uppercase{}})
```

## See also

- [Templates](templates.md) — the `template.Registry` `Prompt` renders from
- [Retrievers / RAG](retrievers.md) — `RAGChain` is a drop-in `Chain`
- [LLM providers](llm.md) — `llm.Provider`, the interface `Prompt` sends requests through
