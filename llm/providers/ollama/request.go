package ollama

import "errors"

type GenerateRequest struct {
	Prompt string
}

func (r *GenerateRequest) Validate() error {
	if r == nil {
		return errors.New("request is required")
	}
	if len(r.Prompt) == 0 {
		return errors.New("prompt is required")
	}
	return nil
}
