package goinspect

import (
	"fmt"
	"go/token"
	"log"
	"strings"
	"testing"

	"github.com/podhmo/goinspect/graph"
	"golang.org/x/tools/go/packages"
)

func TestParse(t *testing.T) {
	fset := token.NewFileSet()
	pkg := "github.com/podhmo/goinspect/internal/x"
	if err := parse(fset, pkg); err != nil {
		t.Errorf("unexpected error: %+v", err)
	}
}

func parse(fset *token.FileSet, pkgpath string) error {
	cfg := &packages.Config{
		Fset: fset,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedTypes | packages.NeedTypesSizes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedDeps,
	}
	pkgs, err := packages.Load(cfg, pkgpath, "github.com/podhmo/goinspect/internal/x/sub")
	if err != nil {
		return err
	}

	g := graph.New(func(s *Subject) string { return s.ID })
	pkgMap := make(map[string]*packages.Package, len(pkgs))
	for _, pkg := range pkgs {
		pkgMap[pkg.ID] = pkg
	}
	scanner := &Scanner{
		g:      g,
		pkgMap: pkgMap,
	}
	scanner.Config.IncludeUnexported = true

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return pkg.Errors[0] // TODO: multierror
		}
		for _, f := range pkg.Syntax {
			if err := scanner.Scan(pkg, f); err != nil {
				log.Printf("! %+v", err)
			}
		}
		fmt.Println("**", pkg.ID)
	}

	fmt.Printf("package %s\n", pkgpath)
	g.WalkPath(func(path []*Node) {
		parts := strings.Split(pkgpath, "/")
		prefix := strings.Join(parts[:len(parts)-1], "/") + "/"
		if len(path) == 1 {
			node := path[0]
			if len(node.From) == 0 && scanner.Need(node.Name) {
				name := strings.ReplaceAll(path[len(path)-1].Value.Object.String(), prefix, "")
				fmt.Printf("%s%s\n", strings.Repeat("  ", len(path)), name)
			}
		} else {
			node := path[len(path)-1]
			if scanner.Need(node.Name) {
				name := strings.ReplaceAll(node.Value.Object.String(), prefix, "")
				fmt.Printf("%s%s\n", strings.Repeat("  ", len(path)), name)
			}
		}
	})

	return nil
}
