package gemini

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/singhJasvinder101/agentic-go/llm"
	"google.golang.org/genai"
)

type geminiMessages struct {
	contents          []*genai.Content
	systemInstruction *genai.Content
	tools             []*genai.Tool
	toolConfig        *genai.ToolConfig
}

func toGeminiMessages(messages []llm.Message, tools []llm.Tool, toolChoice *llm.ToolChoice) (geminiMessages, error) {
	var systemParts []string
	contents := make([]*genai.Content, 0, len(messages))

	for i, msg := range messages {
		switch msg.Role {
		case llm.RoleSystem:
			text, err := llm.JoinTextParts(msg.Parts)
			if err != nil {
				return geminiMessages{}, fmt.Errorf("message at index %d: %w", i, err)
			}
			systemParts = append(systemParts, text)
		case llm.RoleUser, llm.RoleAssistant, llm.RoleTool:
			content, err := toGeminiContent(msg)
			if err != nil {
				return geminiMessages{}, fmt.Errorf("message at index %d: %w", i, err)
			}
			contents = append(contents, content)
		default:
			return geminiMessages{}, fmt.Errorf("message at index %d: unsupported role %q", i, msg.Role)
		}
	}

	result := geminiMessages{contents: contents}
	if len(systemParts) > 0 {
		result.systemInstruction = genai.NewContentFromText(strings.Join(systemParts, "\n\n"), genai.RoleUser)
	}
	if len(contents) == 0 && result.systemInstruction == nil {
		return geminiMessages{}, fmt.Errorf("at least one user or assistant message is required")
	}

	if len(tools) > 0 {
		apiTools, err := toGeminiTools(tools)
		if err != nil {
			return geminiMessages{}, err
		}
		result.tools = apiTools
		result.toolConfig = toGeminiToolConfig(toolChoice)
	}
	return result, nil
}

func toGeminiContent(msg llm.Message) (*genai.Content, error) {
	parts := make([]*genai.Part, 0, len(msg.Parts))
	for i, part := range msg.Parts {
		apiPart, err := toGeminiPart(part)
		if err != nil {
			return nil, fmt.Errorf("part at index %d: %w", i, err)
		}
		parts = append(parts, apiPart)
	}

	role := genai.RoleUser
	switch msg.Role {
	case llm.RoleAssistant:
		role = genai.RoleModel
	case llm.RoleTool, llm.RoleUser:
		role = genai.RoleUser
	}
	return &genai.Content{Role: role, Parts: parts}, nil
}

func toGeminiPart(part llm.ContentPart) (*genai.Part, error) {
	switch part.Type {
	case llm.PartText:
		return genai.NewPartFromText(part.Text), nil
	case llm.PartImageURL:
		mimeType := part.MIMEType
		if mimeType == "" {
			mimeType = "image/jpeg"
		}
		return genai.NewPartFromURI(part.URL, mimeType), nil
	case llm.PartImage, llm.PartFile:
		return genai.NewPartFromBytes(part.Data, part.MIMEType), nil
	case llm.PartToolCall:
		if part.ToolCall == nil {
			return nil, fmt.Errorf("tool call is required")
		}
		var args map[string]any
		if len(part.ToolCall.Arguments) > 0 {
			if err := json.Unmarshal(part.ToolCall.Arguments, &args); err != nil {
				return nil, err
			}
		} else {
			args = map[string]any{}
		}
		return genai.NewPartFromFunctionCall(part.ToolCall.Name, args), nil
	case llm.PartToolResult:
		response := map[string]any{"result": part.Text}
		if part.Text != "" && json.Valid([]byte(part.Text)) {
			var parsed map[string]any
			if err := json.Unmarshal([]byte(part.Text), &parsed); err == nil {
				response = parsed
			}
		}
		return genai.NewPartFromFunctionResponse(part.Name, response), nil
	default:
		return nil, fmt.Errorf("unsupported part type %q", part.Type)
	}
}

func (g geminiMessages) config() *genai.GenerateContentConfig {
	if g.systemInstruction == nil && len(g.tools) == 0 && g.toolConfig == nil {
		return nil
	}
	cfg := &genai.GenerateContentConfig{}
	if g.systemInstruction != nil {
		cfg.SystemInstruction = g.systemInstruction
	}
	if len(g.tools) > 0 {
		cfg.Tools = g.tools
	}
	if g.toolConfig != nil {
		cfg.ToolConfig = g.toolConfig
	}
	return cfg
}
