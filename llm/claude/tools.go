package claude

import (
	"encoding/json"
	"fmt"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/singhJasvinder101/agentic-go/llm"
)

func toClaudeTools(tools []llm.Tool) ([]anthropic.ToolUnionParam, error) {
	out := make([]anthropic.ToolUnionParam, 0, len(tools))
	for i, tool := range tools {
		apiTool, err := toClaudeTool(tool)
		if err != nil {
			return nil, fmt.Errorf("tool at index %d: %w", i, err)
		}
		out = append(out, apiTool)
	}
	return out, nil
}

func toClaudeTool(tool llm.Tool) (anthropic.ToolUnionParam, error) {
	schema := anthropic.ToolInputSchemaParam{}
	if len(tool.Parameters) > 0 {
		var raw map[string]any
		if err := json.Unmarshal(tool.Parameters, &raw); err != nil {
			return anthropic.ToolUnionParam{}, fmt.Errorf("parameters: %w", err)
		}
		if props, ok := raw["properties"]; ok {
			schema.Properties = props
		}
		if req, ok := raw["required"].([]any); ok {
			for _, item := range req {
				if name, ok := item.(string); ok {
					schema.Required = append(schema.Required, name)
				}
			}
		}
	}

	param := anthropic.ToolParam{
		Name:        tool.Name,
		InputSchema: schema,
	}
	if tool.Description != "" {
		param.Description = anthropic.String(tool.Description)
	}
	return anthropic.ToolUnionParam{OfTool: &param}, nil
}

func toClaudeToolChoice(choice *llm.ToolChoice) anthropic.ToolChoiceUnionParam {
	if choice == nil {
		return anthropic.ToolChoiceUnionParam{}
	}
	switch choice.Mode {
	case llm.ToolChoiceNone:
		return anthropic.ToolChoiceUnionParam{
			OfNone: &anthropic.ToolChoiceNoneParam{},
		}
	case llm.ToolChoiceRequired:
		if choice.Name != "" {
			return anthropic.ToolChoiceUnionParam{
				OfTool: &anthropic.ToolChoiceToolParam{Name: choice.Name},
			}
		}
		return anthropic.ToolChoiceUnionParam{
			OfAny: &anthropic.ToolChoiceAnyParam{},
		}
	default:
		return anthropic.ToolChoiceUnionParam{
			OfAuto: &anthropic.ToolChoiceAutoParam{},
		}
	}
}
