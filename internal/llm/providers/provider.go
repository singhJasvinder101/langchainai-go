package providers

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/singhJasvinder101/ai-go/internal/llm"
	"github.com/singhJasvinder101/ai-go/internal/llm/providers/gemini"
)

type ProviderFactory struct {
	providers map[string]llm.Provider
}

var (
	providerFactory *ProviderFactory
	once            sync.Once
	initializeErr   error
)

var providerConstructorMap = map[string]func(ctx context.Context) llm.Provider{
	"gemini": initGeminiProvider,
}

func NewProviderFactory(ctx context.Context) (*ProviderFactory, error) {
	once.Do(func() {
		providerFactory, initializeErr = initializeProviderFactory(ctx)
		if initializeErr != nil {
			log.Fatal("failed to initialize provider factory: %w", initializeErr)
			panic(initializeErr)
		}
	})

	return providerFactory, nil
}

func initializeProviderFactory(ctx context.Context) (*ProviderFactory, error) {
	factory := &ProviderFactory{
		providers: make(map[string]llm.Provider),
	}

	for providerType, constructor := range providerConstructorMap {
		factory.providers[providerType] = constructor(ctx)
	}

	providerFactory = factory
	return factory, nil
}

func (p *ProviderFactory) GetProviderByName(name string) (llm.Provider, error) {
	provider, ok := providerFactory.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return provider, nil
}

func initGeminiProvider(ctx context.Context) llm.Provider {
	return gemini.NewGeminiProvider(ctx)
}
