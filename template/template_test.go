package template

import (
	"errors"
	"testing"
)

func TestRegistryFormatsNativeTemplate(t *testing.T) {
	registry := NewRegistry()

	err := registry.RegisterTemplate("greeting", FormatterNative, "Hello {{.Name}}")
	if err != nil {
		t.Fatalf("register template: %v", err)
	}

	got, err := registry.Format("greeting", map[string]any{"Name": "Jasvinder"})
	if err != nil {
		t.Fatalf("format template: %v", err)
	}

	if got != "Hello Jasvinder" {
		t.Fatalf("expected formatted prompt, got %q", got)
	}
}

func TestRegistryFormatsJinjaTemplate(t *testing.T) {
	registry := NewRegistry()

	err := registry.RegisterTemplate("greeting", FormatterJinja, "Hello {{ user.name }}")
	if err != nil {
		t.Fatalf("register template: %v", err)
	}

	got, err := registry.Format("greeting", map[string]any{
		"user": map[string]any{"name": "Jasvinder"},
	})
	if err != nil {
		t.Fatalf("format template: %v", err)
	}

	if got != "Hello Jasvinder" {
		t.Fatalf("expected formatted prompt, got %q", got)
	}
}

func TestRegistryReturnsTemplateByKey(t *testing.T) {
	registry := NewRegistry()

	err := registry.RegisterTemplate("summary", FormatterNative, "Summarize {{.Topic}}")
	if err != nil {
		t.Fatalf("register template: %v", err)
	}

	got, err := registry.GetTemplate("summary")
	if err != nil {
		t.Fatalf("get template: %v", err)
	}

	if got.Key != "summary" || got.Engine != FormatterNative || got.Prompt != "Summarize {{.Topic}}" {
		t.Fatalf("unexpected template: %+v", got)
	}
}

func TestRegistryRejectsUnknownFormatter(t *testing.T) {
	registry := NewRegistry()

	err := registry.RegisterTemplate("bad", "unknown", "Hello")
	if !errors.Is(err, ErrFormatterNotFound) {
		t.Fatalf("expected ErrFormatterNotFound, got %v", err)
	}
}

func TestRegistryReturnsTemplateNotFound(t *testing.T) {
	registry := NewRegistry()

	_, err := registry.GetTemplate("missing")
	if !errors.Is(err, ErrTemplateNotFound) {
		t.Fatalf("expected ErrTemplateNotFound, got %v", err)
	}
}
