package llm

import (
	"strings"
	"testing"
)

func TestChatRequestValidateMessages(t *testing.T) {
	req := &GenerateRequest{
		Messages: []Message{
			SystemMessage(TextPart("You are helpful.")),
			UserMessage(TextPart("Hi")),
		},
	}
	err := req.Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(req.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(req.Messages))
	}
}

func TestChatRequestRequiresMessages(t *testing.T) {
	err := (&GenerateRequest{}).Validate()
	if err == nil || !strings.Contains(err.Error(), "messages are required") {
		t.Fatalf("expected messages error, got %v", err)
	}
}

func TestValidateMessagesRejectsInvalidRole(t *testing.T) {
	err := (&GenerateRequest{Messages: []Message{{Role: "invalid", Parts: []ContentPart{TextPart("x")}}}}).Validate()
	if err == nil || !strings.Contains(err.Error(), "invalid message role") {
		t.Fatalf("expected invalid role error, got %v", err)
	}
}

func TestContentPartValidation(t *testing.T) {
	if err := (ContentPart{Type: PartImage}).Validate(); err == nil {
		t.Fatal("expected image validation error")
	}
}
