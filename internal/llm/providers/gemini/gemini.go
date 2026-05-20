package gemini

import (
	"context"
	"fmt"
	"log"

	"github.com/singhJasvinder101/cursor-go/internal/init/config"
	"github.com/singhJasvinder101/cursor-go/internal/llm"
	"google.golang.org/genai"
)

type GeminiProvider struct {
	Client *genai.Client
}


func NewGeminiProvider(ctx context.Context) *GeminiProvider {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: config.GetString("gemini.api_key"),
	})
	if err != nil {
		log.Fatal(err)
		panic(err)
	}

	return &GeminiProvider{Client: client}
}

func (p *GeminiProvider) Generate(ctx context.Context, req llm.RequestInterface) (llm.ResponseInterface, error) {
	geminiReq, ok := req.(*GenerateRequest)
	if !ok || geminiReq == nil {
		return nil, fmt.Errorf("gemini: request must be a non-nil *gemini.GenerateRequest")
	}
	if err := geminiReq.Validate(); err != nil {
		return nil, err
	}

	model := config.GetString("gemini.model")
	contents := []Content{{Role: "user", Parts: []Part{{Text: geminiReq.Prompt}}}}
	response, err := p.Client.Models.GenerateContent(ctx, model, contents, nil)
	if err != nil {
		return nil, err
	}
	return &GenerateResponse{response}, nil
}
