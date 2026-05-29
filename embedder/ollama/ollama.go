package ollama

import (
	"context"
	"fmt"

	"github.com/ollama/ollama/api"
	"github.com/singhJasvinder101/langchainai-go/embedder"
	"github.com/singhJasvinder101/langchainai-go/init/config"
	"github.com/singhJasvinder101/langchainai-go/internal/constants"
	"github.com/singhJasvinder101/langchainai-go/pkg/log"
)

type Embedder struct {
	client *api.Client
}

func New(ctx context.Context) (embedder.Embedder, error) {
	client, err := newAPIClient()
	if err != nil {
		log.WithContext(ctx).Error("failed to create ollama embedder client", "error", err)
		return nil, err
	}
	return &Embedder{client: client}, nil
}

func (e *Embedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("ollama: texts are required")
	}

	embeddings := make([][]float32, 0, len(texts))
	for i, text := range texts {
		if text == "" {
			return nil, fmt.Errorf("ollama: text at index %d is required", i)
		}

		response, err := e.client.Embeddings(ctx, &api.EmbeddingRequest{
			Model:  embeddingModelName(),
			Prompt: text,
		})
		if err != nil {
			log.WithContext(ctx).Error("ollama embeddings failed", "error", err, "model", embeddingModelName())
			return nil, err
		}

		embeddings = append(embeddings, float64ToFloat32(response.Embedding))
	}
	return embeddings, nil
}

func (e *Embedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("ollama: text is required")
	}

	embeddings, err := e.EmbedDocuments(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("ollama: embedding response was empty")
	}
	return embeddings[0], nil
}

func embeddingModelName() string {
	model := config.GetString(constants.ConfigOllamaEmbeddingModel)
	if model == "" {
		model = config.GetString(constants.ConfigOllamaModel)
	}
	if model == "" {
		return constants.DefaultOllamaModel
	}
	return model
}

func float64ToFloat32(values []float64) []float32 {
	converted := make([]float32, len(values))
	for i, value := range values {
		converted[i] = float32(value)
	}
	return converted
}
