package formatter

type TemplateFormatter interface {
	Render(template string, variables map[string]any) (string, error)
}
