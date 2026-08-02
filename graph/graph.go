// Package graph is a minimal, general-purpose directed state-machine
// executor — the LangGraph-equivalent piece of this library. Nodes are plain
// functions over a caller-defined state type S; edges (static or
// conditional) decide what runs next, including cycles. Graph has no
// dependency on llm/agent/chain — node functions are free to call into those
// packages themselves.
//
// S is generic and checked at compile time — fields are read/written as
// s.Field, not s["field"].(T), so a typo or wrong type is a build error
// instead of a runtime panic.
package graph

import (
	"context"
	"fmt"
)

// NodeFunc executes one step and returns the (possibly modified) state.
type NodeFunc[S any] func(ctx context.Context, state S) (S, error)

// Router decides the next node given the current state, for conditional
// edges. Its return value is looked up in the routes map passed to
// AddConditionalEdge.
type Router[S any] func(state S) string

// End is the sentinel destination that stops execution.
const End = "__end__"

type edge[S any] struct {
	router Router[S]
	routes map[string]string // conditional edges: router output -> node name (or End)
	to     string            // static edges: destination node name (or End)
}

// Graph is a builder for a directed node graph over state type S. Build it
// with AddNode/AddEdge/AddConditionalEdge/SetEntryPoint, then Compile it into
// a Runnable.
type Graph[S any] struct {
	nodes map[string]NodeFunc[S]
	edges map[string]edge[S]
	entry string
}

// New builds an empty Graph over state type S.
//
//	type SupportState struct {
//		Input, Category, Result string
//	}
//	g := graph.New[SupportState]()
func New[S any]() *Graph[S] {
	return &Graph[S]{nodes: make(map[string]NodeFunc[S]), edges: make(map[string]edge[S])}
}

// AddNode registers a node under name.
func (g *Graph[S]) AddNode(name string, fn NodeFunc[S]) error {
	if _, ok := g.nodes[name]; ok {
		return fmt.Errorf("%w: %s", ErrNodeExists, name)
	}
	g.nodes[name] = fn
	return nil
}

// SetEntryPoint sets the node execution starts from.
func (g *Graph[S]) SetEntryPoint(name string) {
	g.entry = name
}

// AddEdge connects from -> to unconditionally. to may be End.
func (g *Graph[S]) AddEdge(from, to string) {
	g.edges[from] = edge[S]{to: to}
}

// AddConditionalEdge routes from `from` to whichever node router's return
// value maps to in routes. Map a router output to End to terminate there.
func (g *Graph[S]) AddConditionalEdge(from string, router Router[S], routes map[string]string) {
	g.edges[from] = edge[S]{router: router, routes: routes}
}

// Compile validates the graph (entry point set and exists, every edge target
// exists or is End) and returns an executable Runnable.
func (g *Graph[S]) Compile() (*Runnable[S], error) {
	if g.entry == "" {
		return nil, ErrEntryRequired
	}
	if _, ok := g.nodes[g.entry]; !ok {
		return nil, fmt.Errorf("%w: entry point %q", ErrNodeNotFound, g.entry)
	}
	for from, e := range g.edges {
		if e.router == nil {
			if e.to != End {
				if _, ok := g.nodes[e.to]; !ok {
					return nil, fmt.Errorf("%w: edge %s -> %s", ErrNodeNotFound, from, e.to)
				}
			}
			continue
		}
		for out, to := range e.routes {
			if to == End {
				continue
			}
			if _, ok := g.nodes[to]; !ok {
				return nil, fmt.Errorf("%w: conditional edge %s -[%s]-> %s", ErrNodeNotFound, from, out, to)
			}
		}
	}

	nodes := make(map[string]NodeFunc[S], len(g.nodes))
	for k, v := range g.nodes {
		nodes[k] = v
	}
	edges := make(map[string]edge[S], len(g.edges))
	for k, v := range g.edges {
		edges[k] = v
	}
	return &Runnable[S]{nodes: nodes, edges: edges, entry: g.entry}, nil
}

// defaultMaxSteps guards Run against unterminated cycles.
const defaultMaxSteps = 100

// Runnable is a compiled, executable Graph.
type Runnable[S any] struct {
	nodes map[string]NodeFunc[S]
	edges map[string]edge[S]
	entry string
}

// Run executes the graph from its entry point until a node has no outgoing
// edge or a router resolves to End, guarding against unterminated cycles with
// a default step limit. See RunWithLimit to set an explicit limit.
func (r *Runnable[S]) Run(ctx context.Context, initial S) (S, error) {
	return r.RunWithLimit(ctx, initial, defaultMaxSteps)
}

// RunWithLimit is Run with an explicit maximum number of node executions.
func (r *Runnable[S]) RunWithLimit(ctx context.Context, initial S, maxSteps int) (S, error) {
	state := initial

	current := r.entry
	for step := 0; step < maxSteps; step++ {
		fn, ok := r.nodes[current]
		if !ok {
			return state, fmt.Errorf("%w: %s", ErrNodeNotFound, current)
		}

		next, err := fn(ctx, state)
		if err != nil {
			return state, fmt.Errorf("graph: node %q: %w", current, err)
		}
		state = next

		e, ok := r.edges[current]
		if !ok {
			return state, nil // no outgoing edge: implicit end
		}

		if e.router != nil {
			key := e.router(state)
			to, ok := e.routes[key]
			if !ok {
				return state, fmt.Errorf("graph: conditional edge from %q: no route for %q", current, key)
			}
			if to == End {
				return state, nil
			}
			current = to
			continue
		}

		if e.to == End {
			return state, nil
		}
		current = e.to
	}
	return state, fmt.Errorf("graph: exceeded %d steps (possible cycle without termination)", maxSteps)
}
