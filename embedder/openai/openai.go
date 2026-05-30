package openai

import (
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/singhJasvinder101/agentic-go/embedder"
	"github.com/singhJasvinder101/agentic-go/init/config"
	"github.com/singhJasvinder101/agentic-go/internal/constants"
)

type Embedder struct {
	client *openai.Client
}

func New() (embedder.Embedder, error) {
	return &Embedder{
		client: openai.NewClient(config.GetString(constants.ConfigOpenAIAPIKey)),
	}, nil
}

func (e *Embedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("openai: texts are required")
	}
	for i, text := range texts {
		if text == "" {
			return nil, fmt.Errorf("openai: text at index %d is required", i)
		}
	}

	response, err := e.client.CreateEmbeddings(ctx, openai.EmbeddingRequestStrings{
		Model: openai.EmbeddingModel(embeddingModel()),
		Input: texts,
	})
	if err != nil {
		return nil, err
	}

	embeddings := make([][]float32, len(texts))
	for _, item := range response.Data {
		if item.Index >= 0 && item.Index < len(embeddings) {
			embeddings[item.Index] = item.Embedding
		}
	}
	return embeddings, nil
}

func (e *Embedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("openai: text is required")
	}

	embeddings, err := e.EmbedDocuments(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("openai: embedding response was empty")
	}
	return embeddings[0], nil
}

func embeddingModel() string {
	model := config.GetString(constants.ConfigOpenAIEmbeddingModel)
	if model == "" {
		return constants.DefaultOpenAIEmbeddingModel
	}
	return model
}
