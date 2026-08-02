// Package agent runs the generate -> tool-call -> tool-result loop: it calls
// an llm.Provider, executes any tool calls the model requests via
// user-supplied handlers, feeds the results back, and repeats until the
// model returns a final answer.
package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/singhJasvinder101/agentic-go/llm"
	"github.com/singhJasvinder101/agentic-go/memory"
)

// defaultMaxIterations bounds the loop when MaxIterations is unset.
const defaultMaxIterations = 10

// ErrMaxIterations is returned when the model never converges on a final
// answer within MaxIterations turns.
var ErrMaxIterations = errors.New("agent: max iterations reached without a final response")

// Handler executes a tool call and returns its result as a string. Returning
// an error surfaces "error: <message>" to the model as the tool result,
// rather than aborting the run, so the model can see the failure and retry
// or adjust its approach.
type Handler func(ctx context.Context, args json.RawMessage) (string, error)

// Tool pairs an llm.Tool definition with the function that executes it.
type Tool struct {
	Definition llm.Tool
	Handler    Handler
}

// NewTool builds a Tool from a name, description, JSON schema, and handler.
func NewTool(name, description string, parameters json.RawMessage, handler Handler) Tool {
	return Tool{Definition: llm.NewTool(name, description, parameters), Handler: handler}
}

// Agent runs the tool-calling loop against a Provider.
type Agent struct {
	Provider      llm.Provider
	Tools         []Tool
	SystemPrompt  string
	MaxIterations int
	Memory        memory.Memory
}

// Option configures an Agent built with New.
type Option func(*Agent)

// WithTools registers tools the agent may call.
func WithTools(tools ...Tool) Option {
	return func(a *Agent) { a.Tools = append(a.Tools, tools...) }
}

// WithSystemPrompt sets the system prompt sent with every turn.
func WithSystemPrompt(prompt string) Option {
	return func(a *Agent) { a.SystemPrompt = prompt }
}

// WithMaxIterations caps how many generate/tool-execute round trips a single
// Run performs before returning ErrMaxIterations.
func WithMaxIterations(n int) Option {
	return func(a *Agent) { a.MaxIterations = n }
}

// WithMemory attaches conversation history that is read at the start of Run
// and appended to once Run produces a final answer.
func WithMemory(m memory.Memory) Option {
	return func(a *Agent) { a.Memory = m }
}

// New builds an Agent around provider, applying the given options.
func New(provider llm.Provider, opts ...Option) *Agent {
	a := &Agent{Provider: provider, MaxIterations: defaultMaxIterations}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Result is the outcome of a single Run.
type Result struct {
	// Text is the model's final answer.
	Text string
	// Messages is the full transcript for this run, including the system
	// prompt (if any), prior memory, the user input, and every intermediate
	// tool call/result.
	Messages []llm.Message
	// Steps is the number of generate calls it took to reach a final answer.
	Steps int
}

// Run sends input to the model and executes any tool calls it requests,
// repeating until the model returns a final text answer or MaxIterations is
// exceeded.
func (a *Agent) Run(ctx context.Context, input string) (*Result, error) {
	if a.Provider == nil {
		return nil, errors.New("agent: provider is required")
	}

	maxIterations := a.MaxIterations
	if maxIterations <= 0 {
		maxIterations = defaultMaxIterations
	}

	tools := make([]llm.Tool, 0, len(a.Tools))
	handlers := make(map[string]Handler, len(a.Tools))
	for _, tool := range a.Tools {
		tools = append(tools, tool.Definition)
		handlers[tool.Definition.Name] = tool.Handler
	}

	var messages []llm.Message
	if a.SystemPrompt != "" {
		messages = append(messages, llm.SystemMessage(llm.TextPart(a.SystemPrompt)))
	}
	if a.Memory != nil {
		messages = append(messages, a.Memory.Messages()...)
	}
	userMsg := llm.UserMessage(llm.TextPart(input))
	messages = append(messages, userMsg)

	for step := 0; step < maxIterations; step++ {
		resp, err := a.Provider.Generate(ctx, &llm.GenerateRequest{Messages: messages, Tools: tools})
		if err != nil {
			return nil, fmt.Errorf("agent: generate: %w", err)
		}

		assistantMsg := resp.AssistantMessage()
		messages = append(messages, assistantMsg)

		calls := resp.ToolCalls()
		if len(calls) == 0 {
			if a.Memory != nil {
				a.Memory.Add(userMsg, assistantMsg)
			}
			return &Result{Text: resp.Text(), Messages: messages, Steps: step + 1}, nil
		}

		for _, call := range calls {
			messages = append(messages, llm.ToolMessage(call.ID, call.Name, a.execute(ctx, call, handlers)))
		}
	}

	return nil, ErrMaxIterations
}

func (a *Agent) execute(ctx context.Context, call llm.ToolCall, handlers map[string]Handler) string {
	handler, ok := handlers[call.Name]
	if !ok {
		return fmt.Sprintf("error: unknown tool %q", call.Name)
	}
	result, err := handler(ctx, call.Arguments)
	if err != nil {
		return fmt.Sprintf("error: %v", err)
	}
	return result
}
