package memory_test

import (
	"sync"
	"testing"

	"github.com/singhJasvinder101/agentic-go/llm"
	"github.com/singhJasvinder101/agentic-go/memory"
)

func TestBufferAddAndMessages(t *testing.T) {
	b := memory.NewBuffer()
	b.Add(llm.UserMessage(llm.TextPart("hi")))
	b.Add(llm.AssistantMessage(llm.TextPart("hello")))

	got := b.Messages()
	if len(got) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(got))
	}
	if got[0].Role != llm.RoleUser || got[1].Role != llm.RoleAssistant {
		t.Fatalf("unexpected message order/roles: %+v", got)
	}
}

func TestBufferMessagesReturnsCopy(t *testing.T) {
	b := memory.NewBuffer()
	b.Add(llm.UserMessage(llm.TextPart("hi")))

	got := b.Messages()
	got[0] = llm.UserMessage(llm.TextPart("mutated"))

	again := b.Messages()
	if again[0].Parts[0].Text != "hi" {
		t.Fatalf("expected internal state unaffected by caller mutation, got %q", again[0].Parts[0].Text)
	}
}

func TestBufferWithMaxMessagesDropsOldest(t *testing.T) {
	b := memory.NewBuffer(memory.WithMaxMessages(2))
	b.Add(llm.UserMessage(llm.TextPart("one")))
	b.Add(llm.UserMessage(llm.TextPart("two")))
	b.Add(llm.UserMessage(llm.TextPart("three")))

	got := b.Messages()
	if len(got) != 2 {
		t.Fatalf("expected window of 2 messages, got %d", len(got))
	}
	if got[0].Parts[0].Text != "two" || got[1].Parts[0].Text != "three" {
		t.Fatalf("expected oldest message dropped, got %+v", got)
	}
}

func TestBufferClear(t *testing.T) {
	b := memory.NewBuffer()
	b.Add(llm.UserMessage(llm.TextPart("hi")))
	b.Clear()

	if got := b.Messages(); len(got) != 0 {
		t.Fatalf("expected empty history after Clear, got %+v", got)
	}
}

func TestBufferConcurrentAccess(t *testing.T) {
	b := memory.NewBuffer(memory.WithMaxMessages(50))

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b.Add(llm.UserMessage(llm.TextPart("hi")))
			_ = b.Messages()
		}()
	}
	wg.Wait()

	if got := len(b.Messages()); got != 20 {
		t.Fatalf("expected 20 messages, got %d", got)
	}
}
