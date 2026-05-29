package claude

import (
	"context"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/singhJasvinder101/langchainai-go/init/config"
	"github.com/singhJasvinder101/langchainai-go/llm"
)

type ClaudeProvider struct {
	Client anthropic.Client
}

func NewClaudeProvider(_ context.Context) *ClaudeProvider {
	return &ClaudeProvider{
		Client: anthropic.NewClient(
			option.WithAPIKey(config.GetString("claude.api_key")),
		),
	}
}

func (p *ClaudeProvider) Generate(ctx context.Context, req llm.RequestInterface) (llm.ResponseInterface, error) {
	claudeReq, ok := req.(*GenerateRequest)
	if !ok || claudeReq == nil {
		return nil, fmt.Errorf("claude: request must be a non-nil *claude.GenerateRequest")
	}
	if err := claudeReq.Validate(); err != nil {
		return nil, err
	}

	message, err := p.Client.Messages.New(ctx, messageParams(claudeReq.Prompt))
	if err != nil {
		return nil, err
	}
	return &GenerateResponse{Message: message}, nil
}

func (p *ClaudeProvider) GenerateStream(ctx context.Context, req llm.RequestInterface) (<-chan llm.ResponseInterface, <-chan error) {
	responses := make(chan llm.ResponseInterface, 100)
	errs := make(chan error, 1)

	claudeReq, ok := req.(*GenerateRequest)
	if !ok || claudeReq == nil {
		errs <- fmt.Errorf("claude: request must be a non-nil *claude.GenerateRequest")
		closeChannels(responses, errs)
		return responses, errs
	}
	if err := claudeReq.Validate(); err != nil {
		errs <- err
		closeChannels(responses, errs)
		return responses, errs
	}

	go func() {
		defer closeChannels(responses, errs)

		stream := p.Client.Messages.NewStreaming(ctx, messageParams(claudeReq.Prompt))
		for stream.Next() {
			event := stream.Current()
			switch v := event.AsAny().(type) {
			case anthropic.ContentBlockDeltaEvent:
				switch delta := v.Delta.AsAny().(type) {
				case anthropic.TextDelta:
					if delta.Text != "" {
						responses <- &StreamResponse{
							Response: &event,
							Text:     delta.Text,
						}
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
	//TODO: Implement
	return nil
}

func messageParams(prompt string) anthropic.MessageNewParams {
	maxTokens := config.GetInt("claude.max_tokens")
	if maxTokens <= 0 {
		maxTokens = 1024
	}

	return anthropic.MessageNewParams{
		Model:     anthropic.Model(config.GetString("claude.model")),
		MaxTokens: int64(maxTokens),
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(prompt)),
		},
	}
}

func closeChannels(responses chan llm.ResponseInterface, errs chan error) {
	defer close(responses)
	defer close(errs)
}