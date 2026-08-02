# Agents

`agent.Agent` automates the loop you'd otherwise hand-roll around [tool calling](llm.md#tool-calling): call the model, execute any tool calls it requests via your own handlers, feed the results back, and repeat — until the model returns a final text answer or a maximum number of iterations is reached.

## Defining a tool

A tool pairs an `llm.Tool` definition (name, description, JSON schema) with a handler that actually runs it. There are two ways to build one.

### The Pydantic-model way: `NewTypedTool`

Define the tool's arguments as a plain Go struct — the JSON schema is reflected from it automatically (via [`llm.SchemaFor`](llm.md#structured-output)), and the handler receives already-decoded arguments instead of raw JSON:

```go
import (
	"context"

	"github.com/singhJasvinder101/agentic-go/agent"
)

type WeatherArgs struct {
	City string `json:"city" jsonschema_description:"City name, e.g. Paris"`
}

weatherTool := agent.NewTypedTool("get_weather", "Get current weather for a city",
	func(ctx context.Context, args WeatherArgs) (string, error) {
		// call your real weather API here
		return fmt.Sprintf(`{"temp_c":18,"condition":"cloudy","city":%q}`, args.City), nil
	})
```

No hand-written JSON schema string, and no `json.Unmarshal` inside the handler — the struct's `json` tags define field names/optionality (a field without `omitempty` is required), and `jsonschema_description` tags become each field's description for the model. This is the recommended default.

### The manual way: `NewTool`

When the schema needs to be built dynamically (e.g. loaded from configuration or another system) rather than known at compile time as a Go type, build it directly from a JSON schema string and a raw handler:

```go
import "encoding/json"

weatherTool := agent.NewTool(
	"get_weather",
	"Get current weather for a city",
	json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}`),
	func(ctx context.Context, args json.RawMessage) (string, error) {
		var in struct {
			City string `json:"city"`
		}
		if err := json.Unmarshal(args, &in); err != nil {
			return "", err
		}
		return fmt.Sprintf(`{"temp_c":18,"condition":"cloudy","city":%q}`, in.City), nil
	},
)
```

Both return the same `agent.Tool`, so they're interchangeable in `agent.WithTools`. `Handler` (the raw form) is `func(ctx context.Context, args json.RawMessage) (string, error)`; either way, the string returned is sent back to the model verbatim as the tool result.

## Running an agent

```go
import "github.com/singhJasvinder101/agentic-go/agent"

a := agent.New(provider, // any llm.Provider
	agent.WithTools(weatherTool),
	agent.WithSystemPrompt("You are a helpful weather assistant."),
)

result, err := a.Run(ctx, "What's the weather in Paris?")
if err != nil {
	log.Fatal(err)
}
fmt.Println(result.Text)
```

`Run` returns an `*agent.Result`:

```go
type Result struct {
	Text     string        // the model's final answer
	Messages []llm.Message // full transcript: system prompt, memory, user input, every tool call/result
	Steps    int           // how many generate calls it took
}
```

## Options

| Option | Effect |
|---|---|
| `agent.WithTools(tools ...Tool)` | Registers tools the agent may call |
| `agent.WithSystemPrompt(prompt string)` | Sets the system prompt sent with every turn |
| `agent.WithMaxIterations(n int)` | Caps generate/tool-execute round trips (default 10) |
| `agent.WithMemory(m memory.Memory)` | Reads prior history at the start of `Run`; appends the new turn on success |

```go
a := agent.New(provider,
	agent.WithTools(weatherTool, otherTool),
	agent.WithSystemPrompt("You are a helpful assistant."),
	agent.WithMemory(memory.NewBuffer(memory.WithMaxMessages(50))),
	agent.WithMaxIterations(15),
)
```

See [Memory](memory.md) for `memory.Buffer`.

## Error handling in tools

If a handler returns an error, the agent does **not** abort the run — it surfaces `"error: <message>"` as that tool's result, so the model sees the failure and can retry, try a different tool, or explain the problem to the user:

```go
func(ctx context.Context, args json.RawMessage) (string, error) {
	return "", errors.New("weather API rate-limited")
	// model sees: "error: weather API rate-limited"
}
```

An unknown tool name (one the model calls but that wasn't registered) is handled the same way — surfaced as `error: unknown tool "..."` rather than a panic or aborted run.

## Max iterations

If the model never stops requesting tool calls within `MaxIterations` turns, `Run` returns `agent.ErrMaxIterations`:

```go
result, err := a.Run(ctx, input)
if errors.Is(err, agent.ErrMaxIterations) {
	// the model looped without converging — log, alert, or retry with a stricter prompt
}
```

## What Agent does not do (yet)

- No streaming — `Run` is request/response only. Use `llm.Provider.GenerateStream` directly if you need token-by-token output for a single turn.
- No built-in parallel tool execution — tool calls in a single turn are executed sequentially. If your handlers are slow and independent, run them concurrently yourself before feeding results back (or open an issue/PR).

## See also

- [LLM providers](llm.md#tool-calling) — the underlying tool-calling request/response shapes `Agent` is built on
- [Memory](memory.md) — conversation history across `Run` calls
- [Graphs](graphs.md) — for orchestration beyond a single tool loop (e.g. routing between multiple agents/chains)
