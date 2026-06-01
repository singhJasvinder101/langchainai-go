package llm

import "testing"

func TestGenerateResponseText(t *testing.T) {
	resp := &GenerateResponse{
		Choices: []Choice{NewChoice(0, []ContentPart{TextPart("hello"), TextPart("world")}, FinishReasonStop)},
	}
	if resp.Text() != "hello\nworld" {
		t.Fatalf("unexpected text: %q", resp.Text())
	}
}

func TestGenerateResponseAssistantMessage(t *testing.T) {
	resp := &GenerateResponse{
		Choices: []Choice{NewChoice(0, []ContentPart{TextPart("hi")}, FinishReasonStop)},
	}
	msg := resp.AssistantMessage()
	if msg.Role != RoleAssistant || len(msg.Parts) != 1 {
		t.Fatalf("unexpected message: %+v", msg)
	}
}

func TestStreamResponseText(t *testing.T) {
	chunk := &StreamResponse{
		Choices: []StreamChoice{NewStreamChoice(0, []ContentPart{TextPart("delta")}, "")},
	}
	if chunk.Text() != "delta" {
		t.Fatalf("unexpected text: %q", chunk.Text())
	}
}

func TestRawAs(t *testing.T) {
	resp := &GenerateResponse{Raw: "native"}
	v, ok := RawAs[string](resp.Raw)
	if !ok || v != "native" {
		t.Fatalf("unexpected raw: %v %v", v, ok)
	}
}
