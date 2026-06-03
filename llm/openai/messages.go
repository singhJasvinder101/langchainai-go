package openai

import (
	"encoding/base64"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"github.com/singhJasvinder101/agentic-go/llm"
)

func toChatCompletionMessages(messages []llm.Message) ([]openai.ChatCompletionMessage, error) {
	out := make([]openai.ChatCompletionMessage, 0, len(messages))
	for i, msg := range messages {
		apiMsg, err := toChatCompletionMessage(msg)
		if err != nil {
			return nil, fmt.Errorf("message at index %d: %w", i, err)
		}
		out = append(out, apiMsg)
	}
	return out, nil
}

func toChatCompletionMessage(msg llm.Message) (openai.ChatCompletionMessage, error) {
	switch msg.Role {
	case llm.RoleTool:
		return toToolResultMessage(msg)
	case llm.RoleAssistant:
		return toAssistantMessage(msg)
	default:
		role, err := toOpenAIRole(msg.Role)
		if err != nil {
			return openai.ChatCompletionMessage{}, err
		}
		return toUserOrSystemMessage(role, msg.Parts)
	}
}

func toToolResultMessage(msg llm.Message) (openai.ChatCompletionMessage, error) {
	for _, part := range msg.Parts {
		if part.Type != llm.PartToolResult {
			continue
		}
		return openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Content:    part.Text,
			Name:       part.Name,
			ToolCallID: part.ToolCallID,
		}, nil
	}
	return openai.ChatCompletionMessage{}, fmt.Errorf("tool message requires a tool_result part")
}

func toAssistantMessage(msg llm.Message) (openai.ChatCompletionMessage, error) {
	calls := llm.ToolCalls(msg.Parts)
	text := llm.TextFromParts(msg.Parts)

	if len(calls) > 0 {
		apiCalls := make([]openai.ToolCall, 0, len(calls))
		for _, call := range calls {
			apiCalls = append(apiCalls, openai.ToolCall{
				ID:   call.ID,
				Type: openai.ToolTypeFunction,
				Function: openai.FunctionCall{
					Name:      call.Name,
					Arguments: call.ArgumentsString(),
				},
			})
		}
		return openai.ChatCompletionMessage{
			Role:      openai.ChatMessageRoleAssistant,
			Content:   text,
			ToolCalls: apiCalls,
		}, nil
	}

	return toUserOrSystemMessage(openai.ChatMessageRoleAssistant, msg.Parts)
}

func toUserOrSystemMessage(role string, parts []llm.ContentPart) (openai.ChatCompletionMessage, error) {
	if llm.IsTextOnly(parts) {
		text, err := llm.JoinTextParts(parts)
		if err != nil {
			return openai.ChatCompletionMessage{}, err
		}
		return openai.ChatCompletionMessage{Role: role, Content: text}, nil
	}

	multi := make([]openai.ChatMessagePart, 0, len(parts))
	for i, part := range parts {
		if part.Type == llm.PartToolCall || part.Type == llm.PartToolResult {
			return openai.ChatCompletionMessage{}, fmt.Errorf("part at index %d: tool parts are not valid for role %q", i, role)
		}
		apiPart, err := toChatMessagePart(part)
		if err != nil {
			return openai.ChatCompletionMessage{}, fmt.Errorf("part at index %d: %w", i, err)
		}
		multi = append(multi, apiPart)
	}
	return openai.ChatCompletionMessage{Role: role, MultiContent: multi}, nil
}

func toChatMessagePart(part llm.ContentPart) (openai.ChatMessagePart, error) {
	switch part.Type {
	case llm.PartText:
		return openai.ChatMessagePart{
			Type: openai.ChatMessagePartTypeText,
			Text: part.Text,
		}, nil
	case llm.PartImageURL:
		return openai.ChatMessagePart{
			Type: openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{
				URL: part.URL,
			},
		}, nil
	case llm.PartImage:
		return openai.ChatMessagePart{
			Type: openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{
				URL: dataURI(part.MIMEType, part.Data),
			},
		}, nil
	case llm.PartFile:
		return openai.ChatMessagePart{}, fmt.Errorf("file parts are not supported by openai chat completions")
	default:
		return openai.ChatMessagePart{}, fmt.Errorf("unsupported part type %q", part.Type)
	}
}

func dataURI(mimeType string, data []byte) string {
	return fmt.Sprintf("data:%s;base64,%s", mimeType, base64.StdEncoding.EncodeToString(data))
}

func toOpenAIRole(role llm.Role) (string, error) {
	switch role {
	case llm.RoleSystem:
		return openai.ChatMessageRoleSystem, nil
	case llm.RoleUser:
		return openai.ChatMessageRoleUser, nil
	case llm.RoleAssistant:
		return openai.ChatMessageRoleAssistant, nil
	default:
		return "", fmt.Errorf("unsupported role %q", role)
	}
}

func finishReasonFromOpenAI(reason openai.FinishReason) llm.FinishReason {
	if reason == openai.FinishReasonToolCalls {
		return llm.FinishReasonToolCalls
	}
	return llm.FinishReason(reason)
}

func toolCallPartsFromMessage(msg openai.ChatCompletionMessage) []llm.ContentPart {
	if len(msg.ToolCalls) == 0 {
		return nil
	}
	parts := make([]llm.ContentPart, 0, len(msg.ToolCalls))
	for _, call := range msg.ToolCalls {
		parts = append(parts, llm.ToolCallPart(llm.ToolCall{
			ID:        call.ID,
			Name:      call.Function.Name,
			Arguments: []byte(call.Function.Arguments),
		}))
	}
	return parts
}

func partsFromChatMessage(msg openai.ChatCompletionMessage) []llm.ContentPart {
	parts := toolCallPartsFromMessage(msg)
	if msg.Content != "" {
		parts = append([]llm.ContentPart{llm.TextPart(msg.Content)}, parts...)
	} else if len(msg.MultiContent) > 0 {
		for _, block := range msg.MultiContent {
			part, err := partFromChatMessagePart(block)
			if err == nil {
				parts = append(parts, part)
			}
		}
	}
	return parts
}

func partFromChatMessagePart(block openai.ChatMessagePart) (llm.ContentPart, error) {
	switch block.Type {
	case openai.ChatMessagePartTypeText:
		return llm.TextPart(block.Text), nil
	case openai.ChatMessagePartTypeImageURL:
		if block.ImageURL == nil {
			return llm.ContentPart{}, fmt.Errorf("image_url block missing url")
		}
		return llm.ImageURLPart(block.ImageURL.URL), nil
	default:
		return llm.ContentPart{}, fmt.Errorf("unsupported block type %q", block.Type)
	}
}

func partsFromStreamDelta(delta openai.ChatCompletionStreamChoiceDelta) []llm.ContentPart {
	var parts []llm.ContentPart
	if delta.Content != "" {
		parts = append(parts, llm.TextPart(delta.Content))
	}
	for _, call := range delta.ToolCalls {
		if call.Function.Name == "" && call.Function.Arguments == "" {
			continue
		}
		parts = append(parts, llm.ToolCallPart(llm.ToolCall{
			ID:        call.ID,
			Name:      call.Function.Name,
			Arguments: []byte(strings.TrimSpace(call.Function.Arguments)),
		}))
	}
	return parts
}
