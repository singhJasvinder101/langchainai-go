package chain_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/singhJasvinder101/agentic-go/chain"
	"github.com/singhJasvinder101/agentic-go/internal/llmtest"
	"github.com/singhJasvinder101/agentic-go/template"
)

func TestPromptRunFormatsTemplateAndSendsToProvider(t *testing.T) {
	registry := template.NewRegistry()
	if err := registry.RegisterTemplate("summary", template.FormatterNative, "Summarize: {{.Text}}"); err != nil {
		t.Fatalf("register template: %v", err)
	}

	provider := llmtest.NewMockProvider().AddResponse(llmtest.TextResponse("a short summary"))
	p := chain.NewPrompt(provider, registry, "summary")
	p.SystemPrompt = "You are concise."

	got, err := p.Run(context.Background(), map[string]any{"Text": "the quick brown fox"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "a short summary" {
		t.Fatalf("expected %q, got %q", "a short summary", got)
	}

	sent := provider.Calls()[0].Messages
	if len(sent) != 2 {
		t.Fatalf("expected system + user message, got %d", len(sent))
	}
	if sent[1].Parts[0].Text != "Summarize: the quick brown fox" {
		t.Fatalf("expected rendered template as user message, got %q", sent[1].Parts[0].Text)
	}
}

func TestPromptRunRequiresProviderAndRegistry(t *testing.T) {
	registry := template.NewRegistry()
	if _, err := (&chain.Prompt{Registry: registry}).Run(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil provider")
	}
	provider := llmtest.NewMockProvider()
	if _, err := (&chain.Prompt{Provider: provider}).Run(context.Background(), nil); err == nil {
		t.Fatal("expected error when neither Template nor Registry is set")
	}
}

func TestNewPromptTemplateRendersWithoutARegistry(t *testing.T) {
	provider := llmtest.NewMockProvider().AddResponse(llmtest.TextResponse("a short summary"))
	p := chain.NewPromptTemplate(provider, "Summarize: {{.Text}}")

	got, err := p.Run(context.Background(), map[string]any{"Text": "the quick brown fox"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "a short summary" {
		t.Fatalf("expected %q, got %q", "a short summary", got)
	}

	sent := provider.Calls()[0].Messages
	if len(sent) != 1 || sent[0].Parts[0].Text != "Summarize: the quick brown fox" {
		t.Fatalf("expected rendered template as the only message, got %+v", sent)
	}
}

func TestNewPromptTemplateSupportsJinjaEngine(t *testing.T) {
	provider := llmtest.NewMockProvider().AddResponse(llmtest.TextResponse("done"))
	p := chain.NewPromptTemplate(provider, "Summarize: {{ text }}")
	p.Engine = template.FormatterJinja

	if _, err := p.Run(context.Background(), map[string]any{"text": "the quick brown fox"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sent := provider.Calls()[0].Messages
	if sent[0].Parts[0].Text != "Summarize: the quick brown fox" {
		t.Fatalf("expected rendered jinja template, got %q", sent[0].Parts[0].Text)
	}
}

func TestPromptTemplateTakesPriorityOverRegistry(t *testing.T) {
	registry := template.NewRegistry()
	registry.MustRegisterTemplate("summary", template.FormatterNative, "Registry: {{.Text}}")

	provider := llmtest.NewMockProvider().AddResponse(llmtest.TextResponse("done"))
	p := chain.NewPrompt(provider, registry, "summary")
	p.Template = "Direct: {{.Text}}"
	p.Engine = template.FormatterNative

	if _, err := p.Run(context.Background(), map[string]any{"Text": "x"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sent := provider.Calls()[0].Messages
	if sent[0].Parts[0].Text != "Direct: x" {
		t.Fatalf("expected Template to take priority over Registry, got %q", sent[0].Parts[0].Text)
	}
}

type fakeChain struct {
	run func(ctx context.Context, vars map[string]any) (string, error)
}

func (f fakeChain) Run(ctx context.Context, vars map[string]any) (string, error) {
	return f.run(ctx, vars)
}

func TestSequentialPipesOutputIntoNextChainVars(t *testing.T) {
	var secondSawInput string
	first := fakeChain{run: func(ctx context.Context, vars map[string]any) (string, error) {
		return fmt.Sprintf("processed:%v", vars["topic"]), nil
	}}
	second := fakeChain{run: func(ctx context.Context, vars map[string]any) (string, error) {
		secondSawInput, _ = vars["input"].(string)
		return "final:" + secondSawInput, nil
	}}

	c := chain.Sequential([]chain.Chain{first, second})
	got, err := c.Run(context.Background(), map[string]any{"topic": "go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secondSawInput != "processed:go" {
		t.Fatalf("expected second chain to see first chain's output, got %q", secondSawInput)
	}
	if got != "final:processed:go" {
		t.Fatalf("unexpected final output: %q", got)
	}
}

func TestSequentialWithOutputKey(t *testing.T) {
	var secondSawVars map[string]any
	first := fakeChain{run: func(ctx context.Context, vars map[string]any) (string, error) { return "answer", nil }}
	second := fakeChain{run: func(ctx context.Context, vars map[string]any) (string, error) {
		secondSawVars = vars
		return "done", nil
	}}

	c := chain.Sequential([]chain.Chain{first, second}, chain.WithOutputKey("previous"))
	if _, err := c.Run(context.Background(), map[string]any{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if secondSawVars["previous"] != "answer" {
		t.Fatalf("expected custom output key populated, got %+v", secondSawVars)
	}
}

func TestSequentialPropagatesChainError(t *testing.T) {
	wantErr := errors.New("boom")
	failing := fakeChain{run: func(ctx context.Context, vars map[string]any) (string, error) { return "", wantErr }}

	c := chain.Sequential([]chain.Chain{failing})
	_, err := c.Run(context.Background(), map[string]any{})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped chain error, got %v", err)
	}
}
