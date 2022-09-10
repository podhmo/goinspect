package goinspect

import (
	"go/token"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestParse(t *testing.T) {
	pkg := "github.com/podhmo/goinspect/internal/x"
	fset := token.NewFileSet()
	c := &Config{
		Fset:    fset,
		PkgPath: pkg,
		OtherPackages: []string{
			"github.com/podhmo/goinspect/internal/x/sub",
		},

		IncludeUnexported: true,
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
	if err := DumpAll(c, g); err != nil {
		t.Errorf("unexpected error: %+v", err)
	}
}
