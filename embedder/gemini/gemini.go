package gemini

import (
	"context"
	"fmt"

	"github.com/singhJasvinder101/agentic-go/embedder"
	"github.com/singhJasvinder101/agentic-go/init/config"
	"github.com/singhJasvinder101/agentic-go/internal/constants"
	"github.com/singhJasvinder101/agentic-go/pkg/log"
	"google.golang.org/genai"
)

type Embedder struct {
	client *genai.Client
}

func New(ctx context.Context) (embedder.Embedder, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: config.GetString(constants.ConfigGeminiAPIKey),
	})
	if err != nil {
		log.WithContext(ctx).Error("failed to create gemini embedder client", "error", err)
		return nil, err
	}
	return &Embedder{client: client}, nil
}

func (e *Embedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("gemini: texts are required")
	}

	contents := make([]*genai.Content, 0, len(texts))
	for _, text := range texts {
		if text == "" {
			return nil, fmt.Errorf("gemini: text at index %d is required", len(contents))
		}
		contents = append(contents, genai.NewContentFromText(text, genai.RoleUser))
	}

	response, err := e.client.Models.EmbedContent(ctx, embeddingModel(), contents, &genai.EmbedContentConfig{
		TaskType: "RETRIEVAL_DOCUMENT",
	})
	if err != nil {
		log.WithContext(ctx).Error("gemini embed documents failed", "error", err, "model", embeddingModel())
		return nil, err
	}

	return parseEmbeddings(response)
}

func (e *Embedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("gemini: text is required")
	}

	response, err := e.client.Models.EmbedContent(ctx, embeddingModel(), genai.Text(text), &genai.EmbedContentConfig{
		TaskType: "RETRIEVAL_QUERY",
	})
	if err != nil {
		log.WithContext(ctx).Error("gemini embed query failed", "error", err, "model", embeddingModel())
		return nil, err
	}

	embeddings, err := parseEmbeddings(response)
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("gemini: embedding response was empty")
	}
	return embeddings[0], nil
}

func embeddingModel() string {
	model := config.GetString(constants.ConfigGeminiEmbeddingModel)
	if model == "" {
		return constants.DefaultGeminiEmbeddingModel
	}
	return model
}

func parseEmbeddings(response *genai.EmbedContentResponse) ([][]float32, error) {
	if response == nil {
		return nil, fmt.Errorf("gemini: embedding response was nil")
	}

	embeddings := make([][]float32, len(response.Embeddings))
	for i, embedding := range response.Embeddings {
		if embedding == nil {
			return nil, fmt.Errorf("gemini: embedding at index %d was nil", i)
		}
		embeddings[i] = embedding.Values
	}
	return embeddings, nil
}
