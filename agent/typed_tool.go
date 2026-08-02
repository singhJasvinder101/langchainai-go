package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/singhJasvinder101/agentic-go/llm"
)

// TypedHandler executes a tool call with its arguments already decoded into
// Args, instead of raw JSON.
type TypedHandler[Args any] func(ctx context.Context, args Args) (string, error)

// NewTypedTool builds a Tool the Pydantic-model way: define Args as a plain
// Go struct (with `json` tags, and optionally `jsonschema_description` tags
// for field descriptions) and get both the parameters JSON schema and typed
// argument decoding for free — no hand-written JSON schema string, no
// json.Unmarshal in the handler.
//
//	type WeatherArgs struct {
//		City string `json:"city" jsonschema_description:"City name, e.g. Paris"`
//	}
//
//	weatherTool := agent.NewTypedTool("get_weather", "Get current weather for a city",
//		func(ctx context.Context, args WeatherArgs) (string, error) {
//			return fmt.Sprintf(`{"temp_c":18,"city":%q}`, args.City), nil
//		})
//
// Args is a compile-time-known static type, so a schema that fails to build
// (essentially only possible for pathological types) panics rather than
// returning an error, the same way template.Registry.MustRegisterTemplate
// does for its compile-time-known templates.
func NewTypedTool[Args any](name, description string, handler TypedHandler[Args]) Tool {
	schema, err := llm.SchemaFor[Args]()
	if err != nil {
		panic(fmt.Sprintf("agent: build schema for tool %q: %v", name, err))
	}

	return Tool{
		Definition: llm.NewTool(name, description, schema),
		Handler: func(ctx context.Context, raw json.RawMessage) (string, error) {
			var args Args
			if len(raw) > 0 {
				if err := json.Unmarshal(raw, &args); err != nil {
					return "", fmt.Errorf("agent: decode arguments for tool %q: %w", name, err)
				}
			}
			return handler(ctx, args)
		},
	}
}
