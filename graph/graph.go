package graph

import (
	"fmt"
	"os"
)

func Ints(values ...int) *Graph[int, int] {
	return New(func(v int) int { return v }, values...)
}
func Strings(values ...string) *Graph[string, string] {
	return New(func(v string) string { return v }, values...)
}

func New[K comparable, T any](keyFunc func(T) K, values ...T) *Graph[K, T] {
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

func (g *Graph[K, T]) Madd(v T) *Node[T] {
	node, _ := g.Add(v)
	return node
}

func (g *Graph[K, T]) LinkTo(prev *Node[T], node *Node[T]) (added bool) {
	for _, x := range prev.To {
		if x.ID == node.ID {
			return false
		}
	}
	prev.To = append(prev.To, node)
	node.From = append(node.From, prev)
	return true
}

func (g *Graph[K, T]) Walk(fn func(*Node[T])) {
	for _, n := range g.Nodes {
		fn(n)
	}
}

type key struct {
	prev int
	next int
}

// topological sort
func (g *Graph[K, T]) SortedByFrom(nodes []*Node[T]) []*Node[T] {
	if nodes == nil {
		nodes = g.Nodes
	}
	only := make(map[int]bool, len(nodes))
	for _, n := range g.Nodes {
		only[n.ID] = false
	}
	for _, n := range nodes {
		only[n.ID] = true
	}

	r := make([]*Node[T], 0, len(nodes))
	seen := make(map[int]struct{}, len(nodes))
	var walk func(n *Node[T])
	walk = func(n *Node[T]) {
		if _, ok := seen[n.ID]; ok {
			return
		}
		seen[n.ID] = struct{}{}
		if ok := only[n.ID]; !ok {
			return
		}

		for _, next := range n.From {
			walk(next)
		}
		r = append(r, n)
	}
	for _, n := range nodes {
		walk(n)
	}
	return r
}

// topological sort
func (g *Graph[K, T]) SortedByTo(nodes []*Node[T]) []*Node[T] {
	if nodes == nil {
		nodes = g.Nodes
	}
	only := make(map[int]bool, len(nodes))
	for _, n := range g.Nodes {
		only[n.ID] = false
	}
	for _, n := range nodes {
		only[n.ID] = true
	}

	r := make([]*Node[T], 0, len(nodes))
	seen := make(map[int]struct{}, len(nodes))
	var walk func(n *Node[T])
	walk = func(n *Node[T]) {
		if _, ok := seen[n.ID]; ok {
			return
		}
		seen[n.ID] = struct{}{}
		if ok := only[n.ID]; !ok {
			return
		}

		for _, next := range n.To {
			walk(next)
		}
		r = append(r, n)
	}
	for _, n := range nodes {
		walk(n)
	}
	return r
}

func (g *Graph[K, T]) WalkPath(fn func([]*Node[T]), nodes []*Node[T]) {
	if nodes == nil {
		nodes = g.Nodes
	}

	seen := map[key]struct{}{}
	for _, n := range nodes {
		if len(n.From) > 0 {
			continue
		}
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
					// debugprint("SK", path)
					continue
				}

				{
					k := key{prev: next.ID}
					if _, ok := seen[k]; !ok {
						seen[k] = struct{}{}
						fn([]*Node[T]{next})
					}
				}

				seen[k] = struct{}{}
				// debugprint("<-", path)
				fn(path)

				if len(next.To) > 0 {
					nq := make([][]*Node[T], len(next.To))
					for i, nextnext := range next.To {
						copied := make([]*Node[T], len(path)+1)
						copy(copied, path)
						copied[len(copied)-1] = nextnext
						nq[i] = copied
					}
					q = append(nq, q...) // hack: stack like
				}
			}
		}
	}
}

func debugprint[T any](prefix string, path []*Node[T]) {
	{
		xs := make([]string, len(path))
		for i, x := range path {
			xs[i] = x.Name
		}
		fmt.Fprintf(os.Stderr, "\t\t\t\t## %s%v\n", prefix, xs)
	}
}

type Node[T any] struct {
	ID    int
	Name  string
	Value T

	From []*Node[T]
	To   []*Node[T]

	Metadata struct {
		Shape Shape
	}
}

func (n *Node[T]) String() string {
	return fmt.Sprintf("N<%v>", n.Value)
}
