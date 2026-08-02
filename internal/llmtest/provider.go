// Package llmtest provides a scripted llm.Provider for use in tests across
// the agent, chain, retriever, and llm packages, so each doesn't need to
// redefine its own fake provider.
package llmtest

import (
	"context"
	"errors"
	"sync"

	"github.com/singhJasvinder101/agentic-go/llm"
)

type step struct {
	resp *llm.GenerateResponse
	err  error
}

// MockProvider is a scripted llm.Provider: each call to Generate consumes the
// next queued step (a response or an error), in the order they were added.
type MockProvider struct {
	mu    sync.Mutex
	steps []step
	calls []*llm.GenerateRequest
}

// NewMockProvider builds an empty MockProvider; chain AddResponse/AddError to
// script its behavior.
func NewMockProvider() *MockProvider {
	return &MockProvider{}
}

// AddResponse queues a response to be returned by the next Generate call.
func (m *MockProvider) AddResponse(resp *llm.GenerateResponse) *MockProvider {
	m.steps = append(m.steps, step{resp: resp})
	return m
}

// AddError queues an error to be returned by the next Generate call.
func (m *MockProvider) AddError(err error) *MockProvider {
	m.steps = append(m.steps, step{err: err})
	return m
}

func (m *MockProvider) Generate(ctx context.Context, req *llm.GenerateRequest) (*llm.GenerateResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls = append(m.calls, req)
	if len(m.steps) == 0 {
		return nil, errors.New("llmtest: no more scripted responses")
	}
	s := m.steps[0]
	m.steps = m.steps[1:]
	if s.err != nil {
		return nil, s.err
	}
	return s.resp, nil
}

func (m *MockProvider) GenerateStream(ctx context.Context, req *llm.GenerateRequest) (<-chan *llm.StreamResponse, <-chan error) {
	responses := make(chan *llm.StreamResponse)
	errs := make(chan error, 1)
	close(responses)
	errs <- errors.New("llmtest: MockProvider does not support streaming")
	close(errs)
	return responses, errs
}

func (m *MockProvider) Close() error { return nil }

// Calls returns every request passed to Generate so far, in order.
func (m *MockProvider) Calls() []*llm.GenerateRequest {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*llm.GenerateRequest, len(m.calls))
	copy(out, m.calls)
	return out
}

// TextResponse builds a final assistant response with a single text part and
// no tool calls.
func TextResponse(text string) *llm.GenerateResponse {
	return &llm.GenerateResponse{
		Choices: []llm.Choice{llm.NewChoice(0, []llm.ContentPart{llm.TextPart(text)}, llm.FinishReasonStop)},
	}
}

// ToolCallResponse builds an assistant response requesting the given tool calls.
func ToolCallResponse(calls ...llm.ToolCall) *llm.GenerateResponse {
	parts := make([]llm.ContentPart, len(calls))
	for i, c := range calls {
		parts[i] = llm.ToolCallPart(c)
	}
	return &llm.GenerateResponse{
		Choices: []llm.Choice{llm.NewChoice(0, parts, llm.FinishReasonToolCalls)},
	}
}
