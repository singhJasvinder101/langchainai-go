package ollama

import "github.com/ollama/ollama/api"

type GenerateResponse struct {
	Text     string
	Response *api.ChatResponse
}

func (r *GenerateResponse) GetText() string {
	if r == nil {
		return ""
	}
	return r.Text
}

func (r *GenerateResponse) GetResponse() any {
	return r
}

type StreamResponse struct {
	Text     string
	Response *api.ChatResponse
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
