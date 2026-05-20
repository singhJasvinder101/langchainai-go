package gemini

import "google.golang.org/genai"

type GenerateResponse struct {
	*genai.GenerateContentResponse
}

func (r *GenerateResponse) GetText() string {
	return r.Text()
}

func (r *GenerateResponse) GetResponse() any {
	return r
}
