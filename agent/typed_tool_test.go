package agent_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/singhJasvinder101/agentic-go/agent"
	"github.com/singhJasvinder101/agentic-go/internal/llmtest"
	"github.com/singhJasvinder101/agentic-go/llm"
)

type typedWeatherArgs struct {
	City string `json:"city" jsonschema_description:"City name, e.g. Paris"`
}

func TestNewTypedToolBuildsSchemaFromStructTags(t *testing.T) {
	tool := agent.NewTypedTool("get_weather", "Get current weather for a city",
		func(ctx context.Context, args typedWeatherArgs) (string, error) {
			return "ok", nil
		})

	if tool.Definition.Name != "get_weather" {
		t.Fatalf("expected tool name get_weather, got %q", tool.Definition.Name)
	}

	var schema map[string]any
	if err := json.Unmarshal(tool.Definition.Parameters, &schema); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}
	props, _ := schema["properties"].(map[string]any)
	city, _ := props["city"].(map[string]any)
	if city["description"] != "City name, e.g. Paris" {
		t.Fatalf("expected description reflected from the struct tag, got %v", city)
	}
}

func TestNewTypedToolHandlerReceivesDecodedArgs(t *testing.T) {
	var got typedWeatherArgs
	tool := agent.NewTypedTool("get_weather", "Get current weather for a city",
		func(ctx context.Context, args typedWeatherArgs) (string, error) {
			got = args
			return `{"temp_c":18}`, nil
		})

	result, err := tool.Handler(context.Background(), json.RawMessage(`{"city":"Paris"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.City != "Paris" {
		t.Fatalf("expected decoded city=Paris, got %+v", got)
	}
	if result != `{"temp_c":18}` {
		t.Fatalf("unexpected result: %q", result)
	}
}

func TestNewTypedToolHandlerReturnsDecodeError(t *testing.T) {
	tool := agent.NewTypedTool("get_weather", "desc",
		func(ctx context.Context, args typedWeatherArgs) (string, error) { return "", nil })

	if _, err := tool.Handler(context.Background(), json.RawMessage(`not json`)); err == nil {
		t.Fatal("expected decode error for malformed arguments")
	}
}

func TestNewTypedToolIntegratesWithAgentRun(t *testing.T) {
	tool := agent.NewTypedTool("get_weather", "Get current weather for a city",
		func(ctx context.Context, args typedWeatherArgs) (string, error) {
			return fmt.Sprintf(`{"temp_c":18,"city":%q}`, args.City), nil
		})

	provider := llmtest.NewMockProvider().
		AddResponse(llmtest.ToolCallResponse(llm.ToolCall{ID: "call_1", Name: "get_weather", Arguments: json.RawMessage(`{"city":"Paris"}`)})).
		AddResponse(llmtest.TextResponse("It's 18C in Paris."))

	a := agent.New(provider, agent.WithTools(tool))
	result, err := a.Run(context.Background(), "weather in paris?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text != "It's 18C in Paris." {
		t.Fatalf("unexpected result: %q", result.Text)
	}
}
