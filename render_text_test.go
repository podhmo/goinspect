package goinspect

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
		{msg: "no-link", want: "1;\n2;\n3;", g: newGraph(1, 2, 3)},
		{msg: "link1", want: "1 -> 2;\n2 -> 3;",
			g: func() *Graph[int, int] {
				g := newGraph()
				n1, _ := g.Add(1)
				n2, _ := g.LinkTo(n1, 2)
				g.LinkTo(n2, 3)
				return g
			}(),
		},
		{msg: "link2", want: "1 -> 2;\n2 -> 3;\n2 -> 4;\n1 -> 5;",
			g: func() *Graph[int, int] {
				g := newGraph()
				n1, _ := g.Add(1)
				n2, _ := g.LinkTo(n1, 2)
				g.LinkTo(n2, 3)
				g.LinkTo(n2, 4)
				g.LinkTo(n1, 5)
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
