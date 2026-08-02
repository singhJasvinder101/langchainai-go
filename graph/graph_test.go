package graph_test

import (
	"context"
	"errors"
	"testing"

	"github.com/singhJasvinder101/agentic-go/graph"
)

type pathState struct {
	path []string
}

func TestGraphRunsStaticEdgesToEnd(t *testing.T) {
	g := graph.New[pathState]()
	g.AddNode("a", func(ctx context.Context, s pathState) (pathState, error) {
		s.path = append(s.path, "a")
		return s, nil
	})
	g.AddNode("b", func(ctx context.Context, s pathState) (pathState, error) {
		s.path = append(s.path, "b")
		return s, nil
	})
	g.AddEdge("a", "b")
	g.AddEdge("b", graph.End)
	g.SetEntryPoint("a")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	final, err := runnable.Run(context.Background(), pathState{path: []string{}})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(final.path) != 2 || final.path[0] != "a" || final.path[1] != "b" {
		t.Fatalf("expected path [a b], got %v", final.path)
	}
}

type doneState struct {
	done bool
}

func TestGraphNodeWithNoOutgoingEdgeImplicitlyEnds(t *testing.T) {
	g := graph.New[doneState]()
	g.AddNode("only", func(ctx context.Context, s doneState) (doneState, error) {
		s.done = true
		return s, nil
	})
	g.SetEntryPoint("only")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	final, err := runnable.Run(context.Background(), doneState{})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if !final.done {
		t.Fatalf("expected done=true, got %+v", final)
	}
}

type classifyState struct {
	score  int
	result string
}

func TestGraphConditionalRouting(t *testing.T) {
	g := graph.New[classifyState]()
	g.AddNode("classify", func(ctx context.Context, s classifyState) (classifyState, error) {
		return s, nil
	})
	g.AddNode("positive", func(ctx context.Context, s classifyState) (classifyState, error) {
		s.result = "positive branch"
		return s, nil
	})
	g.AddNode("negative", func(ctx context.Context, s classifyState) (classifyState, error) {
		s.result = "negative branch"
		return s, nil
	})
	g.AddConditionalEdge("classify", func(s classifyState) string {
		if s.score > 0 {
			return "pos"
		}
		return "neg"
	}, map[string]string{"pos": "positive", "neg": "negative"})
	g.AddEdge("positive", graph.End)
	g.AddEdge("negative", graph.End)
	g.SetEntryPoint("classify")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	final, err := runnable.Run(context.Background(), classifyState{score: 5})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if final.result != "positive branch" {
		t.Fatalf("expected positive branch, got %+v", final)
	}

	final, err = runnable.Run(context.Background(), classifyState{score: -1})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if final.result != "negative branch" {
		t.Fatalf("expected negative branch, got %+v", final)
	}
}

type countState struct {
	count int
}

func TestGraphCycleGuardedByMaxSteps(t *testing.T) {
	g := graph.New[countState]()
	g.AddNode("loop", func(ctx context.Context, s countState) (countState, error) {
		s.count++
		return s, nil
	})
	g.AddEdge("loop", "loop")
	g.SetEntryPoint("loop")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	_, err = runnable.RunWithLimit(context.Background(), countState{}, 5)
	if err == nil {
		t.Fatal("expected error from unterminated cycle")
	}
}

type emptyState struct{}

func noopNode(ctx context.Context, s emptyState) (emptyState, error) { return s, nil }

func TestGraphCompileRequiresEntryPoint(t *testing.T) {
	g := graph.New[emptyState]()
	g.AddNode("a", noopNode)

	_, err := g.Compile()
	if !errors.Is(err, graph.ErrEntryRequired) {
		t.Fatalf("expected ErrEntryRequired, got %v", err)
	}
}

func TestGraphCompileRejectsUnknownEntryPoint(t *testing.T) {
	g := graph.New[emptyState]()
	g.AddNode("a", noopNode)
	g.SetEntryPoint("missing")

	_, err := g.Compile()
	if !errors.Is(err, graph.ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestGraphCompileRejectsEdgeToUnknownNode(t *testing.T) {
	g := graph.New[emptyState]()
	g.AddNode("a", noopNode)
	g.AddEdge("a", "missing")
	g.SetEntryPoint("a")

	_, err := g.Compile()
	if !errors.Is(err, graph.ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestGraphAddNodeRejectsDuplicateName(t *testing.T) {
	g := graph.New[emptyState]()
	if err := g.AddNode("a", noopNode); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err := g.AddNode("a", noopNode)
	if !errors.Is(err, graph.ErrNodeExists) {
		t.Fatalf("expected ErrNodeExists, got %v", err)
	}
}

func TestGraphRunPropagatesNodeError(t *testing.T) {
	wantErr := errors.New("boom")
	g := graph.New[emptyState]()
	g.AddNode("a", func(ctx context.Context, s emptyState) (emptyState, error) { return emptyState{}, wantErr })
	g.SetEntryPoint("a")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	_, err = runnable.Run(context.Background(), emptyState{})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped node error, got %v", err)
	}
}
