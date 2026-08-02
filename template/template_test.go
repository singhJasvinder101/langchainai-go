package template

import (
	"errors"
	"fmt"
	"sync"
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

func TestRenderNativeWithoutRegistry(t *testing.T) {
	got, err := Render(FormatterNative, "Hello {{.Name}}", map[string]any{"Name": "Ada"})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if got != "Hello Ada" {
		t.Fatalf("expected %q, got %q", "Hello Ada", got)
	}
}

func TestRenderJinjaWithoutRegistry(t *testing.T) {
	got, err := Render(FormatterJinja, "Hello {{ name }}", map[string]any{"name": "Ada"})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if got != "Hello Ada" {
		t.Fatalf("expected %q, got %q", "Hello Ada", got)
	}
}

func TestRenderDefaultsToNativeWhenEngineOmitted(t *testing.T) {
	got, err := Render("", "Hello {{.Name}}", map[string]any{"Name": "Ada"})
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	if got != "Hello Ada" {
		t.Fatalf("expected %q, got %q", "Hello Ada", got)
	}
}

func TestRenderRejectsUnknownEngine(t *testing.T) {
	_, err := Render("unknown", "Hello", nil)
	if !errors.Is(err, ErrFormatterNotFound) {
		t.Fatalf("expected ErrFormatterNotFound, got %v", err)
	}
}

func TestMustRegisterTemplatePanicsOnError(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for an invalid template")
		}
	}()
	NewRegistry().MustRegisterTemplate("bad", "unknown-engine", "Hello")
}

func TestMustRegisterTemplateSucceeds(t *testing.T) {
	registry := NewRegistry()
	registry.MustRegisterTemplate("greeting", FormatterNative, "Hello {{.Name}}")

	got, err := registry.Format("greeting", map[string]any{"Name": "Ada"})
	if err != nil {
		t.Fatalf("format template: %v", err)
	}
	if got != "Hello Ada" {
		t.Fatalf("expected %q, got %q", "Hello Ada", got)
	}
}

// TestRegistryConcurrentRegistration guards against a regression of the race
// where RegisterTemplate checked r.templates for an existing key before
// acquiring r.mu. Run with -race to verify.
func TestRegistryConcurrentRegistration(t *testing.T) {
	registry := NewRegistry()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", i)
			if err := registry.RegisterTemplate(key, FormatterNative, "Hello {{.Name}}"); err != nil {
				t.Errorf("register template %q: %v", key, err)
			}
		}(i)
	}
	wg.Wait()

	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("key-%d", i)
		if _, err := registry.GetTemplate(key); err != nil {
			t.Fatalf("expected template %q to be registered: %v", key, err)
		}
	}
}
