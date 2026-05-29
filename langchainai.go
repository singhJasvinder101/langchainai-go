package langchainaiGo

import (
	"context"

	"github.com/singhJasvinder101/langchainai-go/llm"
	"github.com/singhJasvinder101/langchainai-go/llm/providers"
)

func New(ctx context.Context, providerName string) (llm.Provider, error) {
	provider, err := providers.GetProviderByName(providerName)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

func NewEmbeddings(ctx context.Context, providerName string) (llm.EmbeddingProvider, error) {
	provider, err := providers.GetEmbeddingProviderByName(providerName)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
