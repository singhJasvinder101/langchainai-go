package vectorstore

import (
	"context"

	"github.com/singhJasvinder101/agentic-go/embedder"
)

// document is a type of content stored and retrieved by a vector store
type Document struct {
	ID       string
	Content  string
	Metadata map[string]string
}

// SearchResult pairs a document with its similarity score (higher is more similar).
type SearchResult struct {
	Document Document
	Score    float32
}

// VectorStore persists document embeddings and supports similarity search.
// Implementations accept an embedder.Embedder at construction time so the same
// store can work with any embedding provider.
type VectorStore interface {
	AddDocuments(ctx context.Context, docs []Document) error
	SimilaritySearch(ctx context.Context, query string, k int) ([]SearchResult, error)
	Delete(ctx context.Context, ids []string) error
}

// Embedder is re-exported for convenience when wiring stores.
type Embedder = embedder.Embedder
