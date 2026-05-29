package embedder

import "context"

// Embedder converts text into dense vectors for retrieval and similarity search.
// Implementations live in provider-specific subpackages (e.g. embedder/openai).
type Embedder interface {
	EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error)
	EmbedQuery(ctx context.Context, text string) ([]float32, error)
}
