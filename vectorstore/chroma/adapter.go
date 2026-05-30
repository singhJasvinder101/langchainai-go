package chroma

import (
	"context"

	chromaembeddings "github.com/amikos-tech/chroma-go/pkg/embeddings"
	"github.com/singhJasvinder101/agentic-go/embedder"
)

var _ chromaembeddings.EmbeddingFunction = ChromaEmbedderAdapterFunction{}

type ChromaEmbedderAdapterFunction struct {
	embedder.Embedder
}

func (f ChromaEmbedderAdapterFunction) EmbedDocuments(ctx context.Context, texts []string) ([]chromaembeddings.Embedding, error) {
	vectors, err := f.Embedder.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, err
	}
	out := make([]chromaembeddings.Embedding, len(vectors))
	for i, vector := range vectors {
		out[i] = chromaembeddings.NewEmbeddingFromFloat32(vector)
	}
	return out, nil
}

func (f ChromaEmbedderAdapterFunction) EmbedQuery(ctx context.Context, text string) (chromaembeddings.Embedding, error) {
	vector, err := f.Embedder.EmbedQuery(ctx, text)
	if err != nil {
		return nil, err
	}
	return chromaembeddings.NewEmbeddingFromFloat32(vector), nil
}

func (f ChromaEmbedderAdapterFunction) DefaultSpace() chromaembeddings.DistanceMetric {
	return chromaembeddings.DistanceMetric("cosine")
}

func (f ChromaEmbedderAdapterFunction) SupportedSpaces() []chromaembeddings.DistanceMetric {
	return []chromaembeddings.DistanceMetric{"cosine"}
}

func (f ChromaEmbedderAdapterFunction) Name() string {
	return "chroma-embedder-adapter"
}

func (f ChromaEmbedderAdapterFunction) GetConfig() chromaembeddings.EmbeddingFunctionConfig {
	return chromaembeddings.EmbeddingFunctionConfig{}
}