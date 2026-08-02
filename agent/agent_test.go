package agent_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/singhJasvinder101/agentic-go/agent"
	"github.com/singhJasvinder101/agentic-go/internal/llmtest"
	"github.com/singhJasvinder101/agentic-go/llm"
	"github.com/singhJasvinder101/agentic-go/memory"
)

type weatherArgs struct {
	City string `json:"city"`
}

func weatherTool(t *testing.T, handler agent.Handler) agent.Tool {
	t.Helper()
	return agent.NewTool("get_weather", "Get current weather for a city",
		json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}`), handler)
}

func TestAgentRunReturnsImmediateTextAnswer(t *testing.T) {
	provider := llmtest.NewMockProvider().AddResponse(llmtest.TextResponse("hello there"))
	a := agent.New(provider)

	result, err := a.Run(context.Background(), "hi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != "hello there" {
		t.Fatalf("expected %q, got %q", "hello there", result.Text)
	}
	if result.Steps != 1 {
		t.Fatalf("expected 1 step, got %d", result.Steps)
	}
}

func TestAgentRunExecutesToolCallThenReturnsFinalAnswer(t *testing.T) {
	var gotArgs weatherArgs
	tool := weatherTool(t, func(ctx context.Context, args json.RawMessage) (string, error) {
		if err := json.Unmarshal(args, &gotArgs); err != nil {
			return "", err
		}
		return `{"temp_c":18,"condition":"cloudy"}`, nil
	})

	provider := llmtest.NewMockProvider().
		AddResponse(llmtest.ToolCallResponse(llm.ToolCall{
			ID: "call_1", Name: "get_weather", Arguments: json.RawMessage(`{"city":"Paris"}`),
		})).
		AddResponse(llmtest.TextResponse("It's 18C and cloudy in Paris."))

	a := agent.New(provider, agent.WithTools(tool), agent.WithSystemPrompt("You are a weather assistant."))

	result, err := a.Run(context.Background(), "What's the weather in Paris?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotArgs.City != "Paris" {
		t.Fatalf("expected tool handler to receive city=Paris, got %+v", gotArgs)
	}
	if result.Text != "It's 18C and cloudy in Paris." {
		t.Fatalf("unexpected final answer: %q", result.Text)
	}
	if result.Steps != 2 {
		t.Fatalf("expected 2 steps, got %d", result.Steps)
	}

	calls := provider.Calls()
	if len(calls) != 2 {
		t.Fatalf("expected 2 generate calls, got %d", len(calls))
	}
	// The second call's history must include the tool result message.
	foundToolResult := false
	for _, msg := range calls[1].Messages {
		if msg.Role == llm.RoleTool {
			foundToolResult = true
		}
	}
	if !foundToolResult {
		t.Fatalf("expected tool result message in second call history, got %+v", calls[1].Messages)
	}
}

func TestAgentRunSurfacesHandlerErrorAsToolResult(t *testing.T) {
	tool := weatherTool(t, func(ctx context.Context, args json.RawMessage) (string, error) {
		return "", errors.New("upstream weather API down")
	})

	provider := llmtest.NewMockProvider().
		AddResponse(llmtest.ToolCallResponse(llm.ToolCall{ID: "call_1", Name: "get_weather", Arguments: json.RawMessage(`{"city":"Paris"}`)})).
		AddResponse(llmtest.TextResponse("Sorry, I couldn't fetch the weather."))

	a := agent.New(provider, agent.WithTools(tool))

	result, err := a.Run(context.Background(), "weather in paris?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	calls := provider.Calls()
	last := calls[len(calls)-1]
	var toolMsg *llm.Message
	for i := range last.Messages {
		if last.Messages[i].Role == llm.RoleTool {
			toolMsg = &last.Messages[i]
		}
	}
	if toolMsg == nil {
		t.Fatal("expected a tool result message")
	}
	if got := toolMsg.Parts[0].Text; got != "error: upstream weather API down" {
		t.Fatalf("expected handler error surfaced as tool result, got %q", got)
	}
	if result.Text == "" {
		t.Fatal("expected the agent to still produce a final answer")
	}
}

func TestAgentRunUnknownToolNameSurfacesAsError(t *testing.T) {
	provider := llmtest.NewMockProvider().
		AddResponse(llmtest.ToolCallResponse(llm.ToolCall{ID: "call_1", Name: "does_not_exist", Arguments: json.RawMessage(`{}`)})).
		AddResponse(llmtest.TextResponse("done"))

	a := agent.New(provider)

	if _, err := a.Run(context.Background(), "hi"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	last := provider.Calls()[1]
	found := false
	for _, msg := range last.Messages {
		if msg.Role == llm.RoleTool && msg.Parts[0].Text == `error: unknown tool "does_not_exist"` {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected unknown tool error surfaced as tool result, got %+v", last.Messages)
	}
}

func TestAgentRunReturnsErrMaxIterations(t *testing.T) {
	provider := llmtest.NewMockProvider()
	for i := 0; i < 3; i++ {
		provider.AddResponse(llmtest.ToolCallResponse(llm.ToolCall{ID: "call", Name: "loop", Arguments: json.RawMessage(`{}`)}))
	}
	tool := weatherTool(t, func(ctx context.Context, args json.RawMessage) (string, error) { return "ok", nil })

	a := agent.New(provider, agent.WithTools(tool), agent.WithMaxIterations(3))

	_, err := a.Run(context.Background(), "hi")
	if !errors.Is(err, agent.ErrMaxIterations) {
		t.Fatalf("expected ErrMaxIterations, got %v", err)
	}
}

func TestAgentRunPersistsToMemoryOnSuccess(t *testing.T) {
	mem := memory.NewBuffer()
	mem.Add(llm.UserMessage(llm.TextPart("earlier turn")))

	provider := llmtest.NewMockProvider().AddResponse(llmtest.TextResponse("final answer"))
	a := agent.New(provider, agent.WithMemory(mem))

	if _, err := a.Run(context.Background(), "new question"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The prior turn should have been included in the request sent to the provider.
	sent := provider.Calls()[0].Messages
	if sent[0].Parts[0].Text != "earlier turn" {
		t.Fatalf("expected prior memory included in request, got %+v", sent)
	}

	// And the new turn should now be persisted.
	got := mem.Messages()
	if len(got) != 3 {
		t.Fatalf("expected 3 messages in memory (earlier turn + new user + new assistant), got %d: %+v", len(got), got)
	}
}

func TestAgentRunRequiresProvider(t *testing.T) {
	a := agent.New(nil)
	if _, err := a.Run(context.Background(), "hi"); err == nil {
		t.Fatal("expected error for nil provider")
	}
}
