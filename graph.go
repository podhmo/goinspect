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
	return child, true
}

func (g *Graph[K, T]) LinkFrom(node *Node[T], v T) (child *Node[T], added bool) {
	k := g.KeyFunc(v)
	child, ok := g.seen[k]
	if !ok {
		g.c++
		child = &Node[T]{ID: g.c, Value: v}
		g.seen[k] = child
	}

	for _, x := range node.From {
		if x.ID == child.ID {
			return child, false
		}
	}
	node.From = append(node.From, child)
	return child, true
}

func (g *Graph[K, T]) LinkBoth(node *Node[T], v T) (child *Node[T], added bool) {
	k := g.KeyFunc(v)
	child, ok := g.seen[k]
	if !ok {
		g.c++
		child = &Node[T]{ID: g.c, Value: v}
		g.seen[k] = child
	}

	toAdded := false
	for _, x := range node.To {
		if x.ID == child.ID {
			toAdded = true
			break
		}
	}
	node.To = append(node.To, child)

	fromAdded := false
	for _, x := range node.From {
		if x.ID == child.ID {
			fromAdded = true
			break
		}
	}
	node.From = append(node.From, child)
	return child, toAdded || fromAdded
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

type Node[T any] struct {
	ID    int
	Value T

	From []*Node[T]
	To   []*Node[T]
}
