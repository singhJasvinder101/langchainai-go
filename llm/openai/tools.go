package openai

import (
	"fmt"

	openai "github.com/sashabaranov/go-openai"
	"github.com/singhJasvinder101/agentic-go/llm"
)

func toOpenAITools(tools []llm.Tool) ([]openai.Tool, error) {
	out := make([]openai.Tool, 0, len(tools))
	for i, tool := range tools {
		apiTool, err := toOpenAITool(tool)
		if err != nil {
			return nil, fmt.Errorf("tool at index %d: %w", i, err)
		}
		out = append(out, apiTool)
	}
	return out, nil
}

func toOpenAITool(tool llm.Tool) (openai.Tool, error) {
	params := any(map[string]any{})
	if len(tool.Parameters) > 0 {
		params = tool.Parameters
	}
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  params,
		},
	}, nil
}

func toOpenAIToolChoice(choice *llm.ToolChoice) any {
	if choice == nil {
		return nil
	}
	switch choice.Mode {
	case llm.ToolChoiceNone:
		return "none"
	case llm.ToolChoiceRequired:
		if choice.Name != "" {
			return openai.ToolChoice{
				Type: openai.ToolTypeFunction,
				Function: openai.ToolFunction{
					Name: choice.Name,
				},
			}
		}
		return "required"
	default:
		return "auto"
	}
}
