package graph

import (
	"bytes"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRenderText(t *testing.T) {
	cases := []struct {
		msg  string
		g    *Graph[int, int]
		want string
	}{
		{msg: "no-link", want: "1;\n2;\n3;", g: Ints(1, 2, 3)},
		{msg: "link1", want: "1;\n2;\n1 -> 2;\n3;\n2 -> 3;",
			g: func() *Graph[int, int] {
				g := Ints()
				n1 := g.Madd(1)
				n2 := g.Madd(2)
				g.LinkTo(n1, n2)
				n3 := g.Madd(3)
				g.LinkTo(n2, n3)
				return g
			}(),
		},
		{msg: "link2", want: "1;\n2;\n1 -> 2;\n3;\n2 -> 3;\n4;\n2 -> 4;\n5;\n1 -> 5;",
			g: func() *Graph[int, int] {
				g := Ints()
				n1 := g.Madd(1)
				n2 := g.Madd(2)
				g.LinkTo(n1, n2)
				n3 := g.Madd(3)
				g.LinkTo(n2, n3)
				n4 := g.Madd(4)
				g.LinkTo(n2, n4)
				n5 := g.Madd(5)
				g.LinkTo(n1, n5)
				return g
			}(),
		},
		{msg: "link3", want: "1;\n3;\n1 -> 3;\n4;\n1 -> 4;\n2;\n2 -> 3;\n2 -> 4;",
			g: func() *Graph[int, int] {
				g := Ints()

				n1 := g.Madd(1)
				n2 := g.Madd(2)
				n3 := g.Madd(3)
				n4 := g.Madd(4)

				g.LinkTo(n1, n3)
				g.LinkTo(n1, n4)

				g.LinkTo(n2, n3)
				g.LinkTo(n2, n4)
				return g
			}(),
		},
		{msg: "rec", want: "1;\n2;\n1 -> 2;\n2 -> 2;",
			g: func() *Graph[int, int] {
				g := Ints()

				n1 := g.Madd(1)
				n2 := g.Madd(2)

				g.LinkTo(n1, n2)
				g.LinkTo(n2, n2)
				return g
			}(),
		},
	}

	for _, c := range cases {
		t.Run(c.msg, func(t *testing.T) {
			g := c.g
			buf := new(bytes.Buffer)
			if err := RenderText(buf, g); err != nil {
				t.Errorf("RenderText(), unexpected error: %+v", err)
			}

			got := strings.TrimSpace(buf.String())
			want := c.want
			if diff := cmp.Diff(want, got); diff != "" {
				t.Errorf("RenderText() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
