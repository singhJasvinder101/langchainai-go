package native

import (
	"bytes"
	"text/template"
)

type NativeTemplate struct{}

func New() *NativeTemplate {
	return &NativeTemplate{}
}

func (t *NativeTemplate) Render(promptTemplate string, variables map[string]any) (string, error) {
	tmpl, err := template.New("prompt").Option("missingkey=error").Parse(promptTemplate)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, variableData(variables)); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func variableData(variables map[string]any) map[string]any {
	if len(variables) == 0 {
		return map[string]any{}
	}

	data := make(map[string]any, len(variables))
	for key, value := range variables {
		data[key] = value
	}
	return data
}
