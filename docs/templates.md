# Templates

The `template` package renders prompt templates with two interchangeable engines: Go's native `text/template` syntax, and Jinja2 syntax (via [gonja](https://github.com/NikolaLohinski/gonja)). There are two ways to use it — reach for whichever matches how many templates you have.

## The common case: `Render`, no setup

For a one-off template, `template.Render` formats it directly — no registry to create, no key to invent:

```go
import "github.com/singhJasvinder101/agentic-go/template"

prompt, err := template.Render(template.FormatterNative, "Summarize the following text: {{.Text}}", map[string]any{
	"Text": "The quick brown fox...",
})
```

`Render`'s `engine` argument defaults to `FormatterNative` if left as `""`, so a purely native template can drop it entirely:

```go
prompt, err := template.Render("", "Summarize: {{.Text}}", vars)
```

This is what [`chain.NewPromptTemplate`](chains.md) uses under the hood — you rarely need to call `Render` directly unless you're formatting text outside of a chain.

## Sharing templates: `Registry`

When several templates are reused across your app — registered once at startup, formatted many times by key — use a `Registry` instead:

```go
registry := template.NewRegistry() // both engines are registered automatically

registry.MustRegisterTemplate("summary", template.FormatterNative, "Summarize the following text: {{.Text}}")

prompt, err := registry.Format("summary", map[string]any{"Text": "The quick brown fox..."})
```

`MustRegisterTemplate` panics on error — appropriate for templates registered once at startup, where a bad template is a programming error you want to catch immediately, not handle at runtime. Use `RegisterTemplate` (which returns an `error` instead) if you're registering templates from user input or other runtime data.

`RegisterTemplate`/`MustRegisterTemplate` fail if `key` is already registered — templates are write-once per registry. Use `registry.GetTemplate(key)` to fetch the stored `PromptTemplate{Key, Prompt, Engine}` without rendering it.

## Native (`text/template`) vs Jinja

Both `Render` and `Registry` accept either engine:

```go
template.Render(template.FormatterNative, "Hello {{.Name}}", map[string]any{"Name": "Ada"}) // Go template: capitalized field access
template.Render(template.FormatterJinja, "Hello {{ name }}", map[string]any{"name": "Ada"}) // Jinja: lowercase variable access
```

- **`FormatterNative`** uses Go's standard `text/template` with `missingkey=error`, so referencing an undefined variable is a render error rather than silently rendering `<no value>`. Field access follows normal Go template rules (`{{.Field}}`, `{{.Nested.Field}}`, `{{range}}`, `{{if}}`, etc.).
- **`FormatterJinja`** uses Jinja2 syntax (`{{ variable }}`, `{{ obj.field }}`, `{% for %}`, `{% if %}`, filters, etc.) — familiar if you're porting prompts from Python LangChain.

Pick whichever syntax matches your team's existing prompts; both render to a plain string and there is no performance reason to prefer one over the other.

## Nested variables

Both engines accept arbitrarily nested `map[string]any` values:

```go
template.Render(template.FormatterNative, "{{.Text}} {{.Meta.Author}}", map[string]any{
	"Text": "...",
	"Meta": map[string]any{"Author": "Ada"},
})
```

## Custom formatters

`Registry.RegisterTemplateEngine(name, formatter)` lets you plug in a third engine by implementing the one-method `formatter.TemplateFormatter` interface:

```go
type TemplateFormatter interface {
	Render(template string, variables map[string]any) (string, error)
}
```

## See also

- [Chains](chains.md) — `chain.NewPromptTemplate`/`chain.NewPrompt` render a template and send it to an LLM provider
