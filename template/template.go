package template

import (
	"fmt"
	"strings"
	"sync"

	"github.com/singhJasvinder101/agentic-go/template/formatter"
	"github.com/singhJasvinder101/agentic-go/template/formatter/jinja"
	"github.com/singhJasvinder101/agentic-go/template/formatter/native"
)


type TemplateEngine string

const (
	FormatterNative TemplateEngine = "native"
	FormatterJinja  TemplateEngine = "jinja"
)

func GetTemplateEngines() []TemplateEngine {
	return []TemplateEngine{FormatterNative, FormatterJinja}
}

type PromptTemplate struct {
	Key    string
	Prompt string
	Engine TemplateEngine
}

type Registry struct {
	mu        sync.RWMutex
	engines   map[TemplateEngine]formatter.TemplateFormatter
	templates map[string]PromptTemplate
}

func NewRegistry() *Registry {
	registry := &Registry{
		engines:   make(map[TemplateEngine]formatter.TemplateFormatter),
		templates: make(map[string]PromptTemplate),
	}

	for _, engine := range GetTemplateEngines() {
		_ = registry.RegisterTemplateEngine(engine, initTemplateEngine(engine))
	}
	return registry
}

func (r *Registry) RegisterTemplateEngine(name TemplateEngine, engine formatter.TemplateFormatter) error {
	if r == nil {
		return ErrNilTemplate
	}

	name = name.Normalize()
	if name == "" || engine == nil {
		return ErrEmptyFormatter
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.engines[name] = engine
	return nil
}

func (r *Registry) RegisterTemplate(key string, engineName TemplateEngine, prompt string) error {
	if r == nil {
		return ErrNilTemplate
	}

	key = normalizeKey(key)
	engineName = engineName.Normalize()
	prompt = strings.TrimSpace(prompt)

	if key == "" {
		return ErrEmptyTemplateKey
	}
	if engineName == "" {
		return ErrEmptyFormatter
	}
	if prompt == "" {
		return ErrEmptyTemplate
	}

	if _, ok := r.templates[key]; ok {
		return fmt.Errorf("%w: %s", ErrTemplateAlreadyExists, key)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.engines[engineName]; !ok {
		return fmt.Errorf("%w: %s", ErrFormatterNotFound, engineName)
	}

	r.templates[key] = PromptTemplate{
		Key:    key,
		Prompt: prompt,
		Engine: engineName,
	}
	return nil
}

func (r *Registry) Format(key string, variables map[string]any) (string, error) {
	promptTemplate, err := r.GetTemplate(key)
	if err != nil {
		return "", err
	}
	return r.FormatTemplate(promptTemplate, variables)
}

func (r *Registry) GetTemplate(key string) (PromptTemplate, error) {
	if r == nil {
		return PromptTemplate{}, ErrNilTemplate
	}

	key = normalizeKey(key)
	if key == "" {
		return PromptTemplate{}, ErrEmptyTemplateKey
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	promptTemplate, ok := r.templates[key]
	if !ok {
		return PromptTemplate{}, fmt.Errorf("%w: %s", ErrTemplateNotFound, key)
	}

	return promptTemplate, nil
}

func (r *Registry) FormatTemplate(promptTemplate PromptTemplate, variables map[string]any) (string, error) {
	if r == nil {
		return "", ErrNilTemplate
	}

	r.mu.RLock()
	templateRenderer, ok := r.engines[promptTemplate.Engine]
	r.mu.RUnlock()
	
	if !ok {
		return "", fmt.Errorf("%w: %s", ErrFormatterNotFound, promptTemplate.Engine)
	}

	formatted, err := templateRenderer.Render(promptTemplate.Prompt, variables)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrTemplateExecute, err)
	}

	return formatted, nil
}

func normalizeKey(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func (e TemplateEngine) Normalize() TemplateEngine {
	return TemplateEngine(strings.ToLower(strings.TrimSpace(string(e))))
}

func initTemplateEngine(name TemplateEngine) formatter.TemplateFormatter {
	switch name {
	case FormatterNative:
		return native.New()
	case FormatterJinja:
		return jinja.New()
	default:
		return nil
	}
}