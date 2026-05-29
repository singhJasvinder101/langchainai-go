package gemini

import (
	"context"
	"fmt"

	"github.com/singhJasvinder101/langchainai-go/init/config"
	internalConstants "github.com/singhJasvinder101/langchainai-go/internal/constants"
	"github.com/singhJasvinder101/langchainai-go/llm"
	"github.com/singhJasvinder101/langchainai-go/pkg/log"
	"google.golang.org/genai"
)

type GeminiProvider struct {
	Client *genai.Client
}

func New(ctx context.Context) (*GeminiProvider, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: config.GetString("gemini.api_key"),
	})
	if err != nil {
		log.WithContext(ctx).Error("failed to create gemini client", "error", err)
		return nil, err
	}
	return &GeminiProvider{Client: client}, nil
}

func (p *GeminiProvider) Generate(ctx context.Context, req llm.RequestInterface) (llm.ResponseInterface, error) {
	geminiReq, ok := req.(*GenerateRequest)
	if !ok || geminiReq == nil {
		return nil, fmt.Errorf("gemini: request must be a non-nil *gemini.GenerateRequest")
	}
	if err := geminiReq.Validate(); err != nil {
		log.WithContext(ctx).Error("invalid gemini generate request", "error", err)
		return nil, err
	}

	model := config.GetString("gemini.model")
	contents := []Content{{Role: "user", Parts: []Part{{Text: geminiReq.Prompt}}}}
	response, err := p.Client.Models.GenerateContent(ctx, model, contents, nil)
	if err != nil {
		log.WithContext(ctx).Error("gemini generate content failed", "error", err, "model", model)
		return nil, err
	}
	return &GenerateResponse{response}, nil
}

func (p *GeminiProvider) GenerateStream(ctx context.Context, req llm.RequestInterface) (<-chan llm.ResponseInterface, <-chan error) {
	responses := make(chan llm.ResponseInterface, 100)
	errs := make(chan error, 1)

	geminiReq, ok := req.(*GenerateRequest)
	if !ok || geminiReq == nil {
		errs <- fmt.Errorf("gemini: request must be a non-nil *gemini.GenerateRequest")
		closeChannels(responses, errs)
		return responses, errs
	}
	if err := geminiReq.Validate(); err != nil {
		log.WithContext(ctx).Error("invalid gemini stream request", "error", err)
		errs <- err
		closeChannels(responses, errs)
		return responses, errs
	}

	go func() {
		defer closeChannels(responses, errs)

		model := config.GetString("gemini.model")
		contents := []Content{{Role: "user", Parts: []Part{{Text: geminiReq.Prompt}}}}
		for response, err := range p.Client.Models.GenerateContentStream(ctx, model, contents, nil) {
			if err != nil {
				log.WithContext(ctx).Error("gemini generate content stream failed", "error", err, "model", model)
				errs <- err
				return
			}
			responses <- &GenerateResponse{response}
		}
	}()

	return responses, errs
}

func (p *GeminiProvider) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
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

	response, err := p.Client.Models.EmbedContent(ctx, geminiEmbeddingModel(), contents, &genai.EmbedContentConfig{
		TaskType: "RETRIEVAL_DOCUMENT",
	})
	if err != nil {
		log.WithContext(ctx).Error("gemini embed documents failed", "error", err, "model", geminiEmbeddingModel())
		return nil, err
	}

	return geminiEmbeddings(response)
}

func (p *GeminiProvider) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("gemini: text is required")
	}

	response, err := p.Client.Models.EmbedContent(ctx, geminiEmbeddingModel(), genai.Text(text), &genai.EmbedContentConfig{
		TaskType: "RETRIEVAL_QUERY",
	})
	if err != nil {
		log.WithContext(ctx).Error("gemini embed query failed", "error", err, "model", geminiEmbeddingModel())
		return nil, err
	}

	embeddings, err := geminiEmbeddings(response)
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("gemini: embedding response was empty")
	}
	return embeddings[0], nil
}

func (p *GeminiProvider) Close() error {
	//TODO: Implement
	return nil
}

func geminiEmbeddingModel() string {
	model := config.GetString(internalConstants.ConfigGeminiEmbeddingModel)
	if model == "" {
		return internalConstants.DefaultGeminiEmbeddingModel
	}
	return model
}

func geminiEmbeddings(response *genai.EmbedContentResponse) ([][]float32, error) {
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

func closeChannels(responses chan llm.ResponseInterface, errs chan error) {
	defer close(responses)
	defer close(errs)
}
