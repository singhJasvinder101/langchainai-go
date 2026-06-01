package claude

import (
	"context"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/singhJasvinder101/agentic-go/init/config"
	"github.com/singhJasvinder101/agentic-go/llm"
	"github.com/singhJasvinder101/agentic-go/pkg/log"
)

type ClaudeProvider struct {
	Client anthropic.Client
}

func New() (*ClaudeProvider, error) {
	return &ClaudeProvider{
		Client: anthropic.NewClient(
			option.WithAPIKey(config.GetString("claude.api_key")),
		),
	}, nil
}

func (p *ClaudeProvider) Generate(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
	messages, err := llm.PrepareRequest(req)
	if err != nil {
		log.WithContext(ctx).Error("invalid claude generate request", "error", err)
		return nil, err
	}

	params, err := messageParams(messages)
	if err != nil {
		return nil, err
	}

	message, err := p.Client.Messages.New(ctx, params)
	if err != nil {
		return nil, err
	}
	return generateResponseFromMessage(message), nil
}

func (p *ClaudeProvider) GenerateStream(ctx context.Context, req *llm.GenerateRequest) (<-chan *llm.StreamResponse, <-chan error) {
	responses := make(chan *llm.StreamResponse, 100)
	errs := make(chan error, 1)

	messages, err := llm.PrepareRequest(req)
	if err != nil {
		errs <- err
		return responses, errs
	}

	go func() {
		defer closeChannels(responses, errs)

		params, buildErr := messageParams(messages)
		if buildErr != nil {
			errs <- buildErr
			return
		}

		stream := p.Client.Messages.NewStreaming(ctx, params)
		for stream.Next() {
			event := stream.Current()
			switch v := event.AsAny().(type) {
			case anthropic.ContentBlockDeltaEvent:
				switch delta := v.Delta.AsAny().(type) {
				case anthropic.TextDelta:
					if delta.Text == "" {
						continue
					}
					chunk := streamResponseFromTextDelta(delta.Text, &event)
					if chunk != nil {
						responses <- chunk
					}
				}
			}
		}
		if err := stream.Err(); err != nil {
			errs <- err
		}
	}()

	return responses, errs
}

func (p *ClaudeProvider) Close() error {
	// TODO: Implement
	return nil
}

func messageParams(messages []llm.Message) (anthropic.MessageNewParams, error) {
	maxTokens := config.GetInt("claude.max_tokens")
	if maxTokens <= 0 {
		maxTokens = 1024
	}

	converted, err := toClaudeMessages(messages)
	if err != nil {
		return anthropic.MessageNewParams{}, err
	}

	params := anthropic.MessageNewParams{
		Model:     anthropic.Model(config.GetString("claude.model")),
		MaxTokens: int64(maxTokens),
		Messages:  converted.messages,
	}
	if len(converted.system) > 0 {
		params.System = converted.system
	}
	return params, nil
}

func closeChannels(responses chan *llm.StreamResponse, errs chan error) {
	defer close(responses)
	defer close(errs)
}
