package chroma

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	chroma "github.com/amikos-tech/chroma-go/pkg/api/v2"
	chromaembeddings "github.com/amikos-tech/chroma-go/pkg/embeddings"
	"github.com/singhJasvinder101/langchainai-go/embedder"
	"github.com/singhJasvinder101/langchainai-go/vectorstore"
	"github.com/singhJasvinder101/langchainai-go/vectorstore/internal/storeutil"
)

const (
	defaultBaseURL = "http://localhost:8000"
)

type Options struct {
	BaseURL           string
	Collection        string
	Tenant            string
	Database          string
	EmbeddingFunction embedder.Embedder
}

type Store struct {
	embedder   embedder.Embedder
	client     chroma.Client
	collection chroma.Collection
}

func New(ctx context.Context, e embedder.Embedder, opts Options) (*Store, error) {
	if e == nil {
		return nil, fmt.Errorf("chroma: embedder is required")
	}
	if strings.TrimSpace(opts.Collection) == "" {
		return nil, fmt.Errorf("chroma: collection is required")
	}

	baseURL := opts.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	clientOpts := []chroma.ClientOption{chroma.WithBaseURL(baseURL)}

	switch {
	case opts.Tenant != "" && opts.Database != "":
		clientOpts = append(clientOpts, chroma.WithDatabaseAndTenant(opts.Database, opts.Tenant))
	case opts.Tenant != "":
		clientOpts = append(clientOpts, chroma.WithTenant(opts.Tenant))
	}

	client, err := chroma.NewHTTPClient(clientOpts...)
	if err != nil {
		return nil, err
	}

	chromaEmbedder := opts.EmbeddingFunction
	if chromaEmbedder == nil {
		chromaEmbedder = e
	}

	collection, err := client.GetOrCreateCollection(ctx, opts.Collection,
		chroma.WithEmbeddingFunctionCreate(ChromaEmbedderAdapterFunction{Embedder: chromaEmbedder}),
	)
	if err != nil {
		_ = client.Close()
		return nil, err
	}

	return &Store{
		embedder:   e,
		client:     client,
		collection: collection,
	}, nil
}

func (s *Store) AddDocuments(ctx context.Context, docs []vectorstore.Document) error {
	prepared, err := storeutil.PrepareDocuments(ctx, s.embedder, docs)
	if err != nil {
		return err
	}

	ids := make([]chroma.DocumentID, len(prepared.Documents))
	texts := make([]string, len(prepared.Documents))
	embs := make([]chromaembeddings.Embedding, len(prepared.Vectors))
	metas := make([]chroma.DocumentMetadata, len(prepared.Documents))

	for i, doc := range prepared.Documents {
		ids[i] = chroma.DocumentID(doc.ID)
		texts[i] = doc.Content
		embs[i] = chromaembeddings.NewEmbeddingFromFloat32(prepared.Vectors[i])
		metas[i] = metadataFromDocument(doc)
	}

	return s.collection.Add(ctx,
		chroma.WithIDs(ids...),
		chroma.WithTexts(texts...),
		chroma.WithEmbeddings(embs...),
		chroma.WithMetadatas(metas...),
	)
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

	results, err := s.collection.Query(ctx,
		chroma.WithQueryEmbeddings(chromaembeddings.NewEmbeddingFromFloat32(queryVector)),
		chroma.WithNResults(k),
		chroma.WithInclude(chroma.IncludeDocuments, chroma.IncludeMetadatas, chroma.IncludeDistances),
	)
	if err != nil {
		return nil, err
	}

	return searchResultsFromQuery(results)
}

func (s *Store) Delete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return vectorstore.ErrIDsRequired
	}

	chromaIDs := make([]chroma.DocumentID, len(ids))
	for i, id := range ids {
		chromaIDs[i] = chroma.DocumentID(id)
	}
	return s.collection.Delete(ctx, chroma.WithIDs(chromaIDs...))
}

func metadataFromDocument(doc vectorstore.Document) chroma.DocumentMetadata {
	if len(doc.Metadata) == 0 {
		return nil
	}
	attrs := make([]*chroma.MetaAttribute, 0, len(doc.Metadata))
	for key, value := range doc.Metadata {
		attrs = append(attrs, chroma.NewStringAttribute(key, value))
	}
	return chroma.NewDocumentMetadata(attrs...)
}

func searchResultsFromQuery(results chroma.QueryResult) ([]vectorstore.SearchResult, error) {
	if results == nil || results.CountGroups() == 0 {
		return nil, nil
	}

	ids := results.GetIDGroups()[0]
	documents := results.GetDocumentsGroups()[0]
	metadatas := results.GetMetadatasGroups()[0]
	distances := results.GetDistancesGroups()[0]

	out := make([]vectorstore.SearchResult, 0, len(ids))
	for i, id := range ids {
		content := ""
		if i < len(documents) && documents[i] != nil {
			content = documents[i].ContentString()
		}

		metadata := metadataToMap(metadatas, i)

		score := float32(0)
		if i < len(distances) {
			score = storeutil.DistanceToScore(float32(distances[i]))
		}

		out = append(out, vectorstore.SearchResult{
			Document: vectorstore.Document{
				ID:       string(id),
				Content:  content,
				Metadata: metadata,
			},
			Score: score,
		})
	}
	return out, nil
}

func metadataToMap(metadatas []chroma.DocumentMetadata, index int) map[string]string {
	if index >= len(metadatas) || metadatas[index] == nil {
		return map[string]string{}
	}

	raw, err := json.Marshal(metadatas[index])
	if err != nil {
		return map[string]string{}
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		return map[string]string{}
	}

	metadata := make(map[string]string, len(decoded))
	for key, value := range decoded {
		if text, ok := value.(string); ok {
			metadata[key] = text
		}
	}
	return metadata
}
