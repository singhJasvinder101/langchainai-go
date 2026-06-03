package agenticgo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/singhJasvinder101/agentic-go/llm"
	"github.com/singhJasvinder101/agentic-go/llm/claude"
	"github.com/singhJasvinder101/agentic-go/llm/gemini"
	ollamallm "github.com/singhJasvinder101/agentic-go/llm/ollama"
	openaillm "github.com/singhJasvinder101/agentic-go/llm/openai"
)

type llmGenerator interface {
	Generate(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error)
}

//json schema object + properties + required when you care about shape.
var weatherToolSchema = json.RawMessage(`{
	"type": "object",
	"properties": {
		"city": {
			"type": "string",
			"description": "City name, e.g. Paris"
		}
	},
	"required": ["city"]
}`)

func weatherTools() []llm.Tool {
	return []llm.Tool{
		llm.NewTool("get_weather", "Get current weather for a city", weatherToolSchema),
	}
}

type weatherArgs struct {
	City string `json:"city"`
}

// getWeather is the tool implementation used in tests (deterministic stub, no external API).
func getWeather(args weatherArgs) (string, error) {
	city := strings.TrimSpace(args.City)
	if city == "" {
		return "", fmt.Errorf("city is required")
	}
	tempC := 22
	condition := "sunny"
	switch strings.ToLower(city) {
	case "paris":
		tempC, condition = 18, "partly cloudy"
	case "london":
		tempC, condition = 14, "rainy"
	case "tokyo":
		tempC, condition = 26, "humid"
	}
	out, err := json.Marshal(map[string]any{
		"city":      city,
		"temp_c":    tempC,
		"condition": condition,
		"unit":      "celsius",
	})
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func executeWeatherToolCall(call llm.ToolCall) (string, error) {
	if call.Name != "get_weather" {
		return "", fmt.Errorf("unexpected tool %q", call.Name)
	}
	var args weatherArgs
	if err := call.ParseArguments(&args); err != nil {
		return "", err
	}
	return getWeather(args)
}

func toolCallID(call llm.ToolCall) string {
	if call.ID != "" {
		return call.ID
	}
	return "call_" + call.Name
}

// skipIfProviderUnavailable skips when the failure is credentials, quota, or connectivity.
func skipIfProviderUnavailable(t *testing.T, provider string, err error) {
	t.Helper()
	if err == nil {
		return
	}
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "401"),
		strings.Contains(msg, "403"),
		strings.Contains(msg, "429"),
		strings.Contains(msg, "quota"),
		strings.Contains(msg, "authentication"),
		strings.Contains(msg, "unauthorized"),
		strings.Contains(msg, "api key"),
		strings.Contains(msg, "does not support tools"):
		t.Skipf("%s unavailable: %v", provider, err)
	}
}

// runWeatherToolCalling exercises a full tool loop: model calls get_weather, we respond, model answers.
func runWeatherToolCalling(t *testing.T, provider string, gen llmGenerator) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	tools := weatherTools()

	messages := []llm.Message{
		llm.UserMessage(llm.TextPart(
			"What is the current weather in Paris? You must use the get_weather tool to look it up, then summarize the temperature in Celsius in one sentence.",
		)),
	}

	fmt.Println(tools[0].Parameters)

	first, err := gen.Generate(ctx, &llm.GenerateRequest{
		Messages:   messages,
		Tools:      tools,
		ToolChoice: &llm.ToolChoice{Mode: llm.ToolChoiceRequired},
	})
	if err != nil {
		skipIfProviderUnavailable(t, provider, err)
		if provider == "ollama" {
			t.Skipf("ollama generate failed (is ollama running?): %v", err)
		}
		t.Fatalf("%s first generate failed: %v", provider, err)
	}

	calls := first.ToolCalls()
	if len(calls) == 0 {
		if provider == "ollama" {
			t.Skipf("ollama model did not return tool calls (use a tool-capable model): %q", first.Text())
		}
		t.Fatalf("%s: expected tool calls, got assistant message: %q", provider, first.Text())
	}

	t.Logf("%s first turn tool calls: %+v", provider, calls)

	fmt.Println(calls[0].ArgumentsString())

	messages = append(messages, llm.AssistantToolCallsMessage(first.Text(), calls...))
	for _, call := range calls {
		result, err := executeWeatherToolCall(call)
		if err != nil {
			t.Fatalf("%s execute tool %q: %v", provider, call.Name, err)
		}
		t.Logf("%s tool result for %s: %s", provider, call.Name, result)
		messages = append(messages, llm.ToolMessage(toolCallID(call), call.Name, result))
	}

	second, err := gen.Generate(ctx, &llm.GenerateRequest{
		Messages: messages,
		Tools:    tools,
	})
	if err != nil {
		skipIfProviderUnavailable(t, provider, err)
		t.Fatalf("%s second generate failed: %v", provider, err)
	}

	answer := strings.ToLower(second.Text())
	if answer == "" {
		t.Fatalf("%s: expected final text answer, got empty response", provider)
	}
	if !strings.Contains(answer, "18") && !strings.Contains(answer, "paris") {
		t.Fatalf("%s: final answer should mention Paris weather (18°C); got: %q", provider, second.Text())
	}
	t.Logf("%s final answer: %s", provider, second.Text())
}

func TestOpenAIWeatherToolCalling(t *testing.T) {
	provider, err := openaillm.New()
	if err != nil {
		t.Fatalf("failed to create openai provider: %v", err)
	}
	runWeatherToolCalling(t, "openai", provider)
}

func TestGeminiWeatherToolCalling(t *testing.T) {
	ctx := context.Background()
	provider, err := gemini.New(ctx)
	if err != nil {
		t.Fatalf("failed to create gemini provider: %v", err)
	}
	runWeatherToolCalling(t, "gemini", provider)
}

func TestClaudeWeatherToolCalling(t *testing.T) {
	provider, err := claude.New()
	if err != nil {
		t.Fatalf("failed to create claude provider: %v", err)
	}
	runWeatherToolCalling(t, "claude", provider)
}

func TestOllamaWeatherToolCalling(t *testing.T) {
	ctx := context.Background()
	provider, err := ollamallm.New(ctx)
	if err != nil {
		t.Fatalf("failed to create ollama provider: %v", err)
	}
	runWeatherToolCalling(t, "ollama", provider)
}
