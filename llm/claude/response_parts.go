package claude

import (
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/singhJasvinder101/agentic-go/llm"
)

func partsFromMessage(message *anthropic.Message) []llm.ContentPart {
	if message == nil {
		return nil
	}
	return partsFromContentBlocks(message.Content)
}

func partsFromContentBlocks(blocks []anthropic.ContentBlockUnion) []llm.ContentPart {
	out := make([]llm.ContentPart, 0, len(blocks))
	for _, block := range blocks {
		part, ok := partFromContentBlock(block)
		if ok {
			out = append(out, part)
		}
	}
	return out
}

func partFromContentBlock(block anthropic.ContentBlockUnion) (llm.ContentPart, bool) {
	switch v := block.AsAny().(type) {
	case anthropic.TextBlock:
		if v.Text == "" {
			return llm.ContentPart{}, false
		}
		return llm.TextPart(v.Text), true
	case anthropic.ThinkingBlock:
		if v.Thinking == "" {
			return llm.ContentPart{}, false
		}
		return llm.TextPart(v.Thinking), true
	default:
		return llm.ContentPart{}, false
	}
}
