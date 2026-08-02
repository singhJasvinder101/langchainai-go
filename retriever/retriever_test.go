package retriever_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/singhJasvinder101/agentic-go/internal/llmtest"
	"github.com/singhJasvinder101/agentic-go/retriever"
	"github.com/singhJasvinder101/agentic-go/vectorstore"
	"github.com/singhJasvinder101/agentic-go/vectorstore/memory"
)

// keywordEmbedder is a deterministic fake embedder for tests: each dimension
// is the occurrence count of one vocabulary term, so documents sharing terms
// with a query score higher under cosine similarity than unrelated ones.
type keywordEmbedder struct {
	vocab []string
}

func (e keywordEmbedder) embed(text string) []float32 {
	lower := strings.ToLower(text)
	vec := make([]float32, len(e.vocab))
	for i, term := range e.vocab {
		vec[i] = float32(strings.Count(lower, term))
	}
	return vec
}

func (e keywordEmbedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i, text := range texts {
		out[i] = e.embed(text)
	}
	return out, nil
}

func (e keywordEmbedder) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	return e.embed(text), nil
}

func newFixtureStore(t *testing.T) vectorstore.VectorStore {
	t.Helper()
	emb := keywordEmbedder{vocab: []string{"france", "paris", "germany", "berlin", "capital"}}
	store, err := memory.New(emb)
	if err != nil {
		t.Fatalf("create memory store: %v", err)
	}
	ctx := context.Background()
	err = store.AddDocuments(ctx, []vectorstore.Document{
		{ID: "france", Content: "Paris is the capital of France."},
		{ID: "germany", Content: "Berlin is the capital of Germany."},
	})
	if err != nil {
		t.Fatalf("add documents: %v", err)
	}
	return store
}

func TestFromVectorStoreRetrieve(t *testing.T) {
	r := retriever.FromVectorStore(newFixtureStore(t))

	docs, err := r.Retrieve(context.Background(), "What is the capital of France?", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(docs) != 1 {
		t.Fatalf("expected 1 document, got %d", len(docs))
	}
	if !strings.Contains(docs[0].Content, "France") {
		t.Fatalf("expected the France document to rank first, got %q", docs[0].Content)
	}
}

func TestRAGChainRunAnswersUsingRetrievedContext(t *testing.T) {
	r := retriever.FromVectorStore(newFixtureStore(t))
	provider := llmtest.NewMockProvider().AddResponse(llmtest.TextResponse("Paris"))

	rag := retriever.NewRAGChain(r, provider, 1)
	got, err := rag.Run(context.Background(), map[string]any{"question": "What is the capital of France?"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Paris" {
		t.Fatalf("expected %q, got %q", "Paris", got)
	}

	sent := provider.Calls()[0].Messages
	if len(sent) != 2 {
		t.Fatalf("expected system + user message, got %d", len(sent))
	}
	if !strings.Contains(sent[0].Parts[0].Text, "Paris is the capital of France.") {
		t.Fatalf("expected retrieved context in system prompt, got %q", sent[0].Parts[0].Text)
	}
	if sent[1].Parts[0].Text != "What is the capital of France?" {
		t.Fatalf("expected question as user message, got %q", sent[1].Parts[0].Text)
	}
}

func TestRAGChainRunRequiresQuestion(t *testing.T) {
	rag := retriever.NewRAGChain(retriever.FromVectorStore(newFixtureStore(t)), llmtest.NewMockProvider(), 1)
	if _, err := rag.Run(context.Background(), map[string]any{}); err == nil {
		t.Fatal("expected error when question is missing")
	}
}

func TestRAGChainRunRequiresRetrieverAndProvider(t *testing.T) {
	if _, err := (&retriever.RAGChain{Provider: llmtest.NewMockProvider()}).Run(context.Background(), map[string]any{"question": "x"}); err == nil {
		t.Fatal("expected error for nil retriever")
	}
	if _, err := (&retriever.RAGChain{Retriever: retriever.FromVectorStore(newFixtureStore(t))}).Run(context.Background(), map[string]any{"question": "x"}); err == nil {
		t.Fatal("expected error for nil provider")
	}
}

type failingRetriever struct{}

func (failingRetriever) Retrieve(ctx context.Context, query string, k int) ([]vectorstore.Document, error) {
	return nil, errors.New("boom")
}

func TestRAGChainRunPropagatesRetrieverError(t *testing.T) {
	rag := retriever.NewRAGChain(failingRetriever{}, llmtest.NewMockProvider(), 1)
	_, err := rag.Run(context.Background(), map[string]any{"question": "x"})
	if err == nil || !strings.Contains(err.Error(), "boom") {
		t.Fatalf("expected wrapped retriever error, got %v", err)
	}
}
