package llm

import (
	"strings"
	"testing"
)

func TestConvertToMessages(t *testing.T) {
	messages, err := ConvertToMessages([]Message{
		SystemMessage(TextPart("You are helpful.")),
		UserMessage(TextPart("Hi")),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(messages))
	}
}

func TestConvertToMessagesRequiresMessages(t *testing.T) {
	_, err := ConvertToMessages(nil)
	if err == nil || !strings.Contains(err.Error(), "messages are required") {
		t.Fatalf("expected messages error, got %v", err)
	}
}

func TestConvertToMessagesRejectsInvalidRole(t *testing.T) {
	_, err := ConvertToMessages([]Message{{Role: "invalid", Parts: []ContentPart{TextPart("x")}}})
	if err == nil || !strings.Contains(err.Error(), "invalid message role") {
		t.Fatalf("expected invalid role error, got %v", err)
	}
}

func TestPrepareRequest(t *testing.T) {
	messages, err := PrepareRequest(&GenerateRequest{
		Messages: []Message{UserMessage(TextPart("hello"))},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
}

func TestValidateIsStricterThanConvert(t *testing.T) {
	msg := Message{Role: RoleUser, Parts: []ContentPart{{Type: PartImage}}}
	messages, err := ConvertToMessages([]Message{msg})
	if err != nil {
		t.Fatalf("ConvertToMessages should allow light check only, got %v", err)
	}
	if err := msg.Validate(); err == nil {
		t.Fatal("Validate should reject image part without data")
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
}
