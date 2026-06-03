package openai

import (
	"testing"

	openaisdk "github.com/sashabaranov/go-openai"
	"github.com/singhJasvinder101/agentic-go/llm"
)

func TestToChatCompletionMessages(t *testing.T) {
	apiMessages, err := toChatCompletionMessages([]llm.Message{
		llm.SystemMessage(llm.TextPart("You are helpful.")),
		llm.UserMessage(llm.TextPart("Hi")),
		llm.AssistantMessage(llm.TextPart("Hello!")),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apiMessages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(apiMessages))
	}
	if apiMessages[0].Role != openaisdk.ChatMessageRoleSystem {
		t.Fatalf("unexpected system role: %s", apiMessages[0].Role)
	}
}

func TestToChatCompletionMessageMultimodal(t *testing.T) {
	apiMsg, err := toChatCompletionMessage(llm.UserMessage(
		llm.TextPart("describe this"),
		llm.ImageURLPart("https://example.com/a.png"),
	))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(apiMsg.MultiContent) != 2 {
		t.Fatalf("expected multi content, got %+v", apiMsg)
	}
}
