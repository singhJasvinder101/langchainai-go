# Embedders

Embeddings are deliberately kept separate from [LLM providers](llm.md) — some LLM providers don't have an embedding model (Claude), and you often want a different provider for embeddings than for chat. Each embedder lives under `embedder/<provider>` and implements the shared `embedder.Embedder` interface:

```go
type Embedder interface {
	EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error)
	EmbedQuery(ctx context.Context, text string) ([]float32, error)
}
```

| Package | Provider |
|---|---|
| `embedder/gemini` | Google Gemini |
| `embedder/openai` | OpenAI |
| `embedder/ollama` | Ollama (local models) |

Claude is chat-only and has no `embedder/claude` package.

## Usage

```go
import (
	geminiembedder "github.com/singhJasvinder101/agentic-go/embedder/gemini"
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

`EmbedDocuments` returns one vector per input text, in the same order as the input slice. Each embedder reads its config (API key, embedding model name) from [`configs/config.yaml`](configuration.md) under its provider key (`gemini.embedding_model`, `openai.embedding_model`, `ollama.embedding_model`).

## See also

- [Vector stores](vectorstores.md) — every store takes an `Embedder` at construction and uses it to embed documents on add and queries on search
- [Retrievers / RAG](retrievers.md) — builds on vector stores to answer questions using retrieved context
