package memory

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/singhJasvinder101/langchainai-go/embedder"
	"github.com/singhJasvinder101/langchainai-go/vectorstore"
	"github.com/singhJasvinder101/langchainai-go/vectorstore/internal/storeutil"
)

type storedDocument struct {
	doc    vectorstore.Document
	vector []float32
}

// Store is an in-memory vector store backed by cosine similarity.
type Store struct {
	embedder embedder.Embedder
	mu       sync.RWMutex
	docs     map[string]storedDocument
}

func New(e embedder.Embedder) (*Store, error) {
	if e == nil {
		return nil, fmt.Errorf("memory: embedder is required")
	}
	return &Store{
		embedder: e,
		docs:     make(map[string]storedDocument),
	}, nil
}

func (s *Store) AddDocuments(ctx context.Context, docs []vectorstore.Document) error {
	prepared, err := storeutil.PrepareDocuments(ctx, s.embedder, docs)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, doc := range prepared.Documents {
		s.docs[doc.ID] = storedDocument{
			doc:    doc,
			vector: prepared.Vectors[i],
		}
	}
	return nil
}

func (s *Store) SimilaritySearch(ctx context.Context, query string, k int) ([]vectorstore.SearchResult, error) {
	if query == "" {
		return nil, vectorstore.ErrQueryRequired
	}
	if k <= 0 {
		return nil, vectorstore.ErrInvalidK
	}

	queryVector, err := s.embedder.EmbedQuery(ctx, query)
	if err != nil {
		return nil, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.docs) == 0 {
		return nil, nil
	}

	results := make([]vectorstore.SearchResult, 0, len(s.docs))
	for _, stored := range s.docs {
		results = append(results, vectorstore.SearchResult{
			Document: stored.doc,
			Score:    cosineSimilarity(queryVector, stored.vector),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if len(results) > k {
		results = results[:k]
	}
	return results, nil
}

func (s *Store) Delete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return vectorstore.ErrIDsRequired
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, id := range ids {
		delete(s.docs, id)
	}
	return nil
}

func cosineSimilarity(a, b []float32) float32 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}

	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return float32(dot / (math.Sqrt(normA) * math.Sqrt(normB)))
}
