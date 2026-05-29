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

type ProviderFactory struct {
	providers map[string]llm.Provider
}

var (
	providerFactory *ProviderFactory
	once            sync.Once
)

var providerConstructorMap = map[string]func(ctx context.Context) llm.Provider{
	constants.ProviderGemini: initGeminiProvider,
	constants.ProviderOpenAI: initOpenAIProvider,
	constants.ProviderClaude: initClaudeProvider,
	constants.ProviderOllama: initOllamaProvider,
}

func NewProviderFactory(ctx context.Context) *ProviderFactory {
	once.Do(func() {
		providerFactory = initializeProviderFactory(ctx)
	})

	return providerFactory
}

func initializeProviderFactory(ctx context.Context) *ProviderFactory {
	factory := &ProviderFactory{
		providers: make(map[string]llm.Provider),
	}

	for providerType, constructor := range providerConstructorMap {
		factory.providers[providerType] = constructor(ctx)
	}

	providerFactory = factory
	return factory
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

func GetEmbeddingProviderByName(name string) (llm.EmbeddingProvider, error) {
	provider, err := GetProviderByName(name)
	if err != nil {
		return nil, err
	}

	embeddingProvider, ok := provider.(llm.EmbeddingProvider)
	if !ok {
		return nil, fmt.Errorf("provider does not support embeddings: %s", name)
	}
	return embeddingProvider, nil
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
