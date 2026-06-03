package openai

import (
	"context"
	"errors"
	"io"

	"github.com/sashabaranov/go-openai"
	"github.com/singhJasvinder101/agentic-go/init/config"
	"github.com/singhJasvinder101/agentic-go/llm"
)

type OpenAIProvider struct {
	Client *openai.Client
}

func New() (*OpenAIProvider, error) {
	return &OpenAIProvider{
		Client: openai.NewClient(config.GetString("openai.api_key")),
	}, nil
}

func (p *OpenAIProvider) Generate(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
	messages, err := llm.PrepareRequest(req)
	if err != nil {
		return nil, err
	}

	apiMessages, err := toChatCompletionMessages(messages)
	if err != nil {
		return nil, err
	}

	apiReq := openai.ChatCompletionRequest{
		Model:    config.GetString("openai.model"),
		Messages: apiMessages,
	}
	if len(req.Tools) > 0 {
		tools, err := toOpenAITools(req.Tools)
		if err != nil {
			return nil, err
		}
		apiReq.Tools = tools
		apiReq.ToolChoice = toOpenAIToolChoice(req.ToolChoice)
	}

	response, err := p.Client.CreateChatCompletion(ctx, apiReq)
	if err != nil {
		return nil, err
	}
	return generateResponseFromChatCompletion(&response), nil
}

func (p *OpenAIProvider) GenerateStream(ctx context.Context, req *llm.GenerateRequest) (<-chan *llm.StreamResponse, <-chan error) {
	responses := make(chan *llm.StreamResponse, 100)
	errs := make(chan error, 1)

	messages, err := llm.PrepareRequest(req)
	if err != nil {
		errs <- err
		return responses, errs
	}

	go func() {
		defer closeChannels(responses, errs)

		apiMessages, convertErr := toChatCompletionMessages(messages)
		if convertErr != nil {
			errs <- convertErr
			return
		}

		streamReq := openai.ChatCompletionRequest{
			Model:    config.GetString("openai.model"),
			Messages: apiMessages,
			Stream:   true,
		}
		if len(req.Tools) > 0 {
			tools, toolErr := toOpenAITools(req.Tools)
			if toolErr != nil {
				errs <- toolErr
				return
			}
			streamReq.Tools = tools
			streamReq.ToolChoice = toOpenAIToolChoice(req.ToolChoice)
		}

		stream, err := p.Client.CreateChatCompletionStream(ctx, streamReq)
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

			chunk := streamResponseFromChunk(&response)
			if chunk == nil || len(chunk.Choices) == 0 {
				continue
			}
			responses <- chunk
		}
	}()

	return responses, errs
}

func (p *OpenAIProvider) Close() error {
	//TODO: Implement
	return nil
}

func closeChannels(responses chan *llm.StreamResponse, errs chan error) {
	defer close(responses)
	defer close(errs)
}
