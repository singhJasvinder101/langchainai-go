package main

import (
	"context"
	"fmt"

	"github.com/singhJasvinder101/ai-go/internal/constants"
	initializers "github.com/singhJasvinder101/ai-go/init"
	"github.com/singhJasvinder101/ai-go/init/config"
	"github.com/singhJasvinder101/ai-go/internal/llm"
	"github.com/singhJasvinder101/ai-go/internal/llm/providers"
	claudeprovider "github.com/singhJasvinder101/ai-go/internal/llm/providers/claude"
	"github.com/singhJasvinder101/ai-go/internal/llm/providers/gemini"
	ollamaprovider "github.com/singhJasvinder101/ai-go/internal/llm/providers/ollama"
	openaiprovider "github.com/singhJasvinder101/ai-go/internal/llm/providers/openai"
	"github.com/singhJasvinder101/ai-go/pkg/log"
)

func main() {
	ctx := context.Background()
	if err := initializers.Init(ctx, config.DefaultConfigPath); err != nil {
		log.Fatal("failed to initialize application", "error", err)
	}

	providerName := config.GetString(constants.ConfigLLMProvider)
	if providerName == "" {
		providerName = constants.ProviderGemini
	}

	providers.NewProviderFactory(ctx)

	provider, err := providers.GetProviderByName(providerName)
	if err != nil {
		log.Fatal("failed to get provider", "error", err, "provider", providerName)
	}

	var req llm.RequestInterface
	switch providerName {
	case constants.ProviderOpenAI:
		req = &openaiprovider.GenerateRequest{Prompt: "Why is the sky blue?"}
	case constants.ProviderClaude:
		req = &claudeprovider.GenerateRequest{Prompt: "Why is the sky blue?"}
	case constants.ProviderOllama:
		req = &ollamaprovider.GenerateRequest{Prompt: "Why is the sky blue?"}
	default:
		req = &gemini.GenerateRequest{Prompt: "Why is the sky blue?"}
	}

	responses, errs := provider.GenerateStream(ctx, req)
	for response := range responses {
		fmt.Print(response.GetText())
	}
	for err := range errs {
		if err != nil {
			log.Fatal("failed to generate stream response", "error", err, "provider", providerName)
		}
	}
	fmt.Println()
}
