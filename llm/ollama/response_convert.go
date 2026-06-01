package ollama

import (
	"github.com/ollama/ollama/api"
	"github.com/singhJasvinder101/agentic-go/llm"
)

func generateResponseFromChat(fullText string, model string, resp api.ChatResponse) *llm.GenerateResponse {
	reason := llm.FinishReasonStop
	if resp.DoneReason != "" {
		reason = llm.FinishReason(resp.DoneReason)
	}

	var usage *llm.UsageMetadata
	if resp.PromptEvalCount > 0 || resp.EvalCount > 0 {
		usage = &llm.UsageMetadata{
			PromptTokens:     resp.PromptEvalCount,
			CompletionTokens: resp.EvalCount,
			TotalTokens:      resp.PromptEvalCount + resp.EvalCount,
		}
	}

	return &llm.GenerateResponse{
		Model: model,
		Choices: []llm.Choice{
			llm.NewChoice(0, partsFromOllamaMessage(fullText, resp.Message), reason),
		},
		Usage: usage,
		Raw:   &resp,
	}
}

func streamResponseFromChat(resp api.ChatResponse, model string) *llm.StreamResponse {
	parts := partsFromStreamContent(resp.Message.Content)
	if len(parts) == 0 {
		return nil
	}
	reason := llm.FinishReason("")
	if resp.Done {
		reason = llm.FinishReasonStop
	}
	return &llm.StreamResponse{
		Model: model,
		Choices: []llm.StreamChoice{
			llm.NewStreamChoice(0, parts, reason),
		},
		Raw: &resp,
	}
}
