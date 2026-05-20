package llm

import "context"

type RequestInterface interface {
	Validate() error
}

type ResponseInterface interface {
	GetText() string
	GetResponse() any
}

type Provider interface {
	Generate(ctx context.Context, request RequestInterface) (response ResponseInterface, err error)
	GenerateStream(ctx context.Context, request RequestInterface) (responses <-chan ResponseInterface, errs <-chan error)
	Close() error
}