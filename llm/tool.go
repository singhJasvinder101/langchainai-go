package llm

import (
	"encoding/json"
	"errors"
	"fmt"
)

// Tool describes a function the model may call.
type Tool struct {
	Name        string
	Description string
	// Parameters is a JSON Schema object describing the function inputs.
	Parameters json.RawMessage
}

// NewTool builds a tool definition. parameters must be a JSON object schema (or nil for empty object).
func NewTool(name, description string, parameters json.RawMessage) Tool {
	return Tool{Name: name, Description: description, Parameters: parameters}
}

// ToolCall is a model-requested invocation of a tool.
type ToolCall struct {
	ID        string
	Name      string
	Arguments json.RawMessage
}

// ToolChoiceMode controls how the model uses tools.
type ToolChoiceMode string

const (
	ToolChoiceAuto     ToolChoiceMode = "auto"
	ToolChoiceNone     ToolChoiceMode = "none"
	ToolChoiceRequired ToolChoiceMode = "required"
)

// ToolChoice configures tool use for a request.
type ToolChoice struct {
	Mode ToolChoiceMode
	// Name forces a specific tool when Mode is ToolChoiceRequired.
	Name string
}

func (t Tool) Validate() error {
	if t.Name == "" {
		return errors.New("tool name is required")
	}
	if len(t.Parameters) > 0 && !json.Valid(t.Parameters) {
		return errors.New("tool parameters must be valid JSON")
	}
	return nil
}

func (c ToolCall) Validate() error {
	if c.Name == "" {
		return errors.New("tool call name is required")
	}
	if len(c.Arguments) > 0 && !json.Valid(c.Arguments) {
		return errors.New("tool call arguments must be valid JSON")
	}
	return nil
}

// ArgumentsString returns arguments as a string for providers that expect JSON text.
func (c ToolCall) ArgumentsString() string {
	if len(c.Arguments) == 0 {
		return "{}"
	}
	return string(c.Arguments)
}

// ParseArguments unmarshals arguments into dest.
func (c ToolCall) ParseArguments(dest any) error {
	if len(c.Arguments) == 0 {
		return nil
	}
	return json.Unmarshal(c.Arguments, dest)
}

// ToolCalls extracts tool call parts from a message.
func ToolCalls(parts []ContentPart) []ToolCall {
	calls := make([]ToolCall, 0)
	for _, part := range parts {
		if part.Type == PartToolCall && part.ToolCall != nil {
			calls = append(calls, *part.ToolCall)
		}
	}
	return calls
}

// ValidateTools validates tool definitions when provided.
func ValidateTools(tools []Tool) error {
	for i, tool := range tools {
		if err := tool.Validate(); err != nil {
			return fmt.Errorf("tool at index %d: %w", i, err)
		}
	}
	return nil
}
