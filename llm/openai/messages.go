package openai

import (
	"encoding/base64"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
	"github.com/singhJasvinder101/agentic-go/llm"
)

func toChatCompletionMessages(messages []llm.Message) ([]openai.ChatCompletionMessage, error) {
	out := make([]openai.ChatCompletionMessage, 0, len(messages))
	for i, msg := range messages {
		role, err := toOpenAIRole(msg.Role)
		if err != nil {
			return nil, fmt.Errorf("message at index %d: %w", i, err)
		}
		apiMsg, err := toChatCompletionMessage(role, msg.Parts)
		if err != nil {
			return nil, fmt.Errorf("message at index %d: %w", i, err)
		}
		out = append(out, apiMsg)
	}
	return out, nil
}

func toChatCompletionMessage(role string, parts []llm.ContentPart) (openai.ChatCompletionMessage, error) {
	if llm.IsTextOnly(parts) {
		text, err := llm.JoinTextParts(parts)
		if err != nil {
			return openai.ChatCompletionMessage{}, err
		}
		return openai.ChatCompletionMessage{Role: role, Content: text}, nil
	}

	multi := make([]openai.ChatMessagePart, 0, len(parts))
	for i, part := range parts {
		apiPart, err := toChatMessagePart(part)
		if err != nil {
			return openai.ChatCompletionMessage{}, fmt.Errorf("part at index %d: %w", i, err)
		}
		multi = append(multi, apiPart)
	}
	return openai.ChatCompletionMessage{Role: role, MultiContent: multi}, nil
}

func toChatMessagePart(part llm.ContentPart) (openai.ChatMessagePart, error) {
	switch part.Type {
	case llm.PartText:
		return openai.ChatMessagePart{
			Type: openai.ChatMessagePartTypeText,
			Text: part.Text,
		}, nil
	case llm.PartImageURL:
		return openai.ChatMessagePart{
			Type: openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{
				URL: part.URL,
			},
		}, nil
	case llm.PartImage:
		return openai.ChatMessagePart{
			Type: openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{
				URL: dataURI(part.MIMEType, part.Data),
			},
		}, nil
	case llm.PartFile:
		return openai.ChatMessagePart{}, fmt.Errorf("file parts are not supported by openai chat completions")
	default:
		return openai.ChatMessagePart{}, fmt.Errorf("unsupported part type %q", part.Type)
	}
}

func dataURI(mimeType string, data []byte) string {
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data))
}

func toOpenAIRole(role llm.Role) (string, error) {
	switch role {
	case llm.RoleSystem:
		return openai.ChatMessageRoleSystem, nil
	case llm.RoleUser:
		return openai.ChatMessageRoleUser, nil
	case llm.RoleAssistant:
		return openai.ChatMessageRoleAssistant, nil
	default:
		return "", fmt.Errorf("unsupported role %q", role)
	}
}
