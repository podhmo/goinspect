package graph

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestGraphWalk(t *testing.T) {
	type ref struct{ Xs []int }

	cases := []struct {
		msg  string
		g    *Graph[int, int]
		want []int
	}{
		{msg: "walk ε -> ", g: Ints(), want: nil},
		{msg: "walk 1 -> 1", g: Ints(1), want: []int{1}},
		{msg: "walk 1,2,3 -> 1,2,3", g: Ints(1, 2, 3), want: []int{1, 2, 3}},
		{msg: "walk 1,2,2,1,3 -> 1,2,3", g: Ints(1, 2, 3), want: []int{1, 2, 3}},
		{msg: "walk 1,{2,3,4},{5,6} -> 1,2,3,4,5,6", want: []int{1, 2, 3, 4, 5, 6},
			g: func() *Graph[int, int] {
				// 1,
				g := Ints(1)

				// 2 -> 3 -> 4
				n2, _ := g.Add(2)
				n3, _ := g.Add(3)
				g.LinkTo(n2, n3)
				n4, _ := g.Add(4)
				g.LinkTo(n3, n4)

				// 5 -> 6
				n5, _ := g.Add(5)
				n6, _ := g.Add(6)
				g.LinkTo(n5, n6)
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
