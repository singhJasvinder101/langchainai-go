package chroma

import (
	"context"

	"github.com/singhJasvinder101/langchainai-go/embedder/gemini"
)

func DefaultEmbeddingFunction() (ChromaEmbedderAdapterFunction, error) {
	geminiEmbedder, err := gemini.New(context.Background())
	if err != nil {
		return ChromaEmbedderAdapterFunction{}, err
	}
	return ChromaEmbedderAdapterFunction{Embedder: geminiEmbedder}, nil
}
