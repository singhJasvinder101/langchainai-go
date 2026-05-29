package qdrant

import (
	"context"
	"fmt"
	"strings"

	"github.com/qdrant/go-client/qdrant"
	"github.com/singhJasvinder101/langchainai-go/embedder"
	"github.com/singhJasvinder101/langchainai-go/vectorstore"
	"github.com/singhJasvinder101/langchainai-go/vectorstore/internal/storeutil"
)

const (
	defaultHost = "localhost"
	defaultPort = 6334
)

type Options struct {
	Host           string
	Port           int
	APIKey         string
	UseTLS         bool
	Collection     string
	VectorSize     uint64
	CreateIfAbsent bool
}

type Store struct {
	embedder   embedder.Embedder
	client     *qdrant.Client
	collection string
	vectorSize uint64
}

func New(e embedder.Embedder, opts Options) (*Store, error) {
	if e == nil {
		return nil, fmt.Errorf("qdrant: embedder is required")
	}
	if strings.TrimSpace(opts.Collection) == "" {
		return nil, fmt.Errorf("qdrant: collection is required")
	}

	host := opts.Host
	if host == "" {
		host = defaultHost
	}
	port := opts.Port
	if port == 0 {
		port = defaultPort
	}

	client, err := qdrant.NewClient(&qdrant.Config{
		Host:   host,
		Port:   port,
		APIKey: opts.APIKey,
		UseTLS: opts.UseTLS,
	})
	if err != nil {
		return nil, err
	}

	store := &Store{
		embedder:   e,
		client:     client,
		collection: opts.Collection,
		vectorSize: opts.VectorSize,
	}

	if opts.CreateIfAbsent && opts.VectorSize > 0 {
		if err := store.ensureCollection(context.Background()); err != nil {
			_ = client.Close()
			return nil, err
		}
	}

	return store, nil
}

func (s *Store) AddDocuments(ctx context.Context, docs []vectorstore.Document) error {
	prepared, err := storeutil.PrepareDocuments(ctx, s.embedder, docs)
	if err != nil {
		return err
	}

	if s.vectorSize == 0 && len(prepared.Vectors) > 0 {
		s.vectorSize = uint64(len(prepared.Vectors[0]))
	}
	if err := s.ensureCollection(ctx); err != nil {
		return err
	}

	points := make([]*qdrant.PointStruct, len(prepared.Documents))
	for i, doc := range prepared.Documents {
		points[i] = &qdrant.PointStruct{
			Id:      qdrant.NewID(doc.ID),
			Vectors: qdrant.NewVectors(prepared.Vectors[i]...),
			Payload: qdrant.NewValueMap(storeutil.Payload(doc)),
		}
	}

	_, err = s.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: s.collection,
		Points:         points,
	})
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

	points, err := s.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: s.collection,
		Query:          qdrant.NewQuery(queryVector...),
		Limit:          qdrant.PtrOf(uint64(k)),
		WithPayload:    qdrant.NewWithPayload(true),
	})
	if err != nil {
		return nil, err
	}

	results := make([]vectorstore.SearchResult, 0, len(points))
	for _, point := range points {
		id := pointIDString(point.GetId())
		results = append(results, vectorstore.SearchResult{
			Document: storeutil.DocumentFromPayload(id, payloadToMap(point.GetPayload())),
			Score:    point.GetScore(),
		})
	}
	return results, nil
}

func (s *Store) Delete(ctx context.Context, ids []string) error {
	if len(ids) == 0 {
		return vectorstore.ErrIDsRequired
	}

	pointIDs := make([]*qdrant.PointId, len(ids))
	for i, id := range ids {
		pointIDs[i] = qdrant.NewID(id)
	}

	_, err := s.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: s.collection,
		Points:         qdrant.NewPointsSelector(pointIDs...),
	})
	return err
}


func (s *Store) ensureCollection(ctx context.Context) error {
	exists, err := s.client.CollectionExists(ctx, s.collection)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	if s.vectorSize == 0 {
		return fmt.Errorf("qdrant: vector size is required to create collection %q", s.collection)
	}

	return s.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: s.collection,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     s.vectorSize,
			Distance: qdrant.Distance_Cosine,
		}),
	})
}

func pointIDString(id *qdrant.PointId) string {
	if id == nil {
		return ""
	}
	if uuid := id.GetUuid(); uuid != "" {
		return uuid
	}
	return fmt.Sprintf("%d", id.GetNum())
}

func payloadToMap(payload map[string]*qdrant.Value) map[string]any {
	if len(payload) == 0 {
		return nil
	}
	out := make(map[string]any, len(payload))
	for key, value := range payload {
		if value == nil {
			continue
		}
		if text := value.GetStringValue(); text != "" {
			out[key] = text
			continue
		}
		if value.GetIntegerValue() != 0 {
			out[key] = value.GetIntegerValue()
		}
	}
	return out
}
