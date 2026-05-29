package gemini

import (
	"errors"

	"google.golang.org/genai"
)

type (
	Content = *genai.Content
	Role    = *genai.Role
	Part    = *genai.Part
)

type GenerateRequest struct {
	Prompt string
}

func (r *GenerateRequest) Validate() error {
	if len(r.Prompt) == 0 {
		return errors.New("prompt is required")
	}
	return nil
}