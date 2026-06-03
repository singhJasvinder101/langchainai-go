package gemini

import (
	"encoding/json"
	"fmt"

	"github.com/singhJasvinder101/agentic-go/llm"
	"google.golang.org/genai"
)

func toGeminiTools(tools []llm.Tool) ([]*genai.Tool, error) {
	decls := make([]*genai.FunctionDeclaration, 0, len(tools))
	for i, tool := range tools {
		decl, err := toGeminiFunctionDeclaration(tool)
		if err != nil {
			return nil, fmt.Errorf("tool at index %d: %w", i, err)
		}
		decls = append(decls, decl)
	}
	if len(decls) == 0 {
		return nil, nil
	}
	return []*genai.Tool{{FunctionDeclarations: decls}}, nil
}

func toGeminiFunctionDeclaration(tool llm.Tool) (*genai.FunctionDeclaration, error) {
	decl := &genai.FunctionDeclaration{
		Name:        tool.Name,
		Description: tool.Description,
	}
	if len(tool.Parameters) > 0 {
		var schema any
		if err := json.Unmarshal(tool.Parameters, &schema); err != nil {
			return nil, fmt.Errorf("parameters: %w", err)
		}
		decl.ParametersJsonSchema = schema
	}
	return decl, nil
}

func toGeminiToolConfig(choice *llm.ToolChoice) *genai.ToolConfig {
	if choice == nil {
		return nil
	}
	cfg := &genai.FunctionCallingConfig{}
	switch choice.Mode {
	case llm.ToolChoiceNone:
		cfg.Mode = genai.FunctionCallingConfigModeNone
	case llm.ToolChoiceRequired:
		cfg.Mode = genai.FunctionCallingConfigModeAny
		if choice.Name != "" {
			cfg.AllowedFunctionNames = []string{choice.Name}
		}
	default:
		cfg.Mode = genai.FunctionCallingConfigModeAuto
	}
	return &genai.ToolConfig{FunctionCallingConfig: cfg}
}
