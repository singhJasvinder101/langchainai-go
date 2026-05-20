package main

import (
	"context"
	"fmt"
	"log"

	"github.com/singhJasvinder101/ai-go/internal/init/config"
	"github.com/singhJasvinder101/ai-go/internal/llm"
	"github.com/singhJasvinder101/ai-go/internal/llm/providers"
	"github.com/singhJasvinder101/ai-go/internal/llm/providers/gemini"
	openaiprovider "github.com/singhJasvinder101/ai-go/internal/llm/providers/openai"
)

func main() {
	//example to test
	config.MustInit(config.DefaultConfigPath)

	ctx := context.Background()

	providerName := config.GetString("llm.provider")
	if providerName == "" {
		providerName = "gemini"
	}

	providerFactory, err := providers.NewProviderFactory(ctx)
	if err != nil {
		log.Fatal("failed to create provider factory: %w", err)
		panic(err)
	}

	provider, err := providerFactory.GetProviderByName(providerName)
	if err != nil {
		log.Fatal("failed to create provider: %w", err)
		panic(err)
	}

	var req llm.RequestInterface
	switch providerName {
	case "openai":
		req = &openaiprovider.GenerateRequest{Prompt: "Why is the sky blue?"}
	default:
		req = &gemini.GenerateRequest{Prompt: "Why is the sky blue?"}
	}

	//response, err := provider.Generate(ctx, &req)
	//if err != nil {
	//	log.Fatal("failed to generate response: %w", err)
	//	panic(err)
	//}

	//fmt.Println(response.GetText())

	responses, errs := provider.GenerateStream(ctx, req)
	for response := range responses {
		fmt.Print(response.GetText())
	}
	for err := range errs {
		if err != nil {
			log.Fatal("failed to generate stream response: %w", err)
		}
	}
	fmt.Println()
}
