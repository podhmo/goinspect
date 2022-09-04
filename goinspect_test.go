package goinspect

import (
	"fmt"
	"go/token"
	"testing"

	"golang.org/x/tools/go/packages"
)

func TestParse(t *testing.T) {
	if err := parse("github.com/podhmo/goinspect/internal/x"); err != nil {
		t.Errorf("unexpected error: %+v", err)
	}
}

func parse(path string) error {
	fset := token.NewFileSet()
	cfg := &packages.Config{
		Fset: fset,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedTypes | packages.NeedTypesSizes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedDeps,
	}
	pkgs, err := packages.Load(cfg, path)
	if err != nil {
		return err
	}
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return pkg.Errors[0] // TODO: multierror
		}
		fmt.Println(pkg)
	}
	return nil
}
