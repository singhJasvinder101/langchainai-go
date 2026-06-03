package claude

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/singhJasvinder101/agentic-go/llm"
)

type claudeMessages struct {
	system   []anthropic.TextBlockParam
	messages []anthropic.MessageParam
}

func toClaudeMessages(messages []llm.Message) (claudeMessages, error) {
	var systemParts []string
	chat := make([]anthropic.MessageParam, 0, len(messages))

	for i, msg := range messages {
		switch msg.Role {
		case llm.RoleSystem:
			text, err := llm.JoinTextParts(msg.Parts)
			if err != nil {
				return claudeMessages{}, fmt.Errorf("message at index %d: %w", i, err)
			}
			systemParts = append(systemParts, text)
		case llm.RoleTool:
			blocks, err := toClaudeBlocks(msg.Parts)
			if err != nil {
				return claudeMessages{}, fmt.Errorf("message at index %d: %w", i, err)
			}
			chat = append(chat, anthropic.NewUserMessage(blocks...))
		case llm.RoleUser, llm.RoleAssistant:
			blocks, err := toClaudeBlocks(msg.Parts)
			if err != nil {
				return claudeMessages{}, fmt.Errorf("message at index %d: %w", i, err)
			}
			if msg.Role == llm.RoleUser {
				chat = append(chat, anthropic.NewUserMessage(blocks...))
			} else {
				chat = append(chat, anthropic.NewAssistantMessage(blocks...))
			}
		default:
			return claudeMessages{}, fmt.Errorf("message at index %d: unsupported role %q", i, msg.Role)
		}
	}

	if len(chat) == 0 {
		return claudeMessages{}, fmt.Errorf("at least one user or assistant message is required")
	}

	result := claudeMessages{messages: chat}
	if len(systemParts) > 0 {
		result.system = []anthropic.TextBlockParam{
			{Text: strings.Join(systemParts, "\n\n")},
		}
	}
	return result, nil
}

func toClaudeBlocks(parts []llm.ContentPart) ([]anthropic.ContentBlockParamUnion, error) {
	blocks := make([]anthropic.ContentBlockParamUnion, 0, len(parts))
	for i, part := range parts {
		block, err := toClaudeBlock(part)
		if err != nil {
			return nil, fmt.Errorf("part at index %d: %w", i, err)
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}

func toClaudeBlock(part llm.ContentPart) (anthropic.ContentBlockParamUnion, error) {
	switch part.Type {
	case llm.PartText:
		return anthropic.NewTextBlock(part.Text), nil
	case llm.PartImageURL:
		return anthropic.NewImageBlock(anthropic.URLImageSourceParam{URL: part.URL}), nil
	case llm.PartImage:
		return anthropic.NewImageBlockBase64(part.MIMEType, base64.StdEncoding.EncodeToString(part.Data)), nil
	case llm.PartToolCall:
		if part.ToolCall == nil {
			return anthropic.ContentBlockParamUnion{}, fmt.Errorf("tool call is required")
		}
		var input any
		if len(part.ToolCall.Arguments) > 0 {
			if err := json.Unmarshal(part.ToolCall.Arguments, &input); err != nil {
				return anthropic.ContentBlockParamUnion{}, err
			}
		} else {
			input = map[string]any{}
		}
		id := part.ToolCall.ID
		if id == "" {
			id = "toolu_" + part.ToolCall.Name
		}
		return anthropic.NewToolUseBlock(id, input, part.ToolCall.Name), nil
	case llm.PartToolResult:
		return anthropic.NewToolResultBlock(part.ToolCallID, part.Text, false), nil
	case llm.PartFile:
		switch part.MIMEType {
		case "application/pdf":
			return anthropic.NewDocumentBlock(anthropic.Base64PDFSourceParam{
				Type:      "base64",
				MediaType: "application/pdf",
				Data:      base64.StdEncoding.EncodeToString(part.Data),
			}), nil
		case "text/plain":
			return anthropic.NewDocumentBlock(anthropic.PlainTextSourceParam{
				Type: "text",
				Data: string(part.Data),
			}), nil
		default:
			return anthropic.ContentBlockParamUnion{}, fmt.Errorf("unsupported file mime type %q", part.MIMEType)
		}
	default:
		return anthropic.ContentBlockParamUnion{}, fmt.Errorf("unsupported part type %q", part.Type)
	}
}
