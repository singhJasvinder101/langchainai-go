package openai

import "github.com/sashabaranov/go-openai"

type GenerateResponse struct {
	*openai.ChatCompletionResponse
}

func (r *GenerateResponse) GetText() string {
	if r == nil || r.ChatCompletionResponse == nil || len(r.Choices) == 0 {
		return ""
	}
	return r.Choices[0].Message.Content
}

func (r *GenerateResponse) GetResponse() any {
	return r
}

type StreamResponse struct {
	Response *openai.ChatCompletionStreamResponse
	Text     string
}

func (r *StreamResponse) GetText() string {
	if r == nil {
		return ""
	}
	return r.Text
}

func (r *StreamResponse) GetResponse() any {
	return r
}
