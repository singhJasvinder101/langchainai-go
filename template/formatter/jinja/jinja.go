package jinja

import (
	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/exec"
)

type JinjaTemplate struct{}

func New() *JinjaTemplate {
	return &JinjaTemplate{}
}

func (t *JinjaTemplate) Render(promptTemplate string, variables map[string]any) (string, error) {
	tmpl, err := gonja.FromString(promptTemplate)
	if err != nil {
		return "", err
	}

	return tmpl.ExecuteToString(exec.NewContext(variables))
}
