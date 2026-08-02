# Retrievers / RAG

The `retriever` package bridges [vector stores](vectorstores.md) and [LLM providers](llm.md) into a retrieval-augmented-generation pipeline, without introducing any new storage or embedding code of its own.

## Retriever

```go
type Retriever interface {
	Retrieve(ctx context.Context, query string, k int) ([]vectorstore.Document, error)
}
```

`retriever.FromVectorStore` adapts any `vectorstore.VectorStore` into a `Retriever` by calling its `SimilaritySearch`:

```go
import "github.com/singhJasvinder101/agentic-go/retriever"

r := retriever.FromVectorStore(store) // store is any vectorstore.VectorStore
docs, err := r.Retrieve(ctx, "What is the capital of France?", 3)
```

See [Vector stores](vectorstores.md) for how to construct `store` (in-memory, Chroma, Qdrant, Pinecone, or Weaviate).

## RAGChain

`RAGChain` retrieves the top-K documents for a question and answers it using a provider, joining retrieved content into the system prompt as context:

```go
rag := retriever.NewRAGChain(r, provider, 4) // top 4 documents per question

answer, err := rag.Run(ctx, map[string]any{"question": "What is the capital of France?"})
```

`Run` expects `vars["question"]` to be a non-empty string; everything else about the request (system prompt, message construction) is handled for you.

### Custom system prompt

```go
rag := retriever.NewRAGChain(r, provider, 4)
rag.SystemPrompt = "Answer as a helpful travel guide, citing the context where relevant."
```

If unset, a sensible default is used: *"Answer the question using only the provided context. If the context doesn't contain the answer, say you don't know."*

### RAGChain is a Chain

`RAGChain.Run(ctx, vars) (string, error)` matches [`chain.Chain`](chains.md) exactly, so you can drop it into a `chain.Sequential` pipeline without any adapter:

```go
pipeline := chain.Sequential([]chain.Chain{rag, translate})
```

## Full example

```go
import (
	geminiembedder "github.com/singhJasvinder101/agentic-go/embedder/gemini"
	"github.com/singhJasvinder101/agentic-go/llm/gemini"
	"github.com/singhJasvinder101/agentic-go/retriever"
	"github.com/singhJasvinder101/agentic-go/vectorstore"
	"github.com/singhJasvinder101/agentic-go/vectorstore/memory"
)

emb, _ := geminiembedder.New(ctx)
store, _ := memory.New(emb)
store.AddDocuments(ctx, []vectorstore.Document{
	{ID: "france", Content: "Paris is the capital of France."},
	{ID: "germany", Content: "Berlin is the capital of Germany."},
})

provider, _ := gemini.New(ctx)
rag := retriever.NewRAGChain(retriever.FromVectorStore(store), provider, 2)

answer, err := rag.Run(ctx, map[string]any{"question": "What is the capital of France?"})
fmt.Println(answer)
```

## See also

- [Vector stores](vectorstores.md) — backends `FromVectorStore` can wrap
- [Chains](chains.md) — `RAGChain` composes into `chain.Sequential`
- [Embedders](embedders.md) — required by every vector store
