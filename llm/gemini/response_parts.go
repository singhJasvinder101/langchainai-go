package gemini

import (
	"encoding/json"
	"fmt"

	"github.com/singhJasvinder101/agentic-go/llm"
	"google.golang.org/genai"
)

func partsFromGenaiParts(parts []*genai.Part) []llm.ContentPart {
	if len(parts) == 0 {
		return nil
	}
	out := make([]llm.ContentPart, 0, len(parts))
	for i, part := range parts {
		if part == nil {
			continue
		}
		converted, err := partFromGenaiPart(part)
		if err != nil {
			continue
		}
		if converted.Type != "" {
			out = append(out, converted)
		} else {
			_ = i
		}
	}
	return out
}

func partFromGenaiPart(part *genai.Part) (llm.ContentPart, error) {
	if part.Text != "" {
		return llm.TextPart(part.Text), nil
	}
	if part.InlineData != nil && len(part.InlineData.Data) > 0 {
		mime := part.InlineData.MIMEType
		if mime == "" {
			mime = "application/octet-stream"
		}
		return llm.ImagePart(mime, part.InlineData.Data), nil
	}
	if part.FileData != nil && part.FileData.FileURI != "" {
		mime := part.FileData.MIMEType
		if mime == "" {
			mime = "application/octet-stream"
		}
		return llm.ContentPart{
			Type:     llm.PartImageURL,
			URL:      part.FileData.FileURI,
			MIMEType: mime,
		}, nil
	}
	if part.FunctionCall != nil {
		args, err := json.Marshal(part.FunctionCall.Args)
		if err != nil {
			return llm.ContentPart{}, err
		}
		return llm.ToolCallPart(llm.ToolCall{
			ID:        part.FunctionCall.ID,
			Name:      part.FunctionCall.Name,
			Arguments: args,
		}), nil
	}
	return llm.ContentPart{}, fmt.Errorf("unsupported genai part")
}
