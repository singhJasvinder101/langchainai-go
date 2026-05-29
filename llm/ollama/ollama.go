package ollama

import (
	"context"
	"fmt"
	"strings"

	"github.com/ollama/ollama/api"
	"github.com/singhJasvinder101/langchainai-go/init/config"
	"github.com/singhJasvinder101/langchainai-go/internal/constants"
	"github.com/singhJasvinder101/langchainai-go/llm"
	"github.com/singhJasvinder101/langchainai-go/pkg/log"
)

type OllamaProvider struct {
	client *api.Client
}

func New(ctx context.Context) (*OllamaProvider, error) {
	client, err := newAPIClient()
	if err != nil {
		log.WithContext(ctx).Error("failed to create ollama client", "error", err)
		return nil, err
	}
	return &OllamaProvider{client: client}, nil
}

func (p *OllamaProvider) Generate(ctx context.Context, req llm.RequestInterface) (llm.ResponseInterface, error) {
	ollamaReq, ok := req.(*GenerateRequest)
	if !ok || ollamaReq == nil {
		return nil, fmt.Errorf("ollama: request must be a non-nil *ollama.GenerateRequest")
	}
	if err := ollamaReq.Validate(); err != nil {
		log.WithContext(ctx).Error("invalid ollama generate request", "error", err)
		return nil, err
	}

	model := modelName()
	stream := false
	var fullText strings.Builder

	err := p.client.Chat(ctx, chatRequest(model, ollamaReq.Prompt, &stream), func(resp api.ChatResponse) error {
		fullText.WriteString(resp.Message.Content)
		return nil
	})
	if err != nil {
		log.WithContext(ctx).Error("ollama chat failed", "error", err, "model", model)
		return nil, err
	}

	return &GenerateResponse{Text: fullText.String()}, nil
}

func (p *OllamaProvider) GenerateStream(ctx context.Context, req llm.RequestInterface) (<-chan llm.ResponseInterface, <-chan error) {
	responses := make(chan llm.ResponseInterface, 100)
	errs := make(chan error, 1)

	ollamaReq, ok := req.(*GenerateRequest)
	if !ok || ollamaReq == nil {
		errs <- fmt.Errorf("ollama: request must be a non-nil *ollama.GenerateRequest")
		closeChannels(responses, errs)
		return responses, errs
	}
	if err := ollamaReq.Validate(); err != nil {
		log.WithContext(ctx).Error("invalid ollama stream request", "error", err)
		errs <- err
		closeChannels(responses, errs)
		return responses, errs
	}

	go func() {
		defer closeChannels(responses, errs)

		model := modelName()
		stream := true

		err := p.client.Chat(ctx, chatRequest(model, ollamaReq.Prompt, &stream), func(resp api.ChatResponse) error {
			if resp.Message.Content == "" {
				return nil
			}
			responses <- &StreamResponse{
				Text:     resp.Message.Content,
				Response: &resp,
			}
			return nil
		})
		if err != nil {
			log.WithContext(ctx).Error("ollama chat stream failed", "error", err, "model", model)
			errs <- err
		}
	}()

	return responses, errs
}

func (p *OllamaProvider) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("ollama: texts are required")
	}

	embeddings := make([][]float32, 0, len(texts))
	for i, text := range texts {
		if text == "" {
			return nil, fmt.Errorf("ollama: text at index %d is required", i)
		}

		response, err := p.client.Embeddings(ctx, &api.EmbeddingRequest{
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

func (p *OllamaProvider) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("ollama: text is required")
	}

	embeddings, err := p.EmbedDocuments(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("ollama: embedding response was empty")
	}
	return embeddings[0], nil
}

func (p *OllamaProvider) Close() error {
	//TODO: Implement
	return nil
}

func chatRequest(model, prompt string, stream *bool) *api.ChatRequest {
	return &api.ChatRequest{
		Model: model,
		Messages: []api.Message{
			{Role: "user", Content: prompt},
		},
		Stream: stream,
	}
}

func modelName() string {
	model := config.GetString(constants.ConfigOllamaModel)
	if model == "" {
		return constants.DefaultOllamaModel
	}
	return model
}

func embeddingModelName() string {
	model := config.GetString(constants.ConfigOllamaEmbeddingModel)
	if model == "" {
		return modelName()
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

func closeChannels(responses chan llm.ResponseInterface, errs chan error) {
	defer close(responses)
	defer close(errs)
}
