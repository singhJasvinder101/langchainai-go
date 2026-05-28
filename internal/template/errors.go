package template

import "errors"

var (
	ErrNilTemplate       = errors.New("prompt template is required")
	ErrEmptyTemplate     = errors.New("prompt template body is required")
	ErrInvalidTemplate   = errors.New("invalid prompt template")
	ErrTemplateExecute   = errors.New("failed to execute prompt template")
	ErrEmptyTemplateKey  = errors.New("prompt template key is required")
	ErrTemplateNotFound  = errors.New("prompt template not found")
	ErrEmptyFormatter    = errors.New("prompt template formatter is required")
	ErrFormatterNotFound = errors.New("prompt template formatter not found")
)
