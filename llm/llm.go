package llm

type RequestInterface interface {
	Validate() error
}

type ResponseInterface interface {
	GetText() string
	GetResponse() any
}
