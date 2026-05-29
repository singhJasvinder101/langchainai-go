package memory_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/singhJasvinder101/langchainai-go/vectorstore"
	"github.com/singhJasvinder101/langchainai-go/vectorstore/memory"
)

type mockEmbedder struct {
	vectors map[string][]float32
}

func (m *mockEmbedder) EmbedDocuments(_ context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, errors.New("texts are required")
	}
	out := make([][]float32, len(texts))
	for i, text := range texts {
		vec, ok := m.vectors[text]
		if !ok {
			return nil, errors.New("unknown text")
		}
		out[i] = vec
	}
	return out, nil
}

func (m *mockEmbedder) EmbedQuery(_ context.Context, text string) ([]float32, error) {
	vec, ok := m.vectors[text]
	if !ok {
		return nil, errors.New("unknown text")
	}
	return vec, nil
}

func (m *mockEmbedder) Close() error { return nil }

func TestStoreSimilaritySearch(t *testing.T) {
	embed := &mockEmbedder{
		vectors: map[string][]float32{
			"paris doc":  {1, 0, 0},
			"berlin doc": {0, 1, 0},
			"paris?":     {0.9, 0.1, 0},
		},
	}

	store, err := memory.New(embed)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	ctx := context.Background()
	err = store.AddDocuments(ctx, []vectorstore.Document{
		{ID: "paris", Content: "paris doc"},
		{ID: "berlin", Content: "berlin doc"},
	})
	if err != nil {
		t.Fatalf("failed to add documents: %v", err)
	}

	results, err := store.SimilaritySearch(ctx, "paris?", 1)
	if err != nil {
		t.Fatalf("similarity search failed: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Document.ID != "paris" {
		t.Fatalf("expected paris document, got %q", results[0].Document.ID)
	}
}

func TestStoreValidation(t *testing.T) {
	embed := &mockEmbedder{vectors: map[string][]float32{}}
	store, err := memory.New(embed)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	ctx := context.Background()
	if err := store.AddDocuments(ctx, nil); !errors.Is(err, vectorstore.ErrDocumentsRequired) {
		t.Fatalf("expected ErrDocumentsRequired, got %v", err)
	}
	if _, err := store.SimilaritySearch(ctx, "", 1); !errors.Is(err, vectorstore.ErrQueryRequired) {
		t.Fatalf("expected ErrQueryRequired, got %v", err)
	}
	if _, err := store.SimilaritySearch(ctx, "q", 0); !errors.Is(err, vectorstore.ErrInvalidK) {
		t.Fatalf("expected ErrInvalidK, got %v", err)
	}
	if err := store.Delete(ctx, nil); !errors.Is(err, vectorstore.ErrIDsRequired) {
		t.Fatalf("expected ErrIDsRequired, got %v", err)
	}
}

func TestStoreAssignsIDs(t *testing.T) {
	embed := &mockEmbedder{
		vectors: map[string][]float32{
			"only doc": {1, 0},
			"query":    {1, 0},
		},
	}
	store, err := memory.New(embed)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	ctx := context.Background()
	if err := store.AddDocuments(ctx, []vectorstore.Document{{Content: "only doc"}}); err != nil {
		t.Fatalf("add documents failed: %v", err)
	}

	results, err := store.SimilaritySearch(ctx, "query", 1)
	if err != nil {
		t.Fatalf("similarity search failed: %v", err)
	}
	if len(results) != 1 || strings.TrimSpace(results[0].Document.ID) == "" {
		t.Fatalf("expected generated document id, got %+v", results)
	}
}
