# agentic-go

`agentic-go` is a Go library for building LLM-powered applications and agents — a lean, Go-native alternative to LangChain/LangGraph. Every capability lives in its own package with a direct constructor: import what you use, nothing initializes behind your back, and there's no central registry or factory to configure.

## Why agentic-go

- **Minimal surface, maximum reuse.** A handful of small interfaces (`llm.Provider`, `embedder.Embedder`, `vectorstore.VectorStore`, `memory.Memory`, `chain.Chain`) compose into everything above them — agents, chains, RAG, and graphs are all built on the same primitives you can use directly.
- **No factory, no global registry.** Each provider (`llm/openai`, `llm/claude`, ...) has its own `New()`. Only the packages you import get initialized, and swapping providers is a one-line change.
- **Idiomatic Go.** Explicit errors, `context.Context` everywhere, functional options, generics where they earn their keep (`llm.GenerateStructured[T]`) — nothing here asks you to think in another language's idioms.

## Features

| Capability | Package | Docs |
|---|---|---|
| Chat generation, streaming, tool calling, structured output | `llm`, `llm/openai`, `llm/claude`, `llm/gemini`, `llm/ollama` | [LLM providers](docs/llm.md) |
| Text embeddings | `embedder/openai`, `embedder/gemini`, `embedder/ollama` | [Embedders](docs/embedders.md) |
| Vector storage & similarity search | `vectorstore`, `vectorstore/memory`, `/chroma`, `/qdrant`, `/pinecone`, `/weaviate` | [Vector stores](docs/vectorstores.md) |
| Prompt templating (Go templates + Jinja2) | `template` | [Templates](docs/templates.md) |
| Conversation history | `memory` | [Memory](docs/memory.md) |
| Tool-calling agent loop | `agent` | [Agents](docs/agents.md) |
| Composable prompt → LLM pipelines | `chain` | [Chains](docs/chains.md) |
| Retrieval-augmented generation | `retriever` | [Retrievers / RAG](docs/retrievers.md) |
| Cyclic state-machine orchestration (LangGraph-equivalent) | `graph` | [Graphs](docs/graphs.md) |
| Centralized YAML configuration | `init`, `init/config` | [Configuration](docs/configuration.md) |

## Installation

```bash
go get github.com/singhJasvinder101/agentic-go
```

Vector store backends pull in their own SDK dependencies only when you import them (for example `vectorstore/chroma` adds `chroma-go`).

## Quick start

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

See [Configuration](docs/configuration.md) for `configs/config.yaml`, and [LLM providers](docs/llm.md) for tool calling, streaming, structured output, and multimodal messages.

## Documentation

Full usage guides live in [`docs/`](docs):

- [Configuration](docs/configuration.md) — loading `config.yaml`, reading values
- [LLM providers](docs/llm.md) — generation, messages, tool calling, structured output, streaming, the `Provider` interface
- [Embedders](docs/embedders.md) — `EmbedDocuments` / `EmbedQuery`
- [Vector stores](docs/vectorstores.md) — memory, Chroma, Qdrant, Pinecone, Weaviate
- [Templates](docs/templates.md) — native (`text/template`) and Jinja2 prompt formatting
- [Memory](docs/memory.md) — conversation history buffers
- [Agents](docs/agents.md) — the tool-execution loop
- [Chains](docs/chains.md) — composable prompt → LLM pipelines
- [Retrievers / RAG](docs/retrievers.md) — retrieval-augmented generation
- [Graphs](docs/graphs.md) — cyclic state-machine orchestration

## Package layout

```
llm/              # Chat providers (generate + stream) + Provider interface + GenerateStructured
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
memory/           # Conversation history (Memory interface + Buffer)
agent/            # Tool-execution loop (generate -> tool call -> tool result -> repeat)
chain/            # Composable prompt -> LLM pipelines (Chain, Prompt, Sequential)
retriever/        # VectorStore -> Retriever adapter + RAGChain
graph/            # Cyclic state-machine executor (nodes, static/conditional edges)
template/         # Prompt templates (native + Jinja)
init/config/      # YAML configuration
```

## Supported providers

| Provider | LLM (`llm/`) | Embedder (`embedder/`) |
|---|---|---|
| Gemini | yes | yes |
| OpenAI | yes | yes |
| Claude | yes | no |
| Ollama | yes | yes |

## Example CLI

A placeholder CLI lives at `cmd/agentic-go` (not yet implemented).

## Roadmap

- [x] Tools / agentic workflows
- [x] Role-based messaging
- [x] Tool calling
- [x] Structured output
- [x] Additional LLM and vector store backends
- [x] RAG
- [x] Conversation memory
- [x] Graph / state-machine orchestration
- [ ] MCP support
- [ ] Timeouts and resource limits
- [ ] Guardrails
- [x] Stream response abstraction (like `stream.Next()`)
- [ ] Parallel branch execution and checkpointing in `graph`

## Why Go?

- Goroutines make concurrent chunking and batched embedding trivial for RAG pipelines.
- Static typing catches the request/response shape mismatches that are easy to miss in dynamically-typed equivalents.
- A single static binary with no runtime dependency, which matters for deploying agents as services.
