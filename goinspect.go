package goinspect

import (
	"fmt"
	"go/token"
	"io"
	"log"
	"strings"

	"github.com/podhmo/goinspect/graph"
	"golang.org/x/tools/go/packages"
)

type Config struct {
	Fset *token.FileSet

	PkgPath string
	Padding string

	ExpandAll         bool
	IncludeUnexported bool
	OtherPackages     []string

	skipHeader bool
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

func DumpAll(w io.Writer, c *Config, g *Graph) error {
	return Dump(w, c, g, g.Nodes)
}

func Dump(w io.Writer, c *Config, g *Graph, nodes []*Node) error {
	pkgpath := c.PkgPath
	expand := c.ExpandAll

	rows := make([]*row, 0, len(nodes))
	sameIDRows := map[int][]*row{}

	g.WalkPath(func(path []*Node) {
		parts := strings.Split(pkgpath, "/")
		prefix := strings.Join(parts[:len(parts)-1], "/") + "/"

		node := path[len(path)-1]
		if len(path) == 1 {
			if len(node.From) == 0 && len(node.To) > 0 && c.NeedName(node.Name) && (node.Value.Recv == "" || c.NeedName(node.Value.Recv)) {
				name := strings.ReplaceAll(path[len(path)-1].Value.Object.String(), prefix, "")
				row := &row{indent: len(path), text: name, id: node.ID, hasChildren: len(node.To) > 0, isToplevel: true}
				rows = append(rows, row)
				sameIDRows[node.ID] = append(sameIDRows[node.ID], row)
			}
		} else {
			if c.NeedName(node.Name) && (node.Value.Recv == "" || c.NeedName(node.Value.Recv)) {
				name := strings.ReplaceAll(node.Value.Object.String(), prefix, "")
				row := &row{indent: len(path), text: name, id: node.ID, hasChildren: len(node.To) > 0}
				rows = append(rows, row)
				sameIDRows[node.ID] = append(sameIDRows[node.ID], row)
			}
		}
	}, nodes)

	if !c.skipHeader {
		fmt.Fprintf(w, "package %s\n", pkgpath)
	}

	seen := make(map[int][]int, len(sameIDRows))
	if expand {
		for i, row := range rows {
			if row.isToplevel {
				fmt.Fprintln(w, "")
			}

			if showID := len(sameIDRows[row.id]) > 1 && row.hasChildren; showID {
				if len(seen[row.id]) == 0 {
					fmt.Fprintf(w, "%s%s\n", strings.Repeat(c.Padding, row.indent), row.text)
				} else {
					idx := seen[row.id][0]
					st := rows[idx]
					fmt.Fprintf(w, "%s%s\n", strings.Repeat(c.Padding, row.indent), st.text)
					for _, x := range rows[idx+1:] {
						if x.indent <= st.indent {
							break
						}
						fmt.Fprintf(w, "%s%s\n", strings.Repeat(c.Padding, row.indent+(x.indent-st.indent)), x.text)
					}
				}
				seen[row.id] = append(seen[row.id], i)
			} else {
				fmt.Fprintf(w, "%s%s\n", strings.Repeat(c.Padding, row.indent), row.text)
			}

		}
	} else {
		for i, row := range rows {
			if row.isToplevel {
				fmt.Fprintln(w, "")
			}

			if showID := len(sameIDRows[row.id]) > 1 && row.hasChildren; showID {
				if len(seen[row.id]) == 0 {
					fmt.Fprintf(w, "%s%s // &%d\n", strings.Repeat(c.Padding, row.indent), row.text, row.id)
				} else {
					fmt.Fprintf(w, "%s%s // *%d\n", strings.Repeat(c.Padding, row.indent), row.text, row.id)
				}
				seen[row.id] = append(seen[row.id], i)
			} else {
				fmt.Fprintf(w, "%s%s\n", strings.Repeat(c.Padding, row.indent), row.text)
			}

		}
	}
	return nil
}

type row struct {
	indent      int
	text        string
	id          int
	hasChildren bool
	isToplevel  bool
}
