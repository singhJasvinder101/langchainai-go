package providers

import (
	"context"
	"fmt"
	"sync"

	"github.com/singhJasvinder101/langchainai-go/constants"
	"github.com/singhJasvinder101/langchainai-go/llm"
	"github.com/singhJasvinder101/langchainai-go/llm/providers/claude"
	"github.com/singhJasvinder101/langchainai-go/llm/providers/gemini"
	"github.com/singhJasvinder101/langchainai-go/llm/providers/ollama"
	"github.com/singhJasvinder101/langchainai-go/llm/providers/openai"
)

type providerRegistration struct {
	constructor        func(ctx context.Context) llm.Provider
	supportsEmbeddings bool
}

type ProviderFactory struct {
	providers          map[string]llm.Provider
	embeddingProviders map[string]llm.EmbeddingProvider
}

var (
	providerFactory *ProviderFactory
	once            sync.Once
)

var providerRegistry = map[string]providerRegistration{
	constants.ProviderGemini: {constructor: initGeminiProvider, supportsEmbeddings: true},
	constants.ProviderOpenAI: {constructor: initOpenAIProvider, supportsEmbeddings: true},
	constants.ProviderClaude: {constructor: initClaudeProvider, supportsEmbeddings: false},
	constants.ProviderOllama: {constructor: initOllamaProvider, supportsEmbeddings: true},
}

func NewProviderFactory(ctx context.Context) *ProviderFactory {
	once.Do(func() {
		providerFactory = initializeProviderFactory(ctx)
	})

	return providerFactory
}

func initializeProviderFactory(ctx context.Context) *ProviderFactory {
	factory := &ProviderFactory{
		providers:          make(map[string]llm.Provider),
		embeddingProviders: make(map[string]llm.EmbeddingProvider),
	}

	for providerType, registration := range providerRegistry {
		provider := registration.constructor(ctx)
		factory.providers[providerType] = provider

		if registration.supportsEmbeddings {
			embeddingProvider, ok := provider.(llm.EmbeddingProvider)
			if !ok {
				panic(fmt.Sprintf("provider %q is marked embedding-capable but does not implement llm.EmbeddingProvider", providerType))
			}
			factory.embeddingProviders[providerType] = embeddingProvider
		}
	}

	providerFactory = factory
	return factory
}

func GetEmbeddingProviderByName(name string) (llm.EmbeddingProvider, error) {
	if providerFactory == nil {
		return nil, fmt.Errorf("provider factory is not initialized")
	}

	embeddingProvider, ok := providerFactory.embeddingProviders[name]
	if ok {
		return embeddingProvider, nil
	}

	if _, exists := providerFactory.providers[name]; exists {
		return nil, fmt.Errorf("provider does not support embeddings: %s", name)
	}

	return nil, fmt.Errorf("embedding provider not found: %s", name)
}

func GetProviderByName(name string) (llm.Provider, error) {
	if providerFactory == nil {
		return nil, fmt.Errorf("provider factory is not initialized")
	}

	provider, ok := providerFactory.providers[name]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	return provider, nil
}

func initGeminiProvider(ctx context.Context) llm.Provider {
	return gemini.NewGeminiProvider(ctx)
}

func initOpenAIProvider(ctx context.Context) llm.Provider {
	return openai.NewOpenAIProvider(ctx)
}

func initClaudeProvider(ctx context.Context) llm.Provider {
	return claude.NewClaudeProvider(ctx)
}

func initOllamaProvider(ctx context.Context) llm.Provider {
	return ollama.NewOllamaProvider(ctx)
}
