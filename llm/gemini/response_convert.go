package gemini

import (
	"github.com/singhJasvinder101/agentic-go/llm"
	"google.golang.org/genai"
)

func generateResponseFromGenerateContent(response *genai.GenerateContentResponse, model string) *llm.GenerateResponse {
	if response == nil {
		return &llm.GenerateResponse{Model: model}
	}

	choices := make([]llm.Choice, 0, len(response.Candidates))
	for i, candidate := range response.Candidates {
		var parts []llm.ContentPart
		if candidate.Content != nil {
			parts = partsFromGenaiParts(candidate.Content.Parts)
		}
		choices = append(choices, llm.NewChoice(i, parts, finishReasonFromCandidate(candidate)))
	}

	var usage *llm.UsageMetadata
	if response.UsageMetadata != nil {
		usage = &llm.UsageMetadata{
			PromptTokens:     int(response.UsageMetadata.PromptTokenCount),
			CompletionTokens: int(response.UsageMetadata.CandidatesTokenCount),
			TotalTokens:      int(response.UsageMetadata.TotalTokenCount),
		}
	}

	return &llm.GenerateResponse{
		Model:   model,
		Choices: choices,
		Usage:   usage,
		Raw:     response,
	}
}

func streamResponseFromGenerateContent(response *genai.GenerateContentResponse, model string) *llm.StreamResponse {
	if response == nil {
		return &llm.StreamResponse{Model: model}
	}

	choices := make([]llm.StreamChoice, 0, len(response.Candidates))
	for i, candidate := range response.Candidates {
		var parts []llm.ContentPart
		if candidate.Content != nil {
			parts = partsFromGenaiParts(candidate.Content.Parts)
		}
		if len(parts) == 0 {
			continue
		}
		choices = append(choices, llm.NewStreamChoice(i, parts, finishReasonFromCandidate(candidate)))
	}

	return &llm.StreamResponse{
		Model:   model,
		Choices: choices,
		Raw:     response,
	}
}

func finishReasonFromCandidate(candidate *genai.Candidate) llm.FinishReason {
	if candidate == nil {
		return ""
	}
	switch candidate.FinishReason {
	case genai.FinishReasonStop:
		return llm.FinishReasonStop
	case genai.FinishReasonMaxTokens:
		return llm.FinishReasonLength
	default:
		return llm.FinishReason(candidate.FinishReason)
	}
}
