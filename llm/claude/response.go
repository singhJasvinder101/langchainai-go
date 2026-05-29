package claude

import (
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
)

type GenerateResponse struct {
	*anthropic.Message
}

func (r *GenerateResponse) GetText() string {
	if r == nil || r.Message == nil {
		return ""
	}
	return extractText(r.Message.Content)
}

func (r *GenerateResponse) GetResponse() any {
	return r
}

type StreamResponse struct {
	Response *anthropic.MessageStreamEventUnion
	Text     string
}

func (r *StreamResponse) GetText() string {
	if r == nil {
		return ""
	}
	return r.Text
}

func (r *StreamResponse) GetResponse() any {
	return r
}

func extractText(blocks []anthropic.ContentBlockUnion) string {
	var b strings.Builder
	for _, block := range blocks {
		if text, ok := block.AsAny().(anthropic.TextBlock); ok {
			b.WriteString(text.Text)
		}
	}
	return b.String()
}
