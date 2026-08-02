package llm

import "context"

// Provider is implemented by every chat LLM backend (llm/openai, llm/claude,
// llm/gemini, llm/ollama, ...). Code that orchestrates providers (agent, chain,
// retriever) should depend on this interface rather than a concrete provider
// package, so it works with whichever provider the caller constructs.
type Provider interface {
	Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
	GenerateStream(ctx context.Context, req *GenerateRequest) (<-chan *StreamResponse, <-chan error)
	Close() error
}
