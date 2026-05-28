package aigo

import (
	"context"

	initializers "github.com/singhJasvinder101/ai-go/init"
	"github.com/singhJasvinder101/ai-go/internal/llm"
	"github.com/singhJasvinder101/ai-go/internal/llm/providers"
)

func New(ctx context.Context, providerName string, configSrc string) (llm.Provider, error) {
	initializers.Init(ctx, configSrc)

	provider, err := providers.GetProviderByName(providerName)
	if err != nil {
		return nil, err
	}

	return provider, nil
}
