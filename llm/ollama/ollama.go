package ollama

import (
	"context"
	"strings"

	"github.com/ollama/ollama/api"
	"github.com/singhJasvinder101/agentic-go/init/config"
	"github.com/singhJasvinder101/agentic-go/internal/constants"
	ollamaclient "github.com/singhJasvinder101/agentic-go/internal/ollama"
	"github.com/singhJasvinder101/agentic-go/llm"
	"github.com/singhJasvinder101/agentic-go/pkg/log"
)

type OllamaProvider struct {
	Client *api.Client
}

func New(ctx context.Context) (*OllamaProvider, error) {
	client, err := ollamaclient.NewAPIClient()
	if err != nil {
		log.WithContext(ctx).Error("failed to create ollama client", "error", err)
		return nil, err
	}
	return &OllamaProvider{Client: client}, nil
}

func (p *OllamaProvider) Generate(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
	messages, err := llm.PrepareRequest(req)
	if err != nil {
		log.WithContext(ctx).Error("invalid ollama generate request", "error", err)
		return nil, err
	}

	apiMessages, err := toAPIMessages(messages)
	if err != nil {
		return nil, err
	}

	model := modelName()
	stream := false
	var fullText strings.Builder
	var lastResp api.ChatResponse

	err = p.Client.Chat(ctx, chatRequest(model, apiMessages, &stream), func(resp api.ChatResponse) error {
		lastResp = resp
		fullText.WriteString(resp.Message.Content)
		return nil
	})
	if err != nil {
		log.WithContext(ctx).Error("ollama chat failed", "error", err, "model", model)
		return nil, err
	}

	return generateResponseFromChat(fullText.String(), model, lastResp), nil
}

func (p *OllamaProvider) GenerateStream(ctx context.Context, req *llm.GenerateRequest) (<-chan *llm.StreamResponse, <-chan error) {
	responses := make(chan *llm.StreamResponse, 100)
	errs := make(chan error, 1)

	messages, prepErr := llm.PrepareRequest(req)
	if prepErr != nil {
		log.WithContext(ctx).Error("invalid ollama stream request", "error", prepErr)
		errs <- prepErr
		return responses, errs
	}

	go func() {
		defer closeChannels(responses, errs)

		apiMessages, convertErr := toAPIMessages(messages)
		if convertErr != nil {
			errs <- convertErr
			return
		}

		model := modelName()
		stream := true

		err := p.Client.Chat(ctx, chatRequest(model, apiMessages, &stream), func(resp api.ChatResponse) error {
			chunk := streamResponseFromChat(resp, model)
			if chunk == nil {
				return nil
			}
			responses <- chunk
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

func chatRequest(model string, messages []api.Message, stream *bool) *api.ChatRequest {
	return &api.ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   stream,
	}
}

func modelName() string {
	model := config.GetString(constants.ConfigOllamaModel)
	if model == "" {
		return constants.DefaultOllamaModel
	}
	return model
}

func closeChannels(responses chan *llm.StreamResponse, errs chan error) {
	defer close(responses)
	defer close(errs)
}
