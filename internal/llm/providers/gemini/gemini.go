package gemini

import (
	"context"
	"fmt"

	"github.com/singhJasvinder101/ai-go/internal/init/config"
	"github.com/singhJasvinder101/ai-go/internal/llm"
	"github.com/singhJasvinder101/ai-go/internal/pkg/log"
	"google.golang.org/genai"
)

type GeminiProvider struct {
	Client *genai.Client
}

func NewGeminiProvider(ctx context.Context) *GeminiProvider {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: config.GetString("gemini.api_key"),
	})
	if err != nil {
		log.WithContext(ctx).Fatal("failed to create gemini client", "error", err)
	}

	return &GeminiProvider{Client: client}
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

func (p *GeminiProvider) Close() error {
	return nil
}

func closeChannels(responses chan llm.ResponseInterface, errs chan error) {
	defer close(responses)
	defer close(errs)
}
