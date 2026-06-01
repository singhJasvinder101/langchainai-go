package agenticgo

import (
	"context"
	"os"
	"strings"
	"testing"

	geminiEmbedder "github.com/singhJasvinder101/agentic-go/embedder/gemini"
	ollamaEmbedder "github.com/singhJasvinder101/agentic-go/embedder/ollama"
	"github.com/singhJasvinder101/agentic-go/init/config"
	"github.com/singhJasvinder101/agentic-go/llm"
	"github.com/singhJasvinder101/agentic-go/llm/gemini"
	ollamallm "github.com/singhJasvinder101/agentic-go/llm/ollama"
	openaillm "github.com/singhJasvinder101/agentic-go/llm/openai"
	"github.com/singhJasvinder101/agentic-go/template"
	"github.com/singhJasvinder101/agentic-go/vectorstore"
	"github.com/singhJasvinder101/agentic-go/vectorstore/chroma"
	"github.com/singhJasvinder101/agentic-go/vectorstore/memory"
)

func TestMain(m *testing.M) {
	config.MustInit(config.DefaultConfigPath)
	os.Exit(m.Run())
}

func TestGeminiGenerate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	geminiProvider, err := gemini.New(ctx)
	if err != nil {
		t.Fatalf("failed to create gemini provider: %v", err)
	}

	response, err := geminiProvider.Generate(ctx, &llm.GenerateRequest{
		Messages: []llm.Message{llm.UserMessage(llm.TextPart("Hello, how are you?"))},
	})
	if err != nil {
		t.Fatalf("failed to generate response: %v", err)
	}
	t.Log(response.AssistantMessage())
}

func TestOpenAIGenerate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	provider, err := openaillm.New()
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	response, err := provider.Generate(ctx, &llm.GenerateRequest{
		Messages: []llm.Message{llm.UserMessage(llm.TextPart("Hello, how are you?"))},
	})
	if err != nil {
		t.Fatalf("failed to generate response: %v", err)
	}
	t.Log(response.AssistantMessage())
}

func TestOllamaGenerate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	provider, err := ollamallm.New(ctx)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	response, err := provider.Generate(ctx, &llm.GenerateRequest{
		Messages: []llm.Message{llm.UserMessage(llm.TextPart("Say hello in one short sentence also give me some random image urls."))},
	})
	if err != nil {
		t.Skipf("ollama generate failed (is ollama running?): %v", err)
	}
	if response.Text() == "" {
		t.Fatal("expected non-empty response from ollama")
	}
	t.Log(response.AssistantMessage())
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

func TestEmbedderWithoutDocuments(t *testing.T) {
	ctx := context.Background()
	emb, err := ollamaEmbedder.New(ctx)
	if err != nil {
		t.Fatalf("failed to create embedder: %v", err)
	}

	_, err = emb.EmbedDocuments(ctx, nil)
	if err == nil || !strings.Contains(err.Error(), "texts are required") {
		t.Fatalf("expected documents validation error, got %v", err)
	}

	_, err = emb.EmbedQuery(ctx, "")
	if err == nil || !strings.Contains(err.Error(), "text is required") {
		t.Fatalf("expected query validation error, got %v", err)
	}
}

func TestEmbedderWithDocuments(t *testing.T) {
	ctx := context.Background()
	emb, err := ollamaEmbedder.New(ctx)
	if err != nil {
		t.Fatalf("failed to create embedder: %v", err)
	}

	embeddings, err := emb.EmbedDocuments(ctx, []string{"Hello, world!"})
	if err != nil {
		t.Fatalf("failed to embed documents: %v", err)
	}

	t.Log(embeddings)

	query, err := emb.EmbedQuery(ctx, "Hello, world!")
	if err != nil {
		t.Fatalf("failed to embed query: %v", err)
	}

	t.Log(query)
}

func TestMemoryVectorStoreWithEmbedder(t *testing.T) {
	ctx := context.Background()
	emb, err := ollamaEmbedder.New(ctx)
	if err != nil {
		t.Fatalf("failed to create embedder: %v", err)
	}

	store, err := memory.New(emb)
	if err != nil {
		t.Fatalf("failed to create vector store: %v", err)
	}

	err = store.AddDocuments(ctx, []vectorstore.Document{
		{ID: "france", Content: "Paris is the capital of France."},
		{ID: "germany", Content: "Berlin is the capital of Germany."},
	})
	if err != nil {
		t.Skipf("add documents failed (is ollama running?): %v", err)
	}

	results, err := store.SimilaritySearch(ctx, "What is the capital of France?", 1)
	if err != nil {
		t.Fatalf("similarity search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one search result")
	}
	t.Log(results[0].Document.Content, results[0].Score)
}

func TestChromaVectorStoreWithEmbedder(t *testing.T) {
	ctx := context.Background()
	emb, err := geminiEmbedder.New(ctx)
	if err != nil {
		t.Fatalf("failed to create embedder: %v", err)
	}

	store, err := chroma.New(ctx, emb, chroma.Options{
		Collection: t.Name(),
	})
	if err != nil {
		t.Fatalf("failed to create vector store: %v", err)
	}

	err = store.AddDocuments(ctx, []vectorstore.Document{
		{ID: "france", Content: "Paris is the capital of France."},
		{ID: "germany", Content: "Berlin is the capital of Germany."},
	})
	if err != nil {
		t.Skipf("add documents failed (is ollama running?): %v", err)
	}

	results, err := store.SimilaritySearch(ctx, "What is the capital of France?", 1)
	if err != nil {
		t.Fatalf("similarity search failed: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one search result")
	}
	t.Log(results[0].Document.Content, results[0].Score)
}

func TestMessageHistory(t *testing.T) {
	ctx := context.Background()
	provider, err := ollamallm.New(ctx)
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	history := []llm.Message{
		llm.SystemMessage(llm.TextPart("You are not a coding assistant so don't generate code, you are a helpful assistant.")),
		llm.UserMessage(llm.TextPart("generate code for 2 sum")),
	}
	response, err := provider.Generate(ctx, &llm.GenerateRequest{Messages: history})
	if err != nil {
		t.Fatalf("failed to generate response: %v", err)
	}
	t.Log(response.AssistantMessage())
}