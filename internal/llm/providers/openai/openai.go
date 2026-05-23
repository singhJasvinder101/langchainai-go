package openai

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
	"github.com/singhJasvinder101/ai-go/internal/init/config"
	"github.com/singhJasvinder101/ai-go/internal/llm"
)

type OpenAIProvider struct {
	Client *openai.Client
}

func NewOpenAIProvider(_ context.Context) *OpenAIProvider {
	return &OpenAIProvider{
		Client: openai.NewClient(config.GetString("openai.api_key")),
	}
}

func (p *OpenAIProvider) Generate(ctx context.Context, req llm.RequestInterface) (llm.ResponseInterface, error) {
	openaiReq, ok := req.(*GenerateRequest)
	if !ok || openaiReq == nil {
		return nil, fmt.Errorf("openai: request must be a non-nil *openai.GenerateRequest")
	}
	if err := openaiReq.Validate(); err != nil {
		return nil, err
	}

	response, err := p.Client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: config.GetString("openai.model"),
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: openaiReq.Prompt,
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return &GenerateResponse{ChatCompletionResponse: &response}, nil
}

func (p *OpenAIProvider) GenerateStream(ctx context.Context, req llm.RequestInterface) (<-chan llm.ResponseInterface, <-chan error) {
	responses := make(chan llm.ResponseInterface, 100)
	errs := make(chan error, 1)

	openaiReq, ok := req.(*GenerateRequest)
	if !ok || openaiReq == nil {
		errs <- fmt.Errorf("openai: request must be a non-nil *openai.GenerateRequest")
		closeChannels(responses, errs)
		return responses, errs
	}
	if err := openaiReq.Validate(); err != nil {
		errs <- err
		closeChannels(responses, errs)
		return responses, errs
	}

	go func() {
		closeChannels(responses, errs)

		stream, err := p.Client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
			Model: config.GetString("openai.model"),
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: openaiReq.Prompt,
				},
			},
			Stream: true,
		})
		if err != nil {
			errs <- err
			return
		}
		defer stream.Close()

		for {
			response, recvErr := stream.Recv()
			if errors.Is(recvErr, io.EOF) {
				return
			}
			if recvErr != nil {
				errs <- recvErr
				return
			}

			text := ""
			if len(response.Choices) > 0 {
				text = response.Choices[0].Delta.Content
			}
			responses <- &StreamResponse{
				Response: &response,
				Text:     text,
			}
		}
	}()

	return responses, errs
}

func (p *OpenAIProvider) Close() error {
	return nil
}

func closeChannels(responses chan llm.ResponseInterface, errs chan error) {
	defer close(responses)
	defer close(errs)
}
