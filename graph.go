package goinspect

func NewGraph[K comparable, T any](keyFunc func(T) K, values ...T) *Graph[K, T] {
	g := &Graph[K, T]{
		KeyFunc: keyFunc,
		seen:    make(map[K]*Node[T], len(values)),
	}

	for _, v := range values {
		g.Add(v)
	}
	return g
}

type Graph[K comparable, T any] struct {
	Nodes   []*Node[T]
	KeyFunc func(T) K

	seen map[K]*Node[T]
}

func (g *Graph[K, T]) Add(v T) (added bool) {
	k := g.KeyFunc(v)
	_, ok := g.seen[k]
	if ok {
		return false
	}
	node := &Node[T]{ID: len(g.Nodes), Value: v}
	g.seen[k] = node
	g.Nodes = append(g.Nodes, node)
	return true
}

type Node[T any] struct {
	ID    int
	Value T

	From []*Node[T]
	To   []*Node[T]
}
