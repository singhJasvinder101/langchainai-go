package graph_test

import (
	"context"
	"errors"
	"testing"

	"github.com/singhJasvinder101/agentic-go/graph"
)

func TestGraphRunsStaticEdgesToEnd(t *testing.T) {
	g := graph.New()
	g.AddNode("a", func(ctx context.Context, s graph.State) (graph.State, error) {
		s["path"] = append(s["path"].([]string), "a")
		return s, nil
	})
	g.AddNode("b", func(ctx context.Context, s graph.State) (graph.State, error) {
		s["path"] = append(s["path"].([]string), "b")
		return s, nil
	})
	g.AddEdge("a", "b")
	g.AddEdge("b", graph.End)
	g.SetEntryPoint("a")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	final, err := runnable.Run(context.Background(), graph.State{"path": []string{}})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	path := final["path"].([]string)
	if len(path) != 2 || path[0] != "a" || path[1] != "b" {
		t.Fatalf("expected path [a b], got %v", path)
	}
}

func TestGraphNodeWithNoOutgoingEdgeImplicitlyEnds(t *testing.T) {
	g := graph.New()
	g.AddNode("only", func(ctx context.Context, s graph.State) (graph.State, error) {
		s["done"] = true
		return s, nil
	})
	g.SetEntryPoint("only")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	final, err := runnable.Run(context.Background(), nil)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if final["done"] != true {
		t.Fatalf("expected done=true, got %+v", final)
	}
}

func TestGraphConditionalRouting(t *testing.T) {
	g := graph.New()
	g.AddNode("classify", func(ctx context.Context, s graph.State) (graph.State, error) {
		return s, nil
	})
	g.AddNode("positive", func(ctx context.Context, s graph.State) (graph.State, error) {
		s["result"] = "positive branch"
		return s, nil
	})
	g.AddNode("negative", func(ctx context.Context, s graph.State) (graph.State, error) {
		s["result"] = "negative branch"
		return s, nil
	})
	g.AddConditionalEdge("classify", func(s graph.State) string {
		if s["score"].(int) > 0 {
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

	final, err := runnable.Run(context.Background(), graph.State{"score": 5})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if final["result"] != "positive branch" {
		t.Fatalf("expected positive branch, got %+v", final)
	}

	final, err = runnable.Run(context.Background(), graph.State{"score": -1})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if final["result"] != "negative branch" {
		t.Fatalf("expected negative branch, got %+v", final)
	}
}

func TestGraphCycleGuardedByMaxSteps(t *testing.T) {
	g := graph.New()
	g.AddNode("loop", func(ctx context.Context, s graph.State) (graph.State, error) {
		s["count"] = s["count"].(int) + 1
		return s, nil
	})
	g.AddEdge("loop", "loop")
	g.SetEntryPoint("loop")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("compile: %v", err)
	}

	_, err = runnable.RunWithLimit(context.Background(), graph.State{"count": 0}, 5)
	if err == nil {
		t.Fatal("expected error from unterminated cycle")
	}
}

func TestGraphCompileRequiresEntryPoint(t *testing.T) {
	g := graph.New()
	g.AddNode("a", func(ctx context.Context, s graph.State) (graph.State, error) { return s, nil })

	_, err := g.Compile()
	if !errors.Is(err, graph.ErrEntryRequired) {
		t.Fatalf("expected ErrEntryRequired, got %v", err)
	}
}

func TestGraphCompileRejectsUnknownEntryPoint(t *testing.T) {
	g := graph.New()
	g.AddNode("a", func(ctx context.Context, s graph.State) (graph.State, error) { return s, nil })
	g.SetEntryPoint("missing")

	_, err := g.Compile()
	if !errors.Is(err, graph.ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestGraphCompileRejectsEdgeToUnknownNode(t *testing.T) {
	g := graph.New()
	g.AddNode("a", func(ctx context.Context, s graph.State) (graph.State, error) { return s, nil })
	g.AddEdge("a", "missing")
	g.SetEntryPoint("a")

	_, err := g.Compile()
	if !errors.Is(err, graph.ErrNodeNotFound) {
		t.Fatalf("expected ErrNodeNotFound, got %v", err)
	}
}

func TestGraphAddNodeRejectsDuplicateName(t *testing.T) {
	g := graph.New()
	if err := g.AddNode("a", func(ctx context.Context, s graph.State) (graph.State, error) { return s, nil }); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err := g.AddNode("a", func(ctx context.Context, s graph.State) (graph.State, error) { return s, nil })
	if !errors.Is(err, graph.ErrNodeExists) {
		t.Fatalf("expected ErrNodeExists, got %v", err)
	}
}

func TestGraphRunPropagatesNodeError(t *testing.T) {
	wantErr := errors.New("boom")
	g := graph.New()
	g.AddNode("a", func(ctx context.Context, s graph.State) (graph.State, error) { return nil, wantErr })
	g.SetEntryPoint("a")

	runnable, err := g.Compile()
	if err != nil {
		t.Fatalf("compile: %v", err)
	}
	_, err = runnable.Run(context.Background(), nil)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped node error, got %v", err)
	}
}
