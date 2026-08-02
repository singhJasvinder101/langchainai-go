package llm_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/singhJasvinder101/agentic-go/llm"
)

func chunk(text string) *llm.StreamResponse {
	return &llm.StreamResponse{
		Choices: []llm.StreamChoice{llm.NewStreamChoice(0, []llm.ContentPart{llm.TextPart(text)}, "")},
	}
}

func TestStreamNextIteratesChunksInOrder(t *testing.T) {
	responses := make(chan *llm.StreamResponse)
	errs := make(chan error)

	go func() {
		responses <- chunk("Hello")
		responses <- chunk(" world")
		close(responses)
		close(errs)
	}()

	stream := llm.NewStream(responses, errs)

	var got []string
	for stream.Next() {
		got = append(got, stream.Text())
	}
	if err := stream.Err(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"Hello", " world"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestStreamNextStopsOnErrorAndKeepsPriorChunks(t *testing.T) {
	responses := make(chan *llm.StreamResponse)
	errs := make(chan error)
	wantErr := errors.New("boom")

	go func() {
		responses <- chunk("partial")
		close(responses)
		errs <- wantErr
		close(errs)
	}()

	stream := llm.NewStream(responses, errs)

	var got []string
	for stream.Next() {
		got = append(got, stream.Text())
	}

	if !errors.Is(stream.Err(), wantErr) {
		t.Fatalf("expected error %v, got %v", wantErr, stream.Err())
	}
	if !reflect.DeepEqual(got, []string{"partial"}) {
		t.Fatalf("expected the chunk read before the error to be kept, got %v", got)
	}
}

func TestStreamNextEmptyStreamReturnsFalseImmediately(t *testing.T) {
	responses := make(chan *llm.StreamResponse)
	errs := make(chan error)
	close(responses)
	close(errs)

	stream := llm.NewStream(responses, errs)
	if stream.Next() {
		t.Fatal("expected Next to return false for an empty stream")
	}
	if stream.Err() != nil {
		t.Fatalf("expected no error, got %v", stream.Err())
	}
}

func TestStreamCurrentAndTextBeforeAnyNext(t *testing.T) {
	responses := make(chan *llm.StreamResponse)
	errs := make(chan error)
	close(responses)
	close(errs)

	stream := llm.NewStream(responses, errs)
	if stream.Current() != nil {
		t.Fatalf("expected nil current before any Next call, got %+v", stream.Current())
	}
	if stream.Text() != "" {
		t.Fatalf("expected empty text before any Next call, got %q", stream.Text())
	}
}
