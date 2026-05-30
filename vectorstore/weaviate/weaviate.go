package weaviate

import (
	"context"
	"fmt"
	"strings"

	"github.com/singhJasvinder101/agentic-go/embedder"
	"github.com/singhJasvinder101/agentic-go/vectorstore"
	"github.com/singhJasvinder101/agentic-go/vectorstore/internal/storeutil"
	"github.com/weaviate/weaviate-go-client/v5/weaviate"
	"github.com/weaviate/weaviate-go-client/v5/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"
)

const (
	defaultScheme = "http"
	defaultHost   = "localhost:8080"
)

type Options struct {
	Scheme    string
	Host      string
	APIKey    string
	ClassName string
}

type Store struct {
	embedder embedder.Embedder
	client   *weaviate.Client
	class    string
}

func New(e embedder.Embedder, opts Options) (*Store, error) {
	if e == nil {
		return nil, fmt.Errorf("weaviate: embedder is required")
	}
	if strings.TrimSpace(opts.ClassName) == "" {
		return nil, fmt.Errorf("weaviate: class name is required")
	}

	scheme := opts.Scheme
	if scheme == "" {
		scheme = defaultScheme
	}
	host := opts.Host
	if host == "" {
		host = defaultHost
	}

	cfg := weaviate.Config{
		Scheme: scheme,
		Host:   host,
	}
	if opts.APIKey != "" {
		cfg.Headers = map[string]string{"Authorization": "Bearer " + opts.APIKey}
	}

	client, err := weaviate.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return &Store{
		embedder: e,
		client:   client,
		class:    opts.ClassName,
	}, nil
}

func (s *Store) AddDocuments(ctx context.Context, docs []vectorstore.Document) error {
	prepared, err := storeutil.PrepareDocuments(ctx, s.embedder, docs)
	if err != nil {
		return err
	}

	for i, doc := range prepared.Documents {
		properties := propertiesFromDocument(doc)
		creator := s.client.Data().Creator().
			WithClassName(s.class).
			WithID(doc.ID).
			WithProperties(properties).
			WithVector(prepared.Vectors[i])

		if _, err := creator.Do(ctx); err != nil {
			return err
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

	nearVector := s.client.GraphQL().NearVectorArgBuilder().WithVector(queryVector)
	queryFields := []graphql.Field{
		{Name: storeutil.ContentPayloadKey},
		{Name: "_additional", Fields: []graphql.Field{
			{Name: "id"},
			{Name: "distance"},
		}},
	}

	response, err := s.client.GraphQL().Get().
		WithClassName(s.class).
		WithNearVector(nearVector).
		WithLimit(k).
		WithFields(queryFields...).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	return searchResultsFromGraphQL(s.class, response.Data)
}

func (s *Store) Delete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return vectorstore.ErrIDsRequired
	}

	for _, id := range ids {
		if err := s.client.Data().Deleter().
			WithClassName(s.class).
			WithID(id).
			Do(ctx); err != nil {
			return err
		}
	}
	return nil
}

func propertiesFromDocument(doc vectorstore.Document) map[string]interface{} {
	properties := map[string]interface{}{
		storeutil.ContentPayloadKey: doc.Content,
	}
	for key, value := range doc.Metadata {
		properties[key] = value
	}
	return properties
}

func searchResultsFromGraphQL(className string, data map[string]models.JSONObject) ([]vectorstore.SearchResult, error) {
	getData, ok := data["Get"]
	if !ok {
		return nil, nil
	}

	classData, ok := getData.(map[string]interface{})
	if !ok {
		return nil, nil
	}

	entries, ok := classData[className].([]interface{})
	if !ok {
		return nil, nil
	}

	results := make([]vectorstore.SearchResult, 0)
	for _, entry := range entries {
		object, ok := entry.(map[string]interface{})
		if !ok {
			continue
		}

		content, _ := object[storeutil.ContentPayloadKey].(string)
		metadata := map[string]string{}
		for key, value := range object {
			if key == storeutil.ContentPayloadKey || key == "_additional" {
				continue
			}
			if text, ok := value.(string); ok {
				metadata[key] = text
			}
		}

		id := ""
		score := float32(0)
		if additional, ok := object["_additional"].(map[string]interface{}); ok {
			if additionalID, ok := additional["id"].(string); ok {
				id = additionalID
			}
			if distance, ok := additional["distance"].(float64); ok {
				score = storeutil.DistanceToScore(float32(distance))
			}
		}

		results = append(results, vectorstore.SearchResult{
			Document: vectorstore.Document{
				ID:       id,
				Content:  content,
				Metadata: metadata,
			},
			Score: score,
		})
	}
	return results, nil
}
