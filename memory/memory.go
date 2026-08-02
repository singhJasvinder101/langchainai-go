// Package memory holds conversation history across turns, independent of any
// specific LLM provider or agent.
package memory

import (
	"sync"

	"github.com/singhJasvinder101/agentic-go/llm"
)

// Memory stores and retrieves conversation history.
type Memory interface {
	// Messages returns the stored history, oldest first.
	Messages() []llm.Message
	// Add appends messages to the history.
	Add(msgs ...llm.Message)
	// Clear removes all stored history.
	Clear()
}

// Buffer is an in-process, thread-safe Memory backed by a slice, with an
// optional window limiting how many messages are retained.
type Buffer struct {
	mu       sync.RWMutex
	messages []llm.Message
	maxSize  int // 0 means unbounded
}

// BufferOption configures a Buffer built with NewBuffer.
type BufferOption func(*Buffer)

// WithMaxMessages caps the buffer at n messages; once exceeded, the oldest
// messages are dropped first.
func WithMaxMessages(n int) BufferOption {
	return func(b *Buffer) { b.maxSize = n }
}

// NewBuffer builds an empty Buffer.
func NewBuffer(opts ...BufferOption) *Buffer {
	b := &Buffer{}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

func (b *Buffer) Messages() []llm.Message {
	b.mu.RLock()
	defer b.mu.RUnlock()

	out := make([]llm.Message, len(b.messages))
	copy(out, b.messages)
	return out
}

func (b *Buffer) Add(msgs ...llm.Message) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.messages = append(b.messages, msgs...)
	if b.maxSize > 0 && len(b.messages) > b.maxSize {
		b.messages = b.messages[len(b.messages)-b.maxSize:]
	}
}

func (b *Buffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.messages = nil
}
