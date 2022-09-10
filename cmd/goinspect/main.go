package main

import (
	"fmt"
	"go/token"
	"log"

	"github.com/podhmo/flagstruct"
	"github.com/podhmo/goinspect"
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

	cfg := &goinspect.Config{
		Fset:              fset,
		PkgPath:           pkg,
		OtherPackages:     options.Other,
		IncludeUnexported: options.IncludeUnexported,
	}

	g, err := goinspect.Scan(cfg)
	if err != nil {
		return fmt.Errorf("scan: %w", err)
	}

	if len(options.Only) == 0 {
		if err := goinspect.DumpAll(cfg, g); err != nil {
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
	if err := goinspect.Dump(cfg, g, nodes); err != nil {
		return fmt.Errorf("dump: %w", err)
	}
	return nil
}
