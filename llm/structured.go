package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/invopop/jsonschema"
)

// structuredToolName is the synthetic tool GenerateStructured forces the model
// to call with its final answer.
const structuredToolName = "emit_result"

var schemaReflector = &jsonschema.Reflector{
	// A flat, inline schema (no $defs/$ref) is what providers expect for a
	// tool's parameters.
	DoNotReference: true,
}

// SchemaFor reflects a JSON schema from T's struct tags (Pydantic-style: a
// Go struct is both the type and the schema, no hand-written JSON schema
// string needed). Use `json` tags for field names/omitempty, and
// `jsonschema_description` for a field's description. The schema is flat
// (no $defs/$ref), matching what tool-calling APIs expect for parameters.
//
// GenerateStructured and agent.NewTypedTool both build on this.
func SchemaFor[T any]() (json.RawMessage, error) {
	schema := schemaReflector.Reflect(new(T))
	schema.Version = "" // omit "$schema"; not expected in a tool's parameters
	return schema.MarshalJSON()
}

// GenerateStructured asks the model to answer by calling a synthetic tool
// whose parameters are the JSON schema reflected from T, then decodes the
// tool call arguments into a T value. It builds entirely on the existing
// tool-calling request/response shape (Tool, ToolChoice, ToolCall), so any
// Provider that supports tool calling supports structured output for free.
func GenerateStructured[T any](ctx context.Context, provider Provider, req *GenerateRequest) (T, error) {
	var zero T
	if provider == nil {
		return zero, errors.New("llm: provider is required")
	}
	if req == nil {
		return zero, errors.New("llm: request is required")
	}

	schemaJSON, err := SchemaFor[T]()
	if err != nil {
		return zero, fmt.Errorf("llm: build schema for structured output: %w", err)
	}

	structuredReq := *req
	structuredReq.Tools = append(append([]Tool{}, req.Tools...), NewTool(
		structuredToolName,
		"Call this with the final structured result.",
		schemaJSON,
	))
	structuredReq.ToolChoice = &ToolChoice{Mode: ToolChoiceRequired, Name: structuredToolName}

	resp, err := provider.Generate(ctx, &structuredReq)
	if err != nil {
		return zero, err
	}

	for _, call := range resp.ToolCalls() {
		if call.Name != structuredToolName {
			continue
		}
		var out T
		if err := call.ParseArguments(&out); err != nil {
			return zero, fmt.Errorf("llm: decode structured result: %w", err)
		}
		return out, nil
	}
	return zero, errors.New("llm: model did not return a structured result")
}
