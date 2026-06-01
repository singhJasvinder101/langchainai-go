package ollama

import (
	"github.com/ollama/ollama/api"
	"github.com/singhJasvinder101/agentic-go/llm"
)

func partsFromOllamaMessage(fullText string, msg api.Message) []llm.ContentPart {
	parts := make([]llm.ContentPart, 0, 1+len(msg.Images))

	text := fullText
	if text == "" {
		text = msg.Content
	}
	if text != "" {
		parts = append(parts, llm.TextPart(text))
	}

	for _, img := range msg.Images {
		if len(img) == 0 {
			continue
		}
		parts = append(parts, llm.ImagePart("image/png", img))
	}
	return parts
}

func partsFromStreamContent(content string) []llm.ContentPart {
	if content == "" {
		return nil
	}
	return []llm.ContentPart{llm.TextPart(content)}
}
