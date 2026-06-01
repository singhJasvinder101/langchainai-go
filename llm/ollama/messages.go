package ollama

import (
	"fmt"
	"strings"

	"github.com/ollama/ollama/api"
	"github.com/singhJasvinder101/agentic-go/llm"
)

func toAPIMessages(messages []llm.Message) ([]api.Message, error) {
	out := make([]api.Message, 0, len(messages))
	for i, msg := range messages {
		role, err := toOllamaRole(msg.Role)
		if err != nil {
			return nil, fmt.Errorf("message at index %d: %w", i, err)
		}
		apiMsg, err := toAPIMessage(role, msg.Parts)
		if err != nil {
			return nil, fmt.Errorf("message at index %d: %w", i, err)
		}
		out = append(out, apiMsg)
	}
	return out, nil
}

func toAPIMessage(role string, parts []llm.ContentPart) (api.Message, error) {
	var textParts []string
	var images []api.ImageData

	for i, part := range parts {
		switch part.Type {
		case llm.PartText:
			textParts = append(textParts, part.Text)
		case llm.PartImage:
			images = append(images, api.ImageData(part.Data))
		case llm.PartImageURL:
			return api.Message{}, fmt.Errorf("part at index %d: image_url is not supported by ollama; use image bytes", i)
		case llm.PartFile:
			return api.Message{}, fmt.Errorf("part at index %d: file parts are not supported by ollama", i)
		default:
			return api.Message{}, fmt.Errorf("part at index %d: unsupported part type %q", i, part.Type)
		}
	}

	return api.Message{
		Role:    role,
		Content: strings.Join(textParts, "\n"),
		Images:  images,
	}, nil
}

func toOllamaRole(role llm.Role) (string, error) {
	switch role {
	case llm.RoleSystem, llm.RoleUser, llm.RoleAssistant:
		return string(role), nil
	default:
		return "", fmt.Errorf("unsupported role %q", role)
	}
}
