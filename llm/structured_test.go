package llm_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/singhJasvinder101/agentic-go/internal/llmtest"
	"github.com/singhJasvinder101/agentic-go/llm"
)

type weatherResult struct {
	City      string `json:"city"`
	TempC     int    `json:"temp_c"`
	Condition string `json:"condition"`
}

func toolCallResponseWithArgs(t *testing.T, name string, args any) *llm.GenerateResponse {
	t.Helper()
	raw, err := json.Marshal(args)
	if err != nil {
		t.Fatalf("marshal args: %v", err)
	}
	return llmtest.ToolCallResponse(llm.ToolCall{ID: "call_1", Name: name, Arguments: raw})
}

func TestGenerateStructuredDecodesToolCallArguments(t *testing.T) {
	provider := llmtest.NewMockProvider().AddResponse(toolCallResponseWithArgs(t, "emit_result", weatherResult{
		City: "Paris", TempC: 18, Condition: "cloudy",
	}))

	req := &llm.GenerateRequest{Messages: []llm.Message{llm.UserMessage(llm.TextPart("weather in paris?"))}}
	got, err := llm.GenerateStructured[weatherResult](context.Background(), provider, req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := weatherResult{City: "Paris", TempC: 18, Condition: "cloudy"}
	if got != want {
		t.Fatalf("expected %+v, got %+v", want, got)
	}

	calls := provider.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].ToolChoice == nil || calls[0].ToolChoice.Mode != llm.ToolChoiceRequired || calls[0].ToolChoice.Name != "emit_result" {
		t.Fatalf("expected forced emit_result tool choice, got %+v", calls[0].ToolChoice)
	}
	if len(calls[0].Tools) != 1 || calls[0].Tools[0].Name != "emit_result" {
		t.Fatalf("expected a single emit_result tool definition, got %+v", calls[0].Tools)
	}
}

func TestGenerateStructuredRequiresProviderAndRequest(t *testing.T) {
	if _, err := llm.GenerateStructured[weatherResult](context.Background(), nil, &llm.GenerateRequest{}); err == nil {
		t.Fatal("expected error for nil provider")
	}
	if _, err := llm.GenerateStructured[weatherResult](context.Background(), llmtest.NewMockProvider(), nil); err == nil {
		t.Fatal("expected error for nil request")
	}
}

func TestGenerateStructuredPropagatesProviderError(t *testing.T) {
	wantErr := errors.New("boom")
	provider := llmtest.NewMockProvider().AddError(wantErr)

	_, err := llm.GenerateStructured[weatherResult](context.Background(), provider, &llm.GenerateRequest{
		Messages: []llm.Message{llm.UserMessage(llm.TextPart("hi"))},
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped provider error, got %v", err)
	}
}

func TestGenerateStructuredErrorsWithoutMatchingToolCall(t *testing.T) {
	provider := llmtest.NewMockProvider().AddResponse(llmtest.TextResponse("no tool call here"))

	_, err := llm.GenerateStructured[weatherResult](context.Background(), provider, &llm.GenerateRequest{
		Messages: []llm.Message{llm.UserMessage(llm.TextPart("hi"))},
	})
	if err == nil {
		t.Fatal("expected error when the model does not return a structured tool call")
	}
}
