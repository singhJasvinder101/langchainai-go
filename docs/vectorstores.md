# Vector Stores

All vector store backends implement the shared `vectorstore.VectorStore` interface and take an [`embedder.Embedder`](embedders.md) at construction — the store embeds documents on `AddDocuments` and embeds the query on `SimilaritySearch`, so callers never handle raw vectors directly.

```go
type VectorStore interface {
	AddDocuments(ctx context.Context, docs []Document) error
	SimilaritySearch(ctx context.Context, query string, k int) ([]SearchResult, error)
	Delete(ctx context.Context, ids []string) error
}
```

| Package | Backend | Notes |
|---|---|---|
| `vectorstore/memory` | In-process cosine similarity | No external service |
| `vectorstore/chroma` | [Chroma](https://www.trychroma.com/) | Default `http://localhost:8000` |
| `vectorstore/qdrant` | [Qdrant](https://qdrant.tech/) | Default `localhost:6334` |
| `vectorstore/pinecone` | [Pinecone](https://www.pinecone.io/) | Requires API key + index name |
| `vectorstore/weaviate` | [Weaviate](https://weaviate.io/) | Class must exist with `vectorizer: none` |

Backends pull in their own SDK dependencies only when you import that package (for example `vectorstore/chroma` adds `chroma-go`) — importing `vectorstore/memory` alone pulls in nothing extra.

## In-memory

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

## Chroma

Requires a running server, e.g. `docker run -p 8000:8000 chromadb/chroma`.

Your embedder's vectors are passed explicitly (`WithEmbeddings`); a small adapter bridges `embedder.Embedder` to Chroma's collection `EmbeddingFunction` metadata. If `Options.EmbeddingFunction` is omitted, the embedder passed to `New` is used for both storage operations and collection setup.

> **Important:** a Chroma collection is bound to a fixed vector dimension. Use a unique collection name per embedder model (Gemini ≈ 3072, Ollama `all-minilm` ≈ 384) — reusing a collection created with a different embedder causes dimension mismatch errors.

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

Optional: set `Options.EmbeddingFunction` to a different embedder for Chroma collection metadata only — vectors on add/search still use the embedder passed to `New`.

## Qdrant

```go
import qdrantstore "github.com/singhJasvinder101/agentic-go/vectorstore/qdrant"

store, _ := qdrantstore.New(emb, qdrantstore.Options{
	Collection:     "my-docs",
	VectorSize:     3072,
	CreateIfAbsent: true,
})
```

## Pinecone

```go
import pineconestore "github.com/singhJasvinder101/agentic-go/vectorstore/pinecone"

store, _ := pineconestore.New(emb, pineconestore.Options{
	APIKey:    os.Getenv("PINECONE_API_KEY"),
	IndexName: "my-index",
	Namespace: "default",
})
```

## Weaviate

The target class must already exist in Weaviate with `vectorizer: none` (vectors are supplied by your embedder, not computed by Weaviate). See `vectorstore/weaviate` for `Options`.

## See also

- [Embedders](embedders.md) — the `Embedder` every store takes at construction
- [Retrievers / RAG](retrievers.md) — wraps any `VectorStore` into a `Retriever` and answers questions with retrieved context
