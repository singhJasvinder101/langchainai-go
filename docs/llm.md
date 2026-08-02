# LLM Providers

The `llm` package defines a provider-agnostic request/response model — `Message`, `ContentPart`, `Tool`, `GenerateRequest`/`GenerateResponse` — shared by four concrete providers, one per subpackage:

| Package | Provider | Backing SDK |
|---|---|---|
| `llm/openai` | OpenAI | `sashabaranov/go-openai` |
| `llm/claude` | Anthropic Claude | `anthropics/anthropic-sdk-go` |
| `llm/gemini` | Google Gemini | `google.golang.org/genai` |
| `llm/ollama` | Ollama (local models) | `ollama/ollama` |

There is no central factory: each provider has its own `New()` (or `New(ctx)`) constructor, and only the provider package(s) you import get initialized. Configuration for each is read from [`configs/config.yaml`](configuration.md) under that provider's key.

## Table of contents

- [Basic generation](#basic-generation)
- [The Provider interface](#the-provider-interface)
- [Role-based messages](#role-based-messages)
- [Multimodal content](#multimodal-content)
- [Tool calling](#tool-calling)
- [Structured output](#structured-output)
- [Message validation](#message-validation)
- [Reading responses](#reading-responses)
- [Streaming](#streaming)

## Basic generation

```go
package main

import (
	"context"
	"fmt"
	"log"

	initializers "github.com/singhJasvinder101/agentic-go/init"
	"github.com/singhJasvinder101/agentic-go/llm"
	"github.com/singhJasvinder101/agentic-go/llm/gemini"
)

func main() {
	ctx := context.Background()
	initializers.Init(ctx, "configs/config.yaml")

	provider, err := gemini.New(ctx)
	if err != nil {
		log.Fatal(err)
	}

	resp, err := provider.Generate(ctx, &llm.GenerateRequest{
		Messages: []llm.Message{
			llm.UserMessage(llm.TextPart("Hello, what is the capital of France?")),
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(resp.Text())
}
```

Swapping providers is a one-line import + constructor change — `openai.New()`, `claude.New()`, `ollama.New(ctx)` all return a type satisfying the same `Provider` interface and accept the same `*llm.GenerateRequest`.

## The Provider interface

```go
type Provider interface {
	Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
	GenerateStream(ctx context.Context, req *GenerateRequest) (<-chan *StreamResponse, <-chan error)
	Close() error
}
```

All four provider structs (`ClaudeProvider`, `OpenAIProvider`, `GeminiProvider`, `OllamaProvider`) implement this. Write orchestration code — your own functions, or the [agent](agents.md)/[chain](chains.md)/[retriever](retrievers.md) packages — against `llm.Provider` rather than a concrete type, so it works with whichever provider the caller constructs:

```go
func summarize(ctx context.Context, provider llm.Provider, text string) (string, error) {
	resp, err := provider.Generate(ctx, &llm.GenerateRequest{
		Messages: []llm.Message{llm.UserMessage(llm.TextPart("Summarize: " + text))},
	})
	if err != nil {
		return "", err
	}
	return resp.Text(), nil
}
```

## Role-based messages

Every request is a `Messages []Message` slice. Each `Message` has a `Role` (`system`, `user`, `assistant`, `tool`) and one or more `ContentPart` blocks:

```go
resp, err := provider.Generate(ctx, &llm.GenerateRequest{
	Messages: []llm.Message{
		llm.SystemMessage(llm.TextPart("You are a concise assistant.")),
		llm.UserMessage(llm.TextPart("What is the capital of France?")),
	},
})
```

System messages map to each provider's native field where one exists (Claude's `System`, Gemini's `SystemInstruction`); otherwise they're sent as a regular message.

## Multimodal content

A message can carry more than one `ContentPart`:

| Constructor | Purpose |
|---|---|
| `TextPart(text)` | Plain text |
| `ImageURLPart(url)` | Image by URL or data URI |
| `ImagePart(mimeType, data)` | Raw image bytes |
| `FilePart(mimeType, name, data)` | Document/file bytes (provider support varies) |

```go
llm.UserMessage(
	llm.TextPart("Describe this image."),
	llm.ImageURLPart("https://example.com/photo.jpg"),
)
```

Ollama accepts text and raw image bytes on messages. Provider adapters return an error at conversion time for content a given provider/model doesn't support, rather than silently dropping it.

## Tool calling

Register tools on the request and read tool calls off the response:

```go
import "encoding/json"

tools := []llm.Tool{
	llm.NewTool("get_weather", "Get weather for a city",
		json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}`)),
}

req := &llm.GenerateRequest{
	Messages: []llm.Message{
		llm.UserMessage(llm.TextPart("What's the weather in Paris?")),
	},
	Tools: tools,
}
resp, err := provider.Generate(ctx, req)
if err != nil {
	log.Fatal(err)
}

for _, call := range resp.ToolCalls() {
	// run your function, then continue the conversation:
	result := `{"temp": 18, "unit": "C"}`
	history := append(req.Messages,
		llm.AssistantToolCallsMessage("", call),
		llm.ToolMessage(call.ID, call.Name, result),
	)
	resp, err = provider.Generate(ctx, &llm.GenerateRequest{
		Messages: history,
		Tools:    tools,
	})
}
```

**Helpers:** `ToolCallPart`, `ToolResultPart`, `AssistantToolCallsMessage`, `ToolMessage`, `resp.ToolCalls()`, `choice.ToolCalls()`, `call.ParseArguments(&dest)`.

**Tool choice:** set `req.ToolChoice` to `&llm.ToolChoice{Mode: llm.ToolChoiceAuto}` (default), `ToolChoiceNone`, or `ToolChoiceRequired` (optionally with `Name` to force one specific tool). Supported on OpenAI, Claude, Gemini, and Ollama (where the underlying model supports tools).

**History:** assistant turns that made tool calls use `AssistantToolCallsMessage(text, calls...)`; tool results use `ToolMessage(id, name, result)` or a `RoleTool` message with `ToolResultPart`.

Hand-rolling this loop yourself (as above) works for simple one-shot tool use. For anything with more than one round of tool calls, use [`agent.Agent`](agents.md) instead — it runs exactly this loop until the model returns a final answer, with a max-iteration guard and graceful handler-error recovery built in.

## Structured output

`llm.GenerateStructured[T]` builds entirely on the tool-calling machinery above: it reflects a JSON schema from your Go type `T`, forces the model to call a synthetic tool with that schema, and decodes the arguments back into a `T` value — no manual schema authoring and no `json.Unmarshal` in your code.

```go
type WeatherResult struct {
	City      string `json:"city"`
	TempC     int    `json:"temp_c"`
	Condition string `json:"condition"`
}

result, err := llm.GenerateStructured[WeatherResult](ctx, provider, &llm.GenerateRequest{
	Messages: []llm.Message{llm.UserMessage(llm.TextPart("What's the weather in Paris?"))},
})
if err != nil {
	log.Fatal(err)
}
fmt.Println(result.City, result.TempC, result.Condition)
```

Works with any `llm.Provider` that supports tool calling — no provider-specific code required. Struct tags recognized by the underlying schema reflector (e.g. `jsonschema_description`) let you annotate fields for the model without changing your `json` tags.

## Message validation

`Generate` and `GenerateStream` call `llm.PrepareRequest` automatically. That runs light checks (non-empty messages, valid roles, at least one part per message) — it does **not** validate every part's fields (e.g. image MIME type) on every call, to keep the hot path cheap.

Provider adapters (`toOpenAIMessages`, `toGeminiMessages`, etc.) still return an error for unsupported multimodal content at conversion time.

**Optional strict validation** — call explicitly when accepting untrusted input (forms, external JSON, etc.):

```go
req := &llm.GenerateRequest{Messages: history}
if err := req.Validate(); err != nil {
	return err
}
resp, err := provider.Generate(ctx, req)
```

`Message.Validate()`, `ContentPart.Validate()`, and `GenerateRequest.Validate()` check roles, required fields, and part schemas (e.g. that an image part has both MIME type and bytes). Messages built with `UserMessage`, `SystemMessage`, and `resp.AssistantMessage()` typically don't need this on every turn since they're already well-formed.

`llm.ConvertToMessages(messages)` normalizes a raw `[]Message` slice (the same light checks `PrepareRequest` runs) if you want to validate before constructing a full request.

## Reading responses

`Generate` returns `*llm.GenerateResponse` with `Choices`, `Usage`, and an optional `Raw` (the provider-native response, for anything not covered by the unified shape):

```go
resp, _ := provider.Generate(ctx, req)

fmt.Println(resp.Text())              // first choice's text
choice := resp.FirstChoice()
parts := choice.Parts()               // multimodal parts
history = append(history, resp.AssistantMessage())

if native, ok := llm.RawAs[openai.ChatCompletionResponse](resp.Raw); ok {
	_ = native
}
```

`GenerateStream` yields `*llm.StreamResponse` chunks, each with `Choices[].Delta` — a partial assistant message per event.

## Streaming

`Provider.GenerateStream` returns a `(<-chan *StreamResponse, <-chan error)` pair. Wrap it in `llm.NewStream` for a pull-based iterator — in the style of `bufio.Scanner` or `sql.Rows` — instead of hand-writing the two-channel `select` loop:

```go
provider, _ := openai.New()

stream := llm.NewStream(provider.GenerateStream(ctx, &llm.GenerateRequest{
	Messages: []llm.Message{
		llm.UserMessage(llm.TextPart("Write a short poem about Go.")),
	},
}))

for stream.Next() {
	fmt.Print(stream.Text())
}
if err := stream.Err(); err != nil {
	log.Fatal(err)
}
```

`Stream.Current()` returns the full `*StreamResponse` chunk (`Choices[].Delta`, `Raw`, ...) if you need more than the concatenated text `Text()` gives you. `Next()` returns `false` both when the stream completes normally and when it errors — check `Err()` after the loop to tell the two apart, exactly like `bufio.Scanner.Err()`.

The raw `(responses, errs)` channel pair from `GenerateStream` is still there if you need to `select` on it alongside other channels (a cancellation signal, a ticker, etc.) — `Stream` is a convenience wrapper, not a replacement.

## See also

- [Configuration](configuration.md) — how each provider's `New()` reads its API key/model
- [Agents](agents.md) — automates the tool-calling loop shown above
- [Structured output](llm.md#structured-output) is used directly by no other package, but pairs well with [chains](chains.md) when you want a typed final answer
