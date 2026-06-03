package ollama

import (
	"encoding/json"
	"fmt"

	"github.com/ollama/ollama/api"
	"github.com/singhJasvinder101/agentic-go/llm"
)

func toOllamaTools(tools []llm.Tool) (api.Tools, error) {
	out := make(api.Tools, 0, len(tools))
	for i, tool := range tools {
		apiTool, err := toOllamaTool(tool)
		if err != nil {
			return nil, fmt.Errorf("tool at index %d: %w", i, err)
		}
		out = append(out, apiTool)
	}
	return out, nil
}

func toOllamaTool(tool llm.Tool) (api.Tool, error) {
	fn := api.ToolFunction{
		Name:        tool.Name,
		Description: tool.Description,
		Parameters: api.ToolFunctionParameters{
			Type: "object",
		},
	}
	if len(tool.Parameters) > 0 {
		if err := json.Unmarshal(tool.Parameters, &fn.Parameters); err != nil {
			return api.Tool{}, fmt.Errorf("parameters: %w", err)
		}
	}
	return api.Tool{
		Type:     "function",
		Function: fn,
	}, nil
}
