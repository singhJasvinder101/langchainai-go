package openai

import (
	"testing"

	openaisdk "github.com/sashabaranov/go-openai"
	"github.com/singhJasvinder101/agentic-go/llm"
)

func TestPartsFromChatMessageText(t *testing.T) {
	parts := partsFromChatMessage(openaisdk.ChatCompletionMessage{
		Role:    openaisdk.ChatMessageRoleAssistant,
		Content: "hello",
	})
	if len(parts) != 1 || parts[0].Type != llm.PartText {
		t.Fatalf("unexpected parts: %+v", parts)
	}
}

func TestPartsFromChatMessageMulti(t *testing.T) {
	parts := partsFromChatMessage(openaisdk.ChatCompletionMessage{
		Role: openaisdk.ChatMessageRoleAssistant,
		MultiContent: []openaisdk.ChatMessagePart{
			{Type: openaisdk.ChatMessagePartTypeText, Text: "see"},
			{Type: openaisdk.ChatMessagePartTypeImageURL, ImageURL: &openaisdk.ChatMessageImageURL{URL: "https://example.com/x.png"}},
		},
	})
	if len(parts) != 2 {
		t.Fatalf("expected 2 parts, got %d", len(parts))
	}
}
