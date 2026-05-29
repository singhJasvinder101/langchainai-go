package claude

import "errors"

type GenerateRequest struct {
	Prompt string
}

func (r *GenerateRequest) Validate() error {
	if len(r.Prompt) == 0 {
		return errors.New("prompt is required")
	}
	return nil
}
