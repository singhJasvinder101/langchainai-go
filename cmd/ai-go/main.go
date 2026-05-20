package main

import (
	"context"
	"fmt"
	"log"

	"github.com/singhJasvinder101/ai-go/internal/init/config"
	"github.com/singhJasvinder101/ai-go/internal/llm/providers"
	"github.com/singhJasvinder101/ai-go/internal/llm/providers/gemini"
)

func main() {
	config.MustInit(config.DefaultConfigPath)

	ctx := context.Background()

	fmt.Println(config.GetString("gemini.model"))

	providerFactory, err := providers.NewProviderFactory(ctx)
	if err != nil {
		log.Fatal("failed to create provider factory: %w", err)
		panic(err)
	}

	provider, err := providerFactory.GetProviderByName("gemini")
	if err != nil {
		log.Fatal("failed to create provider: %w", err)
		panic(err)
	}

	req := gemini.GenerateRequest{Prompt: "Why is the sky blue?"}

	//response, err := provider.Generate(ctx, &req)
	//if err != nil {
	//	log.Fatal("failed to generate response: %w", err)
	//	panic(err)
	//}

	//fmt.Println(response.GetText())

	responses, errs := provider.GenerateStream(ctx, &req)
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
