# agentic-go

`agentic-go` is a Go library for working with LLM providers, text embedders, and vector stores. LLM generation, embeddings, and retrieval are kept in separate packages so you import and initialize only what you need—no central provider factory.

## Features

* **LLM providers:** Gemini, OpenAI, Claude, and Ollama with direct `New()` constructors per provider
* **Streaming:** `GenerateStream` for token-by-token responses
* **Embedders:** LangChain-style `EmbedDocuments` / `EmbedQuery` via a shared `embedder.Embedder` interface (Gemini, OpenAI, Ollama)
* **Vector stores:** Shared `vectorstore.VectorStore` interface with memory, Chroma, Qdrant, Pinecone, and Weaviate backends
* **Configuration:** Centralized YAML for API keys, models, and embedding models
* **Templates:** Native and Jinja prompt formatting

## Installation

```bash
go get github.com/singhJasvinder101/agentic-go
```

Vector store backends pull in their own SDK dependencies when you import them (for example `vectorstore/chroma` adds `chroma-go`).

## Configuration

By default the library reads `configs/config.yaml`. Load it once at startup:

```go
import initializers "github.com/singhJasvinder101/agentic-go/init"

initializers.Init(ctx, "path/to/config.yaml")
// or
config.MustInit("path/to/config.yaml")
```

### Example `config.yaml`

```yaml
log:
  level: "info"   # debug, info, warn, error
  format: "json"  # json, text

gemini:
  api_key: "YOUR_GEMINI_API_KEY"
  model: gemini-2.5-flash
  embedding_model: gemini-embedding-2

openai:
  api_key: "YOUR_OPENAI_API_KEY"
  model: gpt-4o-mini
  embedding_model: text-embedding-3-small

claude:
  api_key: "YOUR_CLAUDE_API_KEY"
  model: claude-sonnet-4-20250514
  max_tokens: 1024

ollama:
  base_url: http://127.0.0.1:11434
  model: smollm:135m
  embedding_model: all-minilm
```

## Usage

### LLM generation

Each LLM lives under `llm/<provider>`. Construct the provider you need directly:

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

### Role-based messages

All LLM providers accept `llm.GenerateRequest` with a `Messages` slice. Each message has a `Role` (`system`, `user`, `assistant`, `tool`) and one or more `ContentPart` blocks:

```go
import "github.com/singhJasvinder101/agentic-go/llm"

resp, err := provider.Generate(ctx, &llm.GenerateRequest{
	Messages: []llm.Message{
		llm.SystemMessage(llm.TextPart("You are a concise assistant.")),
		llm.UserMessage(llm.TextPart("What is the capital of France?")),
	},
})
```

**Content part types:** `TextPart`, `ImageURLPart`, `ImagePart` (raw bytes + MIME type), `FilePart` (documents; provider support varies). Multimodal example:

```go
llm.UserMessage(
	llm.TextPart("Describe this image."),
	llm.ImageURLPart("https://example.com/photo.jpg"),
)
```

System messages map to provider-native fields where supported (Claude `System`, Gemini `SystemInstruction`). Ollama accepts text and raw image bytes on messages.

### Tool calling

Register tools on `GenerateRequest` and read tool calls from the response:

```go
import "encoding/json"

tools := []llm.Tool{
	llm.NewTool("get_weather", "Get weather for a city",
		json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}`)),
}

resp, err := provider.Generate(ctx, &llm.GenerateRequest{
	Messages: []llm.Message{
		llm.UserMessage(llm.TextPart("What's the weather in Paris?")),
	},
	Tools: tools,
})
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

**Helpers:** `ToolCallPart`, `ToolResultPart`, `AssistantToolCallsMessage`, `ToolMessage`, `resp.ToolCalls()`, `choice.ToolCalls()`.

**Tool choice:** set `ToolChoice` with `ToolChoiceAuto` (default), `ToolChoiceNone`, or `ToolChoiceRequired` (optionally `Name` to force a specific tool). Supported on OpenAI, Claude, Gemini, and Ollama (where the model supports tools).

**History:** assistant turns with tool calls use `AssistantToolCallsMessage`; tool results use `ToolMessage` or `RoleTool` messages with `ToolResultPart`.

### Message validation (LangChain-style)

`Generate` and `GenerateStream` call `llm.PrepareRequest` automatically. That runs light checks (non-empty messages, valid roles, at least one part per message) and does **not** call `Validate()` on every part.

Provider adapters (`toOpenAIMessages`, `toGeminiMessages`, etc.) return errors for unsupported multimodal content when converting to the API.

**Optional strict validation** — call explicitly when accepting untrusted input (forms, external JSON, etc.):

```go
req := &llm.GenerateRequest{Messages: history}
if err := req.Validate(); err != nil {
	return err
}
resp, err := provider.Generate(ctx, req)
```

`Message.Validate()`, `ContentPart.Validate()`, and `GenerateRequest.Validate()` check roles, required fields, and part schemas (e.g. image MIME type and bytes). Messages built with `UserMessage`, `SystemMessage`, and `resp.AssistantMessage()` typically do not need this on every turn.

You can also normalize input with `llm.ConvertToMessages(messages)` before building a request.

### Responses

`Generate` returns `*llm.GenerateResponse` with `Choices`, `Usage`, and optional `Raw` (provider-native). Each `Choice` has a full assistant `Message` with `ContentPart` slices:

```go
resp, _ := provider.Generate(ctx, req)

fmt.Println(resp.Text())                    // first choice text
choice := resp.FirstChoice()
parts := choice.Parts()                     // multimodal parts
history = append(history, resp.AssistantMessage())

if native, ok := llm.RawAs[openai.ChatCompletionResponse](resp.Raw); ok {
	_ = native
}
```

`GenerateStream` yields `*llm.StreamResponse` chunks with `Choices[].Delta` (partial assistant message per event).

### Streaming

```go
provider, _ := openai.New()
responses, errs := provider.GenerateStream(ctx, &llm.GenerateRequest{
	Messages: []llm.Message{
		llm.UserMessage(llm.TextPart("Write a short poem about Go.")),
	},
})

for responses != nil || errs != nil {
	select {
	case chunk, ok := <-responses:
		if !ok {
			responses = nil
		} else {
			fmt.Print(chunk.Text())
		}
	case err, ok := <-errs:
		if !ok {
			errs = nil
		} else if err != nil {
			log.Fatal(err)
		}
	}
}
```

### Embedders

Embeddings are separate from LLM providers. Use `embedder/<provider>` packages:

```go
import (
	geminiembedder "github.com/singhJasvinder101/agentic-go/embedder/gemini"
	openaiembedder "github.com/singhJasvinder101/agentic-go/embedder/openai"
	ollamaembedder "github.com/singhJasvinder101/agentic-go/embedder/ollama"
)

emb, err := geminiembedder.New(ctx)
if err != nil {
	log.Fatal(err)
}

docs, err := emb.EmbedDocuments(ctx, []string{
	"Paris is the capital of France.",
	"Berlin is the capital of Germany.",
})
query, err := emb.EmbedQuery(ctx, "What is the capital of France?")
```

Supported embedders: **Gemini**, **OpenAI**, **Ollama**. Claude is chat-only (no embedder package).

`EmbedDocuments` returns one vector per input text, in the same order as the input slice.

### Vector stores

All stores implement `vectorstore.VectorStore` and take an `embedder.Embedder` at construction. The store embeds documents on `AddDocuments` and embeds the query on `SimilaritySearch`.

| Package | Backend | Notes |
|---------|---------|--------|
| `vectorstore/memory` | In-process cosine similarity | No external service |
| `vectorstore/chroma` | [Chroma](https://www.trychroma.com/) | Default `http://localhost:8000` |
| `vectorstore/qdrant` | [Qdrant](https://qdrant.tech/) | Default `localhost:6334` |
| `vectorstore/pinecone` | [Pinecone](https://www.pinecone.io/) | Requires API key + index name |
| `vectorstore/weaviate` | [Weaviate](https://weaviate.io/) | Class must exist with `vectorizer: none` |

#### In-memory example

```go
import (
	geminiembedder "github.com/singhJasvinder101/agentic-go/embedder/gemini"
	"github.com/singhJasvinder101/agentic-go/vectorstore"
	"github.com/singhJasvinder101/agentic-go/vectorstore/memory"
)

emb, _ := geminiembedder.New(ctx)
store, _ := memory.New(emb)

_ = store.AddDocuments(ctx, []vectorstore.Document{
	{ID: "france", Content: "Paris is the capital of France."},
	{ID: "germany", Content: "Berlin is the capital of Germany."},
})

results, _ := store.SimilaritySearch(ctx, "capital of France?", 2)
for _, r := range results {
	fmt.Println(r.Document.Content, r.Score)
}
```

#### Chroma example

Chroma requires a running server (for example `docker run -p 8000:8000 chromadb/chroma`).

Your embedder vectors are passed explicitly (`WithEmbeddings`); a small adapter bridges `embedder.Embedder` to Chroma's collection `EmbeddingFunction` metadata. If `Options.EmbeddingFunction` is omitted, the embedder passed to `New` is used for both storage operations and collection setup.

**Important:** A Chroma collection is bound to a fixed vector dimension. Use a unique collection name per embedder model (Gemini ≈ 3072, Ollama `all-minilm` ≈ 384). Reusing a collection created with a different embedder causes dimension mismatch errors.

```go
import (
	geminiembedder "github.com/singhJasvinder101/agentic-go/embedder/gemini"
	"github.com/singhJasvinder101/agentic-go/vectorstore"
	chromastore "github.com/singhJasvinder101/agentic-go/vectorstore/chroma"
)

emb, _ := geminiembedder.New(ctx)

store, err := chromastore.New(ctx, emb, chromastore.Options{
	BaseURL:    "http://localhost:8000",
	Collection: "my-docs-gemini", // unique per embedder / dimension
})
if err != nil {
	log.Fatal(err)
}

_ = store.AddDocuments(ctx, []vectorstore.Document{
	{Content: "Paris is the capital of France."},
})

results, _ := store.SimilaritySearch(ctx, "What is the capital of France?", 1)
```

Optional: set `Options.EmbeddingFunction` to a different embedder for Chroma collection metadata only (vectors on add/search still use the embedder passed to `New`).

#### Qdrant example

```go
import qdrantstore "github.com/singhJasvinder101/agentic-go/vectorstore/qdrant"

store, _ := qdrantstore.New(emb, qdrantstore.Options{
	Collection:     "my-docs",
	VectorSize:     3072,
	CreateIfAbsent: true,
})
```

#### Pinecone example

```go
import pineconestore "github.com/singhJasvinder101/agentic-go/vectorstore/pinecone"

store, _ := pineconestore.New(emb, pineconestore.Options{
	APIKey:    os.Getenv("PINECONE_API_KEY"),
	IndexName: "my-index",
	Namespace: "default",
})
```

## Package layout

```
llm/              # Chat providers (generate + stream)
  gemini/
  openai/
  claude/
  ollama/
embedder/         # Embedding providers (separate from llm/)
  gemini/
  openai/
  ollama/
vectorstore/      # Retrieval backends
  memory/
  chroma/
  qdrant/
  pinecone/
  weaviate/
template/         # Prompt templates (native + Jinja)
init/config/      # YAML configuration
```

## Supported providers

| Provider | LLM (`llm/`) | Embedder (`embedder/`) |
|----------|--------------|-------------------------|
| Gemini | yes | yes |
| OpenAI | yes | yes |
| Claude | yes | no |
| Ollama | yes | yes |

## Example CLI

A placeholder CLI lives at `cmd/agentic-go` (not yet implemented).

## TODO

- [ ] Tools / agentic workflows
- [ ] MCP support
- [x] Additional LLM and vector store backends
- [x] Role-based messaging
- [x] Tool calling
- [ ] Timeouts and resource limits
- [ ] RAG
- [ ] Guardrails
- [ ] Stream response abstraction (like stream.Next())


## Why Go ?
- you can use goroutines for concurrent chunking for rag applications
- python, javascript languages are dyanmic type languages so go's status typing can beat