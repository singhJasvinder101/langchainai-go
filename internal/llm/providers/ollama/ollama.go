package ollama

import (
	"context"
	"fmt"
	"strings"

	"github.com/ollama/ollama/api"
	"github.com/singhJasvinder101/ai-go/internal/constants"
	"github.com/singhJasvinder101/ai-go/internal/init/config"
	"github.com/singhJasvinder101/ai-go/internal/llm"
	"github.com/singhJasvinder101/ai-go/internal/pkg/log"
)

type OllamaProvider struct {
	client *api.Client
}

func NewOllamaProvider(ctx context.Context) *OllamaProvider {
	client, err := newAPIClient()
	if err != nil {
		log.WithContext(ctx).Fatal("failed to create ollama client", "error", err)
	}
	return &OllamaProvider{client: client}
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

func closeChannels(responses chan llm.ResponseInterface, errs chan error) {
	defer close(responses)
	defer close(errs)
}
