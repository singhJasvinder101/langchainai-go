package gemini

import (
	"fmt"
	"strings"

	"github.com/singhJasvinder101/agentic-go/llm"
	"google.golang.org/genai"
)

type geminiMessages struct {
	contents          []*genai.Content
	systemInstruction *genai.Content
}

func toGeminiMessages(messages []llm.Message) (geminiMessages, error) {
	var systemParts []string
	contents := make([]*genai.Content, 0, len(messages))

	for i, msg := range messages {
		switch msg.Role {
		case llm.RoleSystem:
			text, err := llm.JoinTextParts(msg.Parts)
			if err != nil {
				return geminiMessages{}, fmt.Errorf("message at index %d: %w", i, err)
			}
			systemParts = append(systemParts, text)
		case llm.RoleUser, llm.RoleAssistant:
			content, err := toGeminiContent(msg)
			if err != nil {
				return geminiMessages{}, fmt.Errorf("message at index %d: %w", i, err)
			}
			contents = append(contents, content)
		default:
			return geminiMessages{}, fmt.Errorf("message at index %d: unsupported role %q", i, msg.Role)
		}
	}

	result := geminiMessages{contents: contents}
	if len(systemParts) > 0 {
		result.systemInstruction = genai.NewContentFromText(strings.Join(systemParts, "\n\n"), genai.RoleUser)
	}
	if len(contents) == 0 && result.systemInstruction == nil {
		return geminiMessages{}, fmt.Errorf("at least one user or assistant message is required")
	}
	return result, nil
}

func toGeminiContent(msg llm.Message) (*genai.Content, error) {
	parts := make([]*genai.Part, 0, len(msg.Parts))
	for i, part := range msg.Parts {
		apiPart, err := toGeminiPart(part)
		if err != nil {
			return nil, fmt.Errorf("part at index %d: %w", i, err)
		}
		parts = append(parts, apiPart)
	}

	role := genai.RoleUser
	if msg.Role == llm.RoleAssistant {
		role = genai.RoleModel
	}
	return &genai.Content{Role: role, Parts: parts}, nil
}

func toGeminiPart(part llm.ContentPart) (*genai.Part, error) {
	switch part.Type {
	case llm.PartText:
		return genai.NewPartFromText(part.Text), nil
	case llm.PartImageURL:
		mimeType := part.MIMEType
		if mimeType == "" {
			mimeType = "image/jpeg"
		}
		return genai.NewPartFromURI(part.URL, mimeType), nil
	case llm.PartImage, llm.PartFile:
		return genai.NewPartFromBytes(part.Data, part.MIMEType), nil
	default:
		return nil, fmt.Errorf("unsupported part type %q", part.Type)
	}
}

func (g geminiMessages) config() *genai.GenerateContentConfig {
	if g.systemInstruction == nil {
		return nil
	}
	return &genai.GenerateContentConfig{
		SystemInstruction: g.systemInstruction,
	}
}
