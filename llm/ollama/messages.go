package ollama

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ollama/ollama/api"
	"github.com/singhJasvinder101/agentic-go/llm"
)

func toAPIMessages(messages []llm.Message) ([]api.Message, error) {
	out := make([]api.Message, 0, len(messages))
	for i, msg := range messages {
		apiMsg, err := toAPIMessage(msg)
		if err != nil {
			return nil, fmt.Errorf("message at index %d: %w", i, err)
		}
		out = append(out, apiMsg)
	}
	return out, nil
}

func toAPIMessage(msg llm.Message) (api.Message, error) {
	role, err := toOllamaRole(msg.Role)
	if err != nil {
		return api.Message{}, err
	}

	if msg.Role == llm.RoleTool {
		return toToolResultMessage(role, msg.Parts)
	}

	var textParts []string
	var images []api.ImageData
	var toolCalls []api.ToolCall

	for i, part := range msg.Parts {
		switch part.Type {
		case llm.PartText:
			textParts = append(textParts, part.Text)
		case llm.PartImage:
			images = append(images, api.ImageData(part.Data))
		case llm.PartToolCall:
			if part.ToolCall == nil {
				return api.Message{}, fmt.Errorf("part at index %d: tool call is required", i)
			}
			args := api.NewToolCallFunctionArguments()
			if len(part.ToolCall.Arguments) > 0 {
				var raw map[string]any
				if err := json.Unmarshal(part.ToolCall.Arguments, &raw); err != nil {
					return api.Message{}, fmt.Errorf("part at index %d: %w", i, err)
				}
				for k, v := range raw {
					args.Set(k, v)
				}
			}
			toolCalls = append(toolCalls, api.ToolCall{
				ID: part.ToolCall.ID,
				Function: api.ToolCallFunction{
					Name:      part.ToolCall.Name,
					Arguments: args,
				},
			})
		case llm.PartImageURL:
			return api.Message{}, fmt.Errorf("part at index %d: image_url is not supported by ollama; use image bytes", i)
		case llm.PartFile:
			return api.Message{}, fmt.Errorf("part at index %d: file parts are not supported by ollama", i)
		case llm.PartToolResult:
			return api.Message{}, fmt.Errorf("tool_result parts require role tool")
		default:
			return api.Message{}, fmt.Errorf("part at index %d: unsupported part type %q", i, part.Type)
		}
	}

	return api.Message{
		Role:      role,
		Content:   strings.Join(textParts, "\n"),
		Images:    images,
		ToolCalls: toolCalls,
	}, nil
}

func toToolResultMessage(role string, parts []llm.ContentPart) (api.Message, error) {
	for _, part := range parts {
		if part.Type != llm.PartToolResult {
			continue
		}
		return api.Message{
			Role:       role,
			Content:    part.Text,
			ToolName:   part.Name,
			ToolCallID: part.ToolCallID,
		}, nil
	}
	return api.Message{}, fmt.Errorf("tool message requires a tool_result part")
}

func toOllamaRole(role llm.Role) (string, error) {
	switch role {
	case llm.RoleSystem, llm.RoleUser, llm.RoleAssistant, llm.RoleTool:
		return string(role), nil
	default:
		return "", fmt.Errorf("unsupported role %q", role)
	}
}
