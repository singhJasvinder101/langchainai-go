package aigo

import (
	"context"
	"testing"

	"github.com/singhJasvinder101/ai-go/internal/constants"
	"github.com/singhJasvinder101/ai-go/internal/llm/providers/gemini"
	"github.com/singhJasvinder101/ai-go/internal/llm/providers/ollama"
	"github.com/singhJasvinder101/ai-go/internal/llm/providers/openai"
)

func TestNew(t *testing.T) {
	ctx := context.Background()
	provider, err := New(ctx, constants.ProviderOllama, "configs/config.yaml")
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
	defer provider.Close()
}

func TestGeminiGenerate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	provider, err := New(ctx, constants.ProviderGemini, "configs/config.yaml")
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
	defer provider.Close()

	response, err := provider.Generate(ctx, &gemini.GenerateRequest{Prompt: "Hello, how are you?"})
	if err != nil {
		t.Fatalf("failed to generate response: %v", err)
	}
	t.Log(response.GetText())
}

func TestOpenAIGenerate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	provider, err := New(ctx, constants.ProviderOpenAI, "configs/config.yaml")
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
	defer provider.Close()

	response, err := provider.Generate(ctx, &openai.GenerateRequest{Prompt: "Hello, how are you?"})
	if err != nil {
		t.Fatalf("failed to generate response: %v", err)
	}
	t.Log(response.GetText())
}

func TestOllamaGenerate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	provider, err := New(ctx, constants.ProviderOllama, "configs/config.yaml")
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}
	defer provider.Close()

	response, err := provider.Generate(ctx, &ollama.GenerateRequest{Prompt: "Say hello in one short sentence."})
	if err != nil {
		t.Skipf("ollama generate failed (is ollama running?): %v", err)
	}
	if response.GetText() == "" {
		t.Fatal("expected non-empty response from ollama")
	}
	t.Log(response.GetText())
}
