package storeutil

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/singhJasvinder101/langchainai-go/embedder"
	"github.com/singhJasvinder101/langchainai-go/vectorstore"
)

const ContentPayloadKey = "content"

type PreparedDocuments struct {
	Documents []vectorstore.Document
	Vectors   [][]float32
}

func PrepareDocuments(ctx context.Context, e embedder.Embedder, docs []vectorstore.Document) (*PreparedDocuments, error) {
	if len(docs) == 0 {
		return nil, vectorstore.ErrDocumentsRequired
	}
	if e == nil {
		return nil, fmt.Errorf("storeutil: embedder is required")
	}

	prepared := make([]vectorstore.Document, len(docs))
	texts := make([]string, len(docs))
	for i, doc := range docs {
		if doc.Content == "" {
			return nil, fmt.Errorf("%w at index %d", vectorstore.ErrDocumentContent, i)
		}
		id := doc.ID
		if id == "" {
			id = uuid.NewString()
		}
		prepared[i] = vectorstore.Document{
			ID:       id,
			Content:  doc.Content,
			Metadata: doc.Metadata,
		}
		texts[i] = doc.Content
	}

	vectors, err := e.EmbedDocuments(ctx, texts)
	if err != nil {
		return nil, err
	}
	if len(vectors) != len(prepared) {
		return nil, fmt.Errorf("storeutil: expected %d embeddings, got %d", len(prepared), len(vectors))
	}

	return &PreparedDocuments{
		Documents: prepared,
		Vectors:   vectors,
	}, nil
}

func Payload(doc vectorstore.Document) map[string]any {
	payload := map[string]any{ContentPayloadKey: doc.Content}
	for key, value := range doc.Metadata {
		payload[key] = value
	}
	return payload
}

func DocumentFromPayload(id string, payload map[string]any) vectorstore.Document {
	doc := vectorstore.Document{
		ID:       id,
		Metadata: make(map[string]string),
	}
	if payload == nil {
		return doc
	}
	if content, ok := payload[ContentPayloadKey].(string); ok {
		doc.Content = content
	}
	for key, value := range payload {
		if key == ContentPayloadKey {
			continue
		}
		if text, ok := value.(string); ok {
			doc.Metadata[key] = text
		}
	}
	return doc
}

func DistanceToScore(distance float32) float32 {
	if distance <= 0 {
		return 1
	}
	return 1 / (1 + distance)
}
