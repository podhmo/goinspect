package goinspect

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Walk[K comparable, T any](g *Graph[K, T], fn func(*Node[T])) {
	for _, n := range g.Nodes {
		fn(n)
	}
}

func newGraph(values ...int) *Graph[int, int] {
	return NewGraph(func(v int) int { return v }, values...)
}

func TestGraphWalk(t *testing.T) {
	type ref struct{ Xs []int }

	cases := []struct {
		msg  string
		g    *Graph[int, int]
		want []int
	}{
		{msg: "walk ε -> ", g: newGraph(), want: nil},
		{msg: "walk 1 -> 1", g: newGraph(1), want: []int{1}},
		{msg: "walk 1,2,3 -> 1,2,3", g: newGraph(1, 2, 3), want: []int{1, 2, 3}},
		{msg: "walk 1,2,2,1,3 -> 1,2,3", g: newGraph(1, 2, 3), want: []int{1, 2, 3}},
	}

	for _, c := range cases {
		t.Run(c.msg, func(t *testing.T) {
			g := c.g
			want := c.want

			var got []int
			Walk(g, func(n *Node[int]) {
				got = append(got, n.Value)
			})
			if diff := cmp.Diff(ref{want}, ref{got}); diff != "" {
				t.Errorf("Walk() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
