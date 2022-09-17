package goinspect

import (
	"bytes"
	"go/token"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"golang.org/x/tools/go/packages"
)

func TestIT(t *testing.T) {
	pkg := "github.com/podhmo/goinspect/internal/x"
	fset := token.NewFileSet()
	c := &Config{
		Fset:    fset,
		PkgPath: pkg,
		OtherPackages: []string{
			"github.com/podhmo/goinspect/internal/x/sub",
		},
		Padding:           "@",
		IncludeUnexported: true,
		skipHeader:        true,
	}

	cfg := &packages.Config{
		Fset: fset,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedTypes | packages.NeedTypesSizes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedDeps,
	}
	pkgs, err := packages.Load(cfg, append([]string{c.PkgPath}, c.OtherPackages...)...)
	if err != nil {
		t.Errorf("unexpected error: %+v", err)
	}

	g, err := Scan(c, pkgs)
	if err != nil {
		t.Errorf("unexpected error: %+v", err)
	}

	testcases := []struct {
		msg   string
		want  string
		names []string
	}{
		{
			msg: "F", names: []string{"F"},
			want: `
@func x.F(s x.S)
@@func x.log() func()  // &3
@@func x.F0()
@@@func x.log() func()  // *3
@@@func x.F1()
@@@@func x.log() func()  // *3
@@@@func x.H()  // &5
@@func x.H()  // *5`,
		},
		{
			msg: "G", names: []string{"G"},
			want: `
@func x.G()
@@func x.log() func()  // &3
@@func x.G0()
@@@func x.log() func()  // *3
@@@func x.H()
@@func x/sub.X()`,
			},
	}

	for _, tc := range testcases {
		t.Run(tc.msg, func(t *testing.T) {
			buf := new(bytes.Buffer)
			var nodes []*Node
			g.Walk(func(n *Node) {
				for _, name := range tc.names {
					if name == n.Name {
						nodes = append(nodes, n)
					}
				}
			})

			if err := Dump(buf, c, g, nodes); err != nil {
				t.Errorf("unexpected error: %+v", err)
			}

			if diff := cmp.Diff(strings.TrimSpace(tc.want), strings.TrimSpace(buf.String())); diff != "" {
				t.Errorf("Scan() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
