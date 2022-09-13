package main

import (
	"fmt"
	"go/token"
	"log"
	"os"
	"strings"

	"github.com/podhmo/flagstruct"
	"github.com/podhmo/goinspect"
	"golang.org/x/tools/go/packages"
)

type Options struct {
	IncludeUnexported bool `flag:"include-unexported"`
	ExpandAll         bool `flag:"expand-all"`

	Pkg   string   `flag:"pkg" required:"true"`
	Other []string `flag:"other"`
	Only  []string `flag:"only"`
}

func main() {
	options := &Options{}
	flagstruct.Parse(options)

	if err := run(*options); err != nil {
		log.Fatalf("!! %+v", err)
	}
}

func run(options Options) error {
	fset := token.NewFileSet()

	c := &goinspect.Config{
		Fset:          fset,
		PkgPath:       options.Pkg,
		OtherPackages: options.Other,

		Padding:           "  ",
		IncludeUnexported: options.IncludeUnexported,
		ExpandAll:         options.ExpandAll,
	}

	if strings.HasSuffix(c.PkgPath, "...") {
		c.OtherPackages = append(c.OtherPackages, c.PkgPath)
		c.PkgPath = strings.TrimRight(c.PkgPath, "./")
	}

	cfg := &packages.Config{
		Fset: fset,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedTypes | packages.NeedTypesSizes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedDeps,
	}
	pkgs, err := packages.Load(cfg, append([]string{c.PkgPath}, c.OtherPackages...)...)
	if err != nil {
		return fmt.Errorf("load packages: %w", err)
	}

	// TODO: support e.g. ../../....
	if strings.HasPrefix(c.PkgPath, ".") {
		suffix := strings.TrimRight(strings.TrimLeft(c.PkgPath, "."), "/")
		for _, pkg := range pkgs {
			if strings.HasSuffix(pkg.PkgPath, suffix) {
				c.PkgPath = pkg.PkgPath // to fullpath
				break
			}
		}
	}

	g, err := goinspect.Scan(c, pkgs)
	if err != nil {
		return fmt.Errorf("scan: %w", err)
	}

	if len(options.Only) == 0 {
		if err := goinspect.DumpAll(os.Stdout, c, g); err != nil {
			return fmt.Errorf("dump: %w", err)
		}
		return nil
	}

	var nodes []*goinspect.Node
	g.Walk(func(n *goinspect.Node) {
		if n.Value.Recv == "" {
			for _, fullname := range options.Only {
				if fullname == n.Name {
					nodes = append(nodes, n)
					break
				}
			}
		} else {
			for _, fullname := range options.Only {
				if fullname == n.Value.Recv+"."+n.Name {
					nodes = append(nodes, n)
					break
				}
			}
		}
	})
	if err := goinspect.Dump(os.Stdout, c, g, nodes); err != nil {
		return fmt.Errorf("dump: %w", err)
	}
	return nil
}
