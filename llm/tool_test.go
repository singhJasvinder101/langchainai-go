package llm

import (
	"encoding/json"
	"testing"
)

func TestToolValidate(t *testing.T) {
	if err := (Tool{Name: ""}).Validate(); err == nil {
		t.Fatal("expected error for empty name")
	}
	tool := NewTool("get_weather", "Get weather", json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}}}`))
	if err := tool.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAssistantToolCallsMessage(t *testing.T) {
	msg := AssistantToolCallsMessage("thinking", ToolCall{
		ID:        "call_1",
		Name:      "get_weather",
		Arguments: json.RawMessage(`{"city":"Paris"}`),
	})
	if msg.Role != RoleAssistant {
		t.Fatalf("expected assistant role, got %s", msg.Role)
	}
	calls := ToolCalls(msg.Parts)
	if len(calls) != 1 || calls[0].Name != "get_weather" {
		t.Fatalf("unexpected tool calls: %+v", calls)
	}
}

func TestToolMessage(t *testing.T) {
	msg := ToolMessage("call_1", "get_weather", `{"temp":72}`)
	if msg.Role != RoleTool {
		t.Fatalf("expected tool role, got %s", msg.Role)
	}
	if err := msg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGenerateResponseToolCalls(t *testing.T) {
	resp := &GenerateResponse{
		Choices: []Choice{{
			Message: Message{
				Role: RoleAssistant,
				Parts: []ContentPart{
					ToolCallPart(ToolCall{Name: "fn", Arguments: json.RawMessage(`{}`)}),
				},
			},
		}},
	}
	calls := resp.ToolCalls()
	if len(calls) != 1 || calls[0].Name != "fn" {
		t.Fatalf("unexpected calls: %+v", calls)
	}
}
