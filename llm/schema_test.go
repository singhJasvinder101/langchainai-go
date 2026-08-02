package llm_test

import (
	"encoding/json"
	"testing"

	"github.com/singhJasvinder101/agentic-go/llm"
)

type schemaArgs struct {
	City string `json:"city" jsonschema_description:"City name, e.g. Paris"`
}

func TestSchemaForReflectsStructTags(t *testing.T) {
	raw, err := llm.SchemaFor[schemaArgs]()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("schema is not valid JSON: %v", err)
	}

	if _, hasSchemaKey := decoded["$schema"]; hasSchemaKey {
		t.Fatalf("expected $schema to be omitted from a tool's parameters schema, got %v", decoded)
	}
	if decoded["type"] != "object" {
		t.Fatalf("expected type object, got %v", decoded["type"])
	}

	props, ok := decoded["properties"].(map[string]any)
	if !ok {
		t.Fatalf("expected a properties object, got %v", decoded["properties"])
	}
	city, ok := props["city"].(map[string]any)
	if !ok {
		t.Fatalf("expected a city property, got %v", props)
	}
	if city["type"] != "string" {
		t.Fatalf("expected city type string, got %v", city["type"])
	}
	if city["description"] != "City name, e.g. Paris" {
		t.Fatalf("expected description reflected from the jsonschema_description tag, got %v", city["description"])
	}

	required, _ := decoded["required"].([]any)
	if len(required) != 1 || required[0] != "city" {
		t.Fatalf("expected city to be required (no omitempty), got %v", decoded["required"])
	}
}
