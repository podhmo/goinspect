package goinspect

import (
	"go/token"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	pkg := "github.com/podhmo/goinspect/internal/x"
	fset := token.NewFileSet()
	cfg := &Config{
		Fset:    fset,
		PkgPath: pkg,
		OtherPackages: []string{
			"github.com/podhmo/goinspect/internal/x/sub",
		},

		IncludeUnexported: true,
	}

	g, err := Scan(cfg)
	if err != nil {
		t.Errorf("unexpected error: %+v", err)
	}

	var nodes []*Node
	g.Walk(func(n *Node) {
		if strings.HasPrefix(n.Name, "F") {
			nodes = append(nodes, n)
		}
	})
	if err := Dump(cfg, g, nodes); err != nil {
		t.Errorf("unexpected error: %+v", err)
	}
}
