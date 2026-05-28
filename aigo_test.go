package aigo

import (
	"context"
	"strings"
	"testing"

	"github.com/singhJasvinder101/ai-go/internal/constants"
	"github.com/singhJasvinder101/ai-go/internal/llm/providers/gemini"
	"github.com/singhJasvinder101/ai-go/internal/llm/providers/ollama"
	"github.com/singhJasvinder101/ai-go/internal/llm/providers/openai"
	"github.com/singhJasvinder101/ai-go/internal/template"
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

func TestTemplate(t *testing.T) {
	registry := template.NewRegistry()
	err := registry.RegisterTemplate("summary", template.FormatterNative, "Summarize the following text: {{.Text}}")
	if err != nil {
		t.Fatalf("failed to register template: %v", err)
	}
	formatted, err := registry.Format("summary", map[string]any{"Text": "The quick brown fox jumps over the lazy dog."})
	if err != nil {
		t.Fatalf("failed to format template: %v", err)
	}
	t.Log(formatted)
}

func TestTemplateJinja(t *testing.T) {
	registry := template.NewRegistry()
	err := registry.RegisterTemplate("summary", template.FormatterJinja, "Summarize the following text: {{ text }}")
	if err != nil {
		t.Fatalf("failed to register template: %v", err)
	}
	formatted, err := registry.Format("summary", map[string]any{"text": "The quick brown fox jumps over the lazy dog."})
	if err != nil {
		t.Fatalf("failed to format template: %v", err)
	}
	t.Log(formatted)
}

func TestComplexNestedTemplate(t *testing.T) {
	registry := template.NewRegistry()
	err := registry.RegisterTemplate("summary", template.FormatterNative, "Summarize the following text: {{.Text}} {{.Nested.Text}}")
	if err != nil {
		t.Fatalf("failed to register template: %v", err)
	}
	formatted, err := registry.Format("summary", map[string]any{"Text": "The quick brown fox jumps over the lazy dog.", "Nested": map[string]any{"Text": "The quick brown fox jumps over the lazy dog."}})
	if err != nil {
		t.Fatalf("failed to format template: %v", err)
	}
	t.Log(formatted)
}


func TestNewEmbeddingsWithoutDocuments(t *testing.T) {
	ctx := context.Background()
	provider, err := NewEmbeddings(ctx, constants.ProviderOllama, "configs/config.yaml")
	if err != nil {
		t.Fatalf("failed to create embeddings provider: %v", err)
	}
	defer provider.Close()

	if provider == nil {
		t.Fatal("expected non-nil embeddings provider")
	}

	_, err = provider.EmbedDocuments(ctx, nil)
	if err == nil || !strings.Contains(err.Error(), "texts are required") {
		t.Fatalf("expected documents validation error, got %v", err)
	}

	_, err = provider.EmbedQuery(ctx, "")
	if err == nil || !strings.Contains(err.Error(), "text is required") {
		t.Fatalf("expected query validation error, got %v", err)
	}
}

func TestNewEmbeddingsWithDocuments(t *testing.T) {
	ctx := context.Background()
	provider, err := NewEmbeddings(ctx, constants.ProviderOllama, "configs/config.yaml")
	if err != nil {
		t.Fatalf("failed to create embeddings provider: %v", err)
	}
	defer provider.Close()

	if provider == nil {
		t.Fatal("expected non-nil embeddings provider")
	}

	embeddings, err := provider.EmbedDocuments(ctx, []string{"Hello, world!"})
	if err != nil {
		t.Fatalf("failed to embed documents: %v", err)
	}

	t.Log(embeddings)

	query, err := provider.EmbedQuery(ctx, "Hello, world!")
	if err != nil {
		t.Fatalf("failed to embed query: %v", err)
	}
	
	t.Log(query)
}

func TestNewEmbeddingsUnsupportedProvider(t *testing.T) {
	ctx := context.Background()
	provider, err := NewEmbeddings(ctx, constants.ProviderClaude, "configs/config.yaml")
	if err == nil {
		_ = provider.Close()
		t.Fatal("expected error for unsupported embeddings provider")
	}
	if !strings.Contains(err.Error(), "does not support embeddings") {
		t.Fatalf("expected unsupported embeddings error, got %v", err)
	}
}