package openai

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/sashabaranov/go-openai"
	"github.com/singhJasvinder101/langchainai-go/init/config"
	"github.com/singhJasvinder101/langchainai-go/internal/constants"
	"github.com/singhJasvinder101/langchainai-go/llm"
)

type OpenAIProvider struct {
	Client *openai.Client
}

func New() (*OpenAIProvider, error) {
	return &OpenAIProvider{
		Client: openai.NewClient(config.GetString("openai.api_key")),
	}, nil
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
		defer closeChannels(responses, errs)

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

func (p *OpenAIProvider) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("openai: texts are required")
	}
	for i, text := range texts {
		if text == "" {
			return nil, fmt.Errorf("openai: text at index %d is required", i)
		}
	}

	response, err := p.Client.CreateEmbeddings(ctx, openai.EmbeddingRequestStrings{
		Model: openai.EmbeddingModel(openAIEmbeddingModel()),
		Input: texts,
	})
	if err != nil {
		return nil, err
	}

	embeddings := make([][]float32, len(texts))
	for _, item := range response.Data {
		if item.Index >= 0 && item.Index < len(embeddings) {
			embeddings[item.Index] = item.Embedding
		}
	}
	return embeddings, nil
}

func (p *OpenAIProvider) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("openai: text is required")
	}

	embeddings, err := p.EmbedDocuments(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("openai: embedding response was empty")
	}
	return embeddings[0], nil
}

func (p *OpenAIProvider) Close() error {
	//TODO: Implement
	return nil
}

func openAIEmbeddingModel() string {
	model := config.GetString(constants.ConfigOpenAIEmbeddingModel)
	if model == "" {
		return constants.DefaultOpenAIEmbeddingModel
	}
	return model
}

func closeChannels(responses chan llm.ResponseInterface, errs chan error) {
	defer close(responses)
	defer close(errs)
}
