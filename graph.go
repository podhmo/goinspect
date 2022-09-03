package goinspect

func NewGraph[T any](values ...T) *Graph[T] {
	nodes := make([]*Node[T], len(values))
	for i, v := range values {
		nodes[i] = &Node[T]{ID: i, Value: v}
	}
	return &Graph[T]{
		Nodes: nodes,
	}
}

type Graph[T any] struct {
	Nodes []*Node[T]
}

type Node[T any] struct {
	ID    int
	Value T

	From []*Node[T]
	To   []*Node[T]
}
