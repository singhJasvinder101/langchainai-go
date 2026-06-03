package gemini

import (
	"context"

	"github.com/singhJasvinder101/agentic-go/init/config"
	"github.com/singhJasvinder101/agentic-go/llm"
	"github.com/singhJasvinder101/agentic-go/pkg/log"
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

func (p *GeminiProvider) Generate(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
	messages, err := llm.PrepareRequest(req)
	if err != nil {
		log.WithContext(ctx).Error("invalid gemini generate request", "error", err)
		return nil, err
	}

	converted, err := toGeminiMessages(messages, req.Tools, req.ToolChoice)
	if err != nil {
		return nil, err
	}

	model := config.GetString("gemini.model")
	response, err := p.Client.Models.GenerateContent(ctx, model, converted.contents, converted.config())
	if err != nil {
		log.WithContext(ctx).Error("gemini generate content failed", "error", err, "model", model)
		return nil, err
	}
	return generateResponseFromGenerateContent(response, model), nil
}

func (p *GeminiProvider) GenerateStream(ctx context.Context, req *llm.GenerateRequest) (<-chan *llm.StreamResponse, <-chan error) {
	responses := make(chan *llm.StreamResponse, 100)
	errs := make(chan error, 1)

	messages, prepErr := llm.PrepareRequest(req)
	if prepErr != nil {
		log.WithContext(ctx).Error("invalid gemini stream request", "error", prepErr)
		errs <- prepErr
		return responses, errs
	}

	go func() {
		defer closeChannels(responses, errs)

		converted, convertErr := toGeminiMessages(messages, req.Tools, req.ToolChoice)
		if convertErr != nil {
			errs <- convertErr
			return
		}

		model := config.GetString("gemini.model")
		for response, err := range p.Client.Models.GenerateContentStream(ctx, model, converted.contents, converted.config()) {
			if err != nil {
				log.WithContext(ctx).Error("gemini generate content stream failed", "error", err, "model", model)
				errs <- err
				return
			}
			chunk := streamResponseFromGenerateContent(response, model)
			if chunk == nil || len(chunk.Choices) == 0 {
				continue
			}
			responses <- chunk
		}
	}()

	return responses, errs
}

func (p *GeminiProvider) Close() error {
	// TODO: Implement
	return nil
}

func closeChannels(responses chan *llm.StreamResponse, errs chan error) {
	defer close(responses)
	defer close(errs)
}
