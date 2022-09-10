package main

import (
	"fmt"
	"go/token"
	"log"

	"github.com/podhmo/flagstruct"
	"github.com/podhmo/goinspect"
	"golang.org/x/tools/go/packages"
)

type Options struct {
	IncludeUnexported bool     `flag:"include-unexported"`
	Pkg               string   `flag:"pkg" required:"true"`
	Other             []string `flag:"other"`
	Only              []string `flag:"only"`
}

func main() {
	options := &Options{}
	flagstruct.Parse(options)

	if err := run(*options); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func run(options Options) error {
	pkg := options.Pkg
	fset := token.NewFileSet()

	c := &goinspect.Config{
		Fset:              fset,
		PkgPath:           pkg,
		OtherPackages:     options.Other,
		IncludeUnexported: options.IncludeUnexported,
	}

	cfg := &packages.Config{
		Fset: fset,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedTypes | packages.NeedTypesSizes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedDeps,
	}
	pkgs, err := packages.Load(cfg, append([]string{c.PkgPath}, c.OtherPackages...)...)
	if err != nil {
		return fmt.Errorf("load packages: %w", err)
	}

	g, err := goinspect.Scan(c, pkgs)
	if err != nil {
		return fmt.Errorf("scan: %w", err)
	}

	if len(options.Only) == 0 {
		if err := goinspect.DumpAll(c, g); err != nil {
			return fmt.Errorf("dump: %w", err)
		}
	}

	var nodes []*goinspect.Node
	g.Walk(func(n *goinspect.Node) {
		for _, name := range options.Only {
			if name == n.Name {
				nodes = append(nodes, n)
				break
			}
		}
	})
	if err := goinspect.Dump(c, g, nodes); err != nil {
		return fmt.Errorf("dump: %w", err)
	}
	return nil
}
