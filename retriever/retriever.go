// Package retriever adapts a vectorstore.VectorStore into a document
// retriever and provides a RAGChain that answers questions using retrieved
// context.
package retriever

import (
	"context"

	"github.com/singhJasvinder101/agentic-go/vectorstore"
)

// Retriever fetches the k documents most relevant to query.
type Retriever interface {
	Retrieve(ctx context.Context, query string, k int) ([]vectorstore.Document, error)
}

type storeRetriever struct {
	store vectorstore.VectorStore
}

// FromVectorStore adapts any vectorstore.VectorStore into a Retriever.
func FromVectorStore(store vectorstore.VectorStore) Retriever {
	return &storeRetriever{store: store}
}

func (r *storeRetriever) Retrieve(ctx context.Context, query string, k int) ([]vectorstore.Document, error) {
	results, err := r.store.SimilaritySearch(ctx, query, k)
	if err != nil {
		return nil, err
	}
	docs := make([]vectorstore.Document, len(results))
	for i, res := range results {
		docs[i] = res.Document
	}
	return docs, nil
}
