// Package graph is a minimal, general-purpose directed state-machine
// executor — the LangGraph-equivalent piece of this library. Nodes are plain
// functions over a shared State; edges (static or conditional) decide what
// runs next, including cycles. Graph has no dependency on llm/agent/chain —
// node functions are free to call into those packages themselves.
package graph

import (
	"context"
	"fmt"
)

// State is the mutable data passed between nodes.
type State map[string]any

// NodeFunc executes one step and returns the (possibly modified) state.
type NodeFunc func(ctx context.Context, state State) (State, error)

// Router decides the next node given the current state, for conditional
// edges. Its return value is looked up in the routes map passed to
// AddConditionalEdge.
type Router func(state State) string

// End is the sentinel destination that stops execution.
const End = "__end__"

type edge struct {
	router Router
	routes map[string]string // conditional edges: router output -> node name (or End)
	to     string            // static edges: destination node name (or End)
}

// Graph is a builder for a directed node graph over a shared State. Build it
// with AddNode/AddEdge/AddConditionalEdge/SetEntryPoint, then Compile it into
// a Runnable.
type Graph struct {
	nodes map[string]NodeFunc
	edges map[string]edge
	entry string
}

// New builds an empty Graph.
func New() *Graph {
	return &Graph{nodes: make(map[string]NodeFunc), edges: make(map[string]edge)}
}

// AddNode registers a node under name.
func (g *Graph) AddNode(name string, fn NodeFunc) error {
	if _, ok := g.nodes[name]; ok {
		return fmt.Errorf("%w: %s", ErrNodeExists, name)
	}
	g.nodes[name] = fn
	return nil
}

// SetEntryPoint sets the node execution starts from.
func (g *Graph) SetEntryPoint(name string) {
	g.entry = name
}

// AddEdge connects from -> to unconditionally. to may be End.
func (g *Graph) AddEdge(from, to string) {
	g.edges[from] = edge{to: to}
}

// AddConditionalEdge routes from `from` to whichever node router's return
// value maps to in routes. Map a router output to End to terminate there.
func (g *Graph) AddConditionalEdge(from string, router Router, routes map[string]string) {
	g.edges[from] = edge{router: router, routes: routes}
}

// Compile validates the graph (entry point set and exists, every edge target
// exists or is End) and returns an executable Runnable.
func (g *Graph) Compile() (*Runnable, error) {
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

	nodes := make(map[string]NodeFunc, len(g.nodes))
	for k, v := range g.nodes {
		nodes[k] = v
	}
	edges := make(map[string]edge, len(g.edges))
	for k, v := range g.edges {
		edges[k] = v
	}
	return &Runnable{nodes: nodes, edges: edges, entry: g.entry}, nil
}

// defaultMaxSteps guards Run against unterminated cycles.
const defaultMaxSteps = 100

// Runnable is a compiled, executable Graph.
type Runnable struct {
	nodes map[string]NodeFunc
	edges map[string]edge
	entry string
}

// Run executes the graph from its entry point until a node has no outgoing
// edge or a router resolves to End, guarding against unterminated cycles with
// a default step limit. See RunWithLimit to set an explicit limit.
func (r *Runnable) Run(ctx context.Context, initial State) (State, error) {
	return r.RunWithLimit(ctx, initial, defaultMaxSteps)
}

// RunWithLimit is Run with an explicit maximum number of node executions.
func (r *Runnable) RunWithLimit(ctx context.Context, initial State, maxSteps int) (State, error) {
	state := initial
	if state == nil {
		state = State{}
	}

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
