package goinspect

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

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
		{msg: "walk 1,{2,3,4},{5,6} -> 1,2,3,4,5,6", want: []int{1, 2, 3, 4, 5, 6},
			g: func() *Graph[int, int] {
				// 1,
				g := newGraph(1)

				// 2 -> 3 -> 4
				n2, _ := g.Add(2)
				n3, _ := g.LinkTo(n2, 3)
				g.LinkTo(n3, 4)

				// 5 -> 6
				n5, _ := g.Add(5)
				g.LinkTo(n5, 6)
				return g
			}(),
		},
	}

	for _, c := range cases {
		t.Run(c.msg, func(t *testing.T) {
			g := c.g
			want := c.want

			var got []int
			g.Walk(func(n *Node[int]) {
				got = append(got, n.Value)
			})
			if diff := cmp.Diff(ref{want}, ref{got}); diff != "" {
				t.Errorf("Walk() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
