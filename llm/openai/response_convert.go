package openai

import (
	openai "github.com/sashabaranov/go-openai"
	"github.com/singhJasvinder101/agentic-go/llm"
)

func generateResponseFromChatCompletion(response *openai.ChatCompletionResponse) *llm.GenerateResponse {
	if response == nil {
		return &llm.GenerateResponse{}
	}

	choices := make([]llm.Choice, 0, len(response.Choices))
	for _, c := range response.Choices {
		choices = append(choices, llm.NewChoice(
			c.Index,
			partsFromChatMessage(c.Message),
			finishReasonFromOpenAI(c.FinishReason),
		))
	}

	var usage *llm.UsageMetadata
	if response.Usage.TotalTokens > 0 || response.Usage.PromptTokens > 0 {
		usage = &llm.UsageMetadata{
			PromptTokens:     response.Usage.PromptTokens,
			CompletionTokens: response.Usage.CompletionTokens,
			TotalTokens:      response.Usage.TotalTokens,
		}
	}

	return &llm.GenerateResponse{
		ID:      response.ID,
		Model:   response.Model,
		Object:  string(response.Object),
		Created: response.Created,
		Choices: choices,
		Usage:   usage,
		Raw:     response,
	}
}

func streamResponseFromChunk(response *openai.ChatCompletionStreamResponse) *llm.StreamResponse {
	if response == nil {
		return &llm.StreamResponse{}
	}

	choices := make([]llm.StreamChoice, 0, len(response.Choices))
	for _, c := range response.Choices {
		choices = append(choices, llm.NewStreamChoice(
			c.Index,
			partsFromStreamDelta(c.Delta),
			finishReasonFromOpenAI(c.FinishReason),
		))
	}

	return &llm.StreamResponse{
		ID:      response.ID,
		Model:   response.Model,
		Choices: choices,
		Raw:     response,
	}
}
