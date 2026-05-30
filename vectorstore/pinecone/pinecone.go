package pinecone

import (
	"context"
	"fmt"
	"strings"

	pinecone "github.com/pinecone-io/go-pinecone/v5/pinecone"
	"github.com/singhJasvinder101/agentic-go/embedder"
	"github.com/singhJasvinder101/agentic-go/vectorstore"
	"github.com/singhJasvinder101/agentic-go/vectorstore/internal/storeutil"
)

type Options struct {
	APIKey    string
	IndexName string
	Namespace string
}

type Store struct {
	embedder embedder.Embedder
	index    *pinecone.IndexConnection
}

func New(e embedder.Embedder, opts Options) (*Store, error) {
	if e == nil {
		return nil, fmt.Errorf("pinecone: embedder is required")
	}
	if strings.TrimSpace(opts.APIKey) == "" {
		return nil, fmt.Errorf("pinecone: api key is required")
	}
	if strings.TrimSpace(opts.IndexName) == "" {
		return nil, fmt.Errorf("pinecone: index name is required")
	}

	client, err := pinecone.NewClient(pinecone.NewClientParams{ApiKey: opts.APIKey})
	if err != nil {
		return nil, err
	}

	index, err := client.DescribeIndex(context.Background(), opts.IndexName)
	if err != nil {
		return nil, err
	}

	conn, err := client.Index(pinecone.NewIndexConnParams{
		Host:      index.Host,
		Namespace: opts.Namespace,
	})
	if err != nil {
		return nil, err
	}

	return &Store{
		embedder: e,
		index:    conn,
	}, nil
}

func (s *Store) AddDocuments(ctx context.Context, docs []vectorstore.Document) error {
	prepared, err := storeutil.PrepareDocuments(ctx, s.embedder, docs)
	if err != nil {
		return err
	}

	vectors := make([]*pinecone.Vector, len(prepared.Documents))
	for i, doc := range prepared.Documents {
		values := prepared.Vectors[i]
		metadata, err := pinecone.NewMetadata(payloadToInterfaceMap(storeutil.Payload(doc)))
		if err != nil {
			return err
		}

		vectors[i] = &pinecone.Vector{
			Id:       doc.ID,
			Values:   &values,
			Metadata: metadata,
		}
	}

	_, err = s.index.UpsertVectors(ctx, vectors)
	return err
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

	response, err := s.index.QueryByVectorValues(ctx, &pinecone.QueryByVectorValuesRequest{
		Vector:          queryVector,
		TopK:            uint32(k),
		IncludeMetadata: true,
	})
	if err != nil {
		return nil, err
	}

	results := make([]vectorstore.SearchResult, 0, len(response.Matches))
	for _, match := range response.Matches {
		if match == nil || match.Vector == nil {
			continue
		}
		results = append(results, vectorstore.SearchResult{
			Document: documentFromVector(match.Vector),
			Score:    match.Score,
		})
	}
	return results, nil
}

func (s *Store) Delete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return vectorstore.ErrIDsRequired
	}
	return s.index.DeleteVectorsById(ctx, ids)
}

func documentFromVector(vector *pinecone.Vector) vectorstore.Document {
	if vector.Metadata == nil {
		return vectorstore.Document{ID: vector.Id, Metadata: map[string]string{}}
	}
	return storeutil.DocumentFromPayload(vector.Id, vector.Metadata.AsMap())
}

func payloadToInterfaceMap(payload map[string]any) map[string]interface{} {
	out := make(map[string]interface{}, len(payload))
	for key, value := range payload {
		out[key] = value
	}
	return out
}
