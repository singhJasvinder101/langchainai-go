package openai

import (
	"fmt"

	openai "github.com/sashabaranov/go-openai"
	"github.com/singhJasvinder101/agentic-go/llm"
)

func partsFromChatMessage(msg openai.ChatCompletionMessage) []llm.ContentPart {
	if len(msg.MultiContent) > 0 {
		parts := make([]llm.ContentPart, 0, len(msg.MultiContent))
		for _, block := range msg.MultiContent {
			part, err := partFromChatMessagePart(block)
			if err == nil {
				parts = append(parts, part)
			}
		}
		return parts
	}
	if msg.Content != "" {
		return []llm.ContentPart{llm.TextPart(msg.Content)}
	}
	return nil
}

func partFromChatMessagePart(block openai.ChatMessagePart) (llm.ContentPart, error) {
	switch block.Type {
	case openai.ChatMessagePartTypeText:
		return llm.TextPart(block.Text), nil
	case openai.ChatMessagePartTypeImageURL:
		if block.ImageURL == nil {
			return llm.ContentPart{}, fmt.Errorf("image_url block missing url")
		}
		return llm.ImageURLPart(block.ImageURL.URL), nil
	default:
		return llm.ContentPart{}, fmt.Errorf("unsupported block type %q", block.Type)
	}
}

func partsFromStreamDelta(delta openai.ChatCompletionStreamChoiceDelta) []llm.ContentPart {
	if delta.Content == "" {
		return nil
	}
	return []llm.ContentPart{llm.TextPart(delta.Content)}
}
