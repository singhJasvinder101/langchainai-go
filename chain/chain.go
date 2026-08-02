// Package chain composes prompt templates and LLM providers into small,
// reusable units of work. It is intentionally minimal: a Chain takes named
// variables and returns text, nothing more.
package chain

import (
	"context"
	"fmt"

	"github.com/singhJasvinder101/agentic-go/llm"
	"github.com/singhJasvinder101/agentic-go/template"
)

// Chain is a composable unit of work: render input variables into some
// operation and return text output.
type Chain interface {
	Run(ctx context.Context, vars map[string]any) (string, error)
}

// Prompt renders a template and sends the result to a provider as a single
// user turn. Build one with NewPromptTemplate for a one-off template (no
// Registry needed), or NewPrompt to reuse a template already registered by
// key in a shared Registry.
type Prompt struct {
	Provider     llm.Provider
	SystemPrompt string

	// Template + Engine render directly, with no Registry involved. Set by
	// NewPromptTemplate.
	Template string
	Engine   template.TemplateEngine // defaults to template.FormatterNative

	// Registry + TemplateKey render a template already registered elsewhere.
	// Set by NewPrompt. Ignored if Template is set.
	Registry    *template.Registry
	TemplateKey string
}

// NewPromptTemplate builds a Prompt chain directly from template text — no
// Registry required. This is the common case: a single prompt used by one
// chain. Defaults to native (Go text/template) syntax; set the returned
// Prompt's Engine field for Jinja.
func NewPromptTemplate(provider llm.Provider, promptText string) *Prompt {
	return &Prompt{Provider: provider, Template: promptText, Engine: template.FormatterNative}
}

// NewPrompt builds a Prompt chain that formats templateKey from registry and
// sends it to provider. Reach for this when several chains share templates
// registered once in a common Registry; otherwise prefer NewPromptTemplate.
func NewPrompt(provider llm.Provider, registry *template.Registry, templateKey string) *Prompt {
	return &Prompt{Provider: provider, Registry: registry, TemplateKey: templateKey}
}

func (p *Prompt) Run(ctx context.Context, vars map[string]any) (string, error) {
	if p.Provider == nil {
		return "", fmt.Errorf("chain: provider is required")
	}

	rendered, err := p.render(vars)
	if err != nil {
		return "", err
	}

	var messages []llm.Message
	if p.SystemPrompt != "" {
		messages = append(messages, llm.SystemMessage(llm.TextPart(p.SystemPrompt)))
	}
	messages = append(messages, llm.UserMessage(llm.TextPart(rendered)))

	resp, err := p.Provider.Generate(ctx, &llm.GenerateRequest{Messages: messages})
	if err != nil {
		return "", err
	}
	return resp.Text(), nil
}

func (p *Prompt) render(vars map[string]any) (string, error) {
	if p.Template != "" {
		return template.Render(p.Engine, p.Template, vars)
	}
	if p.Registry == nil {
		return "", fmt.Errorf("chain: template or registry is required")
	}
	return p.Registry.Format(p.TemplateKey, vars)
}

type sequential struct {
	chains    []Chain
	outputKey string
}

// SequentialOption configures a Chain built with Sequential.
type SequentialOption func(*sequential)

// WithOutputKey sets the vars key each chain's output is stored under for the
// next chain in the sequence. Defaults to "input".
func WithOutputKey(key string) SequentialOption {
	return func(s *sequential) { s.outputKey = key }
}

// Sequential runs chains in order. Each chain receives the vars passed to
// Run, extended with the previous chain's text output under outputKey (see
// WithOutputKey), and Sequential returns the last chain's output.
func Sequential(chains []Chain, opts ...SequentialOption) Chain {
	s := &sequential{chains: chains, outputKey: "input"}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *sequential) Run(ctx context.Context, vars map[string]any) (string, error) {
	current := vars
	var output string
	for i, c := range s.chains {
		out, err := c.Run(ctx, current)
		if err != nil {
			return "", fmt.Errorf("chain at index %d: %w", i, err)
		}
		output = out

		next := make(map[string]any, len(current)+1)
		for k, v := range current {
			next[k] = v
		}
		next[s.outputKey] = output
		current = next
	}
	return output, nil
}
