package goinspect

import "fmt"

func IntGraph(values ...int) *Graph[int, int] {
	return NewGraph(func(v int) int { return v }, values...)
}
func StringGraph(values ...string) *Graph[string, string] {
	return NewGraph(func(v string) string { return v }, values...)
}

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
	c    int
}

func (g *Graph[K, T]) Add(v T) (node *Node[T], added bool) {
	k := g.KeyFunc(v)
	node, ok := g.seen[k]
	if ok {
		return node, false
	}
	g.c++
	node = &Node[T]{ID: g.c, Value: v}
	g.seen[k] = node
	g.Nodes = append(g.Nodes, node)
	return node, true
}

func (g *Graph[K, T]) LinkTo(node *Node[T], v T) (child *Node[T], added bool) {
	k := g.KeyFunc(v)
	child, ok := g.seen[k]
	if !ok {
		g.c++
		child = &Node[T]{ID: g.c, Value: v}
		g.seen[k] = child
	}

	for _, x := range node.To {
		if x.ID == child.ID {
			return child, false
		}
	}
	node.To = append(node.To, child)
	child.From = append(child.From, node)
	return child, true
}

func (g *Graph[K, T]) Walk(fn func(*Node[T])) {
	seen := map[int]struct{}{}
	for _, n := range g.Nodes {
		if _, ok := seen[n.ID]; ok {
			continue
		}
		seen[n.ID] = struct{}{}
		fn(n)

		if len(n.To) > 0 {
			q := n.To[:]
			for len(q) > 0 {
				n, q = q[0], q[1:]
				if _, ok := seen[n.ID]; ok {
					continue
				}
				seen[n.ID] = struct{}{}
				fn(n)
				if len(n.To) > 0 {
					q = append(q, n.To...)
				}
			}
		}
	}
}

type key struct {
	prev int
	next int
}

func (g *Graph[K, T]) WalkPath(fn func([]*Node[T])) {
	seen := map[key]struct{}{}
	for _, n := range g.Nodes {
		k := key{prev: n.ID}
		if _, ok := seen[k]; ok {
			continue
		}
		seen[k] = struct{}{}

		fn([]*Node[T]{n})
		if len(n.To) > 0 {
			q := make([][]*Node[T], 0, len(n.To))
			for _, next := range n.To {
				q = append(q, []*Node[T]{n, next})
			}

			var path []*Node[T]
			for len(q) > 0 {
				path, q = q[0], q[1:]
				n, next := path[len(path)-2], path[len(path)-1]
				k := key{prev: n.ID, next: next.ID}
				if _, ok := seen[k]; ok {
					continue
				}
				seen[k] = struct{}{}

				{
					k := key{prev: next.ID}
					if _, ok := seen[k]; !ok {
						fn([]*Node[T]{next})
						seen[k] = struct{}{}
					}
				}

				fn(path)
				if len(next.To) > 0 {
					nq := make([][]*Node[T], len(next.To))
					for i, nextnext := range next.To {
						nq[i] = append(path[:], nextnext)
					}
					q = append(nq, q...) // hack: stack like
				}
			}
		}
	}
}

type Node[T any] struct {
	ID    int
	Name  string
	Value T

	From []*Node[T]
	To   []*Node[T]
}

func (n *Node[T]) String() string {
	return fmt.Sprintf("N<%v>", n.Value)
}
