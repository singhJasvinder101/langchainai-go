package ollama

import (
	"encoding/json"

	"github.com/ollama/ollama/api"
	"github.com/singhJasvinder101/agentic-go/llm"
)

func partsFromOllamaMessage(msg api.Message) []llm.ContentPart {
	parts := make([]llm.ContentPart, 0, 1+len(msg.ToolCalls)+len(msg.Images))

	if msg.Content != "" {
		parts = append(parts, llm.TextPart(msg.Content))
	}

	for _, call := range msg.ToolCalls {
		args, err := json.Marshal(call.Function.Arguments.ToMap())
		if err != nil {
			continue
		}
		parts = append(parts, llm.ToolCallPart(llm.ToolCall{
			ID:        call.ID,
			Name:      call.Function.Name,
			Arguments: args,
		}))
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
