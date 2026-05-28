package openai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	goopenai "github.com/sashabaranov/go-openai"
	"github.com/singhJasvinder101/langchainai-go/internal/llm"
)

func TestOpenAIProviderImplementsEmbeddingProvider(t *testing.T) {
	var _ llm.EmbeddingProvider = (*OpenAIProvider)(nil)
}

func TestOpenAIEmbedDocuments(t *testing.T) {
	var requestBody struct {
		Input []string `json:"input"`
		Model string   `json:"model"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/embeddings" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"object": "list",
			"model": "text-embedding-3-small",
			"data": [
				{"object": "embedding", "index": 1, "embedding": [3, 4]},
				{"object": "embedding", "index": 0, "embedding": [1, 2]}
			],
			"usage": {"prompt_tokens": 1, "total_tokens": 1}
		}`))
	}))
	defer server.Close()

	cfg := goopenai.DefaultConfig("test-token")
	cfg.BaseURL = server.URL + "/v1"
	provider := &OpenAIProvider{Client: goopenai.NewClientWithConfig(cfg)}

	embeddings, err := provider.EmbedDocuments(context.Background(), []string{"first", "second"})
	if err != nil {
		t.Fatalf("EmbedDocuments returned error: %v", err)
	}

	if !reflect.DeepEqual(requestBody.Input, []string{"first", "second"}) {
		t.Fatalf("unexpected input: %#v", requestBody.Input)
	}
	if requestBody.Model != "text-embedding-3-small" {
		t.Fatalf("unexpected model: %s", requestBody.Model)
	}

	expected := [][]float32{{1, 2}, {3, 4}}
	if !reflect.DeepEqual(embeddings, expected) {
		t.Fatalf("unexpected embeddings: %#v", embeddings)
	}
}

func TestOpenAIEmbedDocumentsValidation(t *testing.T) {
	provider := &OpenAIProvider{}

	_, err := provider.EmbedDocuments(context.Background(), nil)
	if err == nil || !strings.Contains(err.Error(), "texts are required") {
		t.Fatalf("expected texts required error, got %v", err)
	}

	_, err = provider.EmbedDocuments(context.Background(), []string{"ok", ""})
	if err == nil || !strings.Contains(err.Error(), "text at index 1 is required") {
		t.Fatalf("expected indexed text required error, got %v", err)
	}

	_, err = provider.EmbedQuery(context.Background(), "")
	if err == nil || !strings.Contains(err.Error(), "text is required") {
		t.Fatalf("expected text required error, got %v", err)
	}
}
