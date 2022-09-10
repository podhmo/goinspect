package goinspect

import (
	"fmt"
	"go/token"
	"log"
	"strings"

	"github.com/podhmo/goinspect/graph"
	"golang.org/x/tools/go/packages"
)

type Config struct {
	Fset *token.FileSet

	PkgPath string

	IncludeUnexported bool
	OtherPackages     []string
}

func (c *Config) NeedName(name string) bool {
	return c.IncludeUnexported || token.IsExported(name)
}

func Scan(c *Config, pkgs []*packages.Package) (*Graph, error) {
	if c.Fset == nil {
		c.Fset = token.NewFileSet()
	}

	g := graph.New(func(s *Subject) string { return s.ID })
	pkgMap := make(map[string]*packages.Package, len(pkgs))
	for _, pkg := range pkgs {
		pkgMap[pkg.ID] = pkg
	}
	scanner := &Scanner{
		g:      g,
		pkgMap: pkgMap,
		Config: c,
	}

	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			return nil, pkg.Errors[0] // TODO: multierror
		}
		for _, f := range pkg.Syntax {
			if err := scanner.Scan(pkg, f); err != nil {
				log.Printf("! %+v", err)
			}
		}
	}
	return g, nil
}

func DumpAll(c *Config, g *Graph) error {
	return Dump(c, g, g.Nodes)
}

func Dump(c *Config, g *Graph, nodes []*Node) error {
	pkgpath := c.PkgPath

	fmt.Printf("package %s\n", pkgpath)
	g.WalkPath(func(path []*Node) {
		parts := strings.Split(pkgpath, "/")
		prefix := strings.Join(parts[:len(parts)-1], "/") + "/"
		if len(path) == 1 {
			node := path[0]
			if len(node.From) == 0 && len(node.To) > 0 && c.NeedName(node.Name) && (node.Value.Recv == "" || c.NeedName(node.Value.Recv)) {
				name := strings.ReplaceAll(path[len(path)-1].Value.Object.String(), prefix, "")
				fmt.Printf("\n%s%s\n", strings.Repeat("  ", len(path)), name)
			}
		} else {
			node := path[len(path)-1]
			if c.NeedName(node.Name) && (node.Value.Recv == "" || c.NeedName(node.Value.Recv)) {
				name := strings.ReplaceAll(node.Value.Object.String(), prefix, "")
				fmt.Printf("%s%s\n", strings.Repeat("  ", len(path)), name)
			}
		}
	}, nodes)
	return nil
}
