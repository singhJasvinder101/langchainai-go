package claude

import (
	"github.com/anthropics/anthropic-sdk-go"
	"github.com/singhJasvinder101/agentic-go/llm"
)

func generateResponseFromMessage(message *anthropic.Message) *llm.GenerateResponse {
	if message == nil {
		return &llm.GenerateResponse{}
	}

	reason := llm.FinishReasonStop
	switch message.StopReason {
	case anthropic.StopReasonEndTurn, anthropic.StopReasonStopSequence:
		reason = llm.FinishReasonStop
	case anthropic.StopReasonMaxTokens:
		reason = llm.FinishReasonLength
	case anthropic.StopReasonToolUse:
		reason = llm.FinishReasonToolCalls
	}

	usage := &llm.UsageMetadata{
		PromptTokens:     int(message.Usage.InputTokens),
		CompletionTokens: int(message.Usage.OutputTokens),
		TotalTokens:      int(message.Usage.InputTokens + message.Usage.OutputTokens),
	}

	return &llm.GenerateResponse{
		ID:    message.ID,
		Model: string(message.Model),
		Choices: []llm.Choice{
			llm.NewChoice(0, partsFromMessage(message), reason),
		},
		Usage: usage,
		Raw:   message,
	}
}

func streamResponseFromTextDelta(text string, event *anthropic.MessageStreamEventUnion) *llm.StreamResponse {
	if text == "" {
		return nil
	}
	return &llm.StreamResponse{
		Choices: []llm.StreamChoice{
			llm.NewStreamChoice(0, []llm.ContentPart{llm.TextPart(text)}, ""),
		},
		Raw: event,
	}
}
