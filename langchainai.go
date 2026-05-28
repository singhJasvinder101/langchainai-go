package aigo

import (
	"context"

	initializers "github.com/singhJasvinder101/langchainai-go/init"
	"github.com/singhJasvinder101/langchainai-go/internal/llm"
	"github.com/singhJasvinder101/langchainai-go/internal/llm/providers"
)

func New(ctx context.Context, providerName string, configSrc string) (llm.Provider, error) {
	initializers.Init(ctx, configSrc)

	provider, err := providers.GetProviderByName(providerName)
	if err != nil {
		return nil, err
	}

	return provider, nil
}

func NewEmbeddings(ctx context.Context, providerName string, configSrc string) (llm.EmbeddingProvider, error) {
	initializers.Init(ctx, configSrc)

	provider, err := providers.GetEmbeddingProviderByName(providerName)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
