package goinspect

import (
	"fmt"
	"go/token"
	"io"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/podhmo/goinspect/graph"
	"golang.org/x/tools/go/packages"
)

type Config struct {
	Fset *token.FileSet

	PkgPath    string
	Padding    string
	TrimPrefix string

	ExpandAll         bool
	IncludeUnexported bool
	IncludeStruct     bool
	OtherPackages     []string

	Debug           bool
	skipHeader      bool
	forceIncludeMap map[string]bool
}

func (c *Config) NeedName(name string) bool {
	return c.IncludeUnexported || token.IsExported(name) || c.forceIncludeMap[name]
}

func Scan(c *Config, pkgs []*packages.Package) (*Graph, error) {
	if c.Fset == nil {
		c.Fset = token.NewFileSet()
	}
	if c.forceIncludeMap == nil {
		c.forceIncludeMap = map[string]bool{}
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

	matched := false
	for _, pkg := range pkgs {
		if pkg.PkgPath != c.PkgPath {
			continue
		}

		matched = true
		if len(pkg.Errors) > 0 {
			return nil, pkg.Errors[0] // TODO: multierror
		}

		// when main package, include main() forcely.
		if pkg.Name == "main" {
			c.forceIncludeMap["main"] = true
			c.forceIncludeMap["run"] = true
		}

		for _, f := range pkg.Syntax {
			if err := scanner.Scan(pkg, f); err != nil {
				return nil, err
			}
		}
	}
	if !matched {
		return nil, fmt.Errorf("pkg is not found, %q", c.PkgPath)
	}
	return g, nil
}

func DumpAll(w io.Writer, c *Config, g *Graph) error {
	return dump(w, c, g, g.Nodes, nil)
}

func Dump(w io.Writer, c *Config, g *Graph, nodes []*Node) error {
	selected := make([]*Node, 0, len(nodes))
	seen := make(map[int]struct{}, len(g.Nodes))

	{
		q := nodes[:]
		var n *Node
		for len(q) > 0 {
			n, q = q[0], q[1:]
			if _, ok := seen[n.ID]; ok {
				continue
			}
			seen[n.ID] = struct{}{}
			if len(n.From) == 0 {
				selected = append(selected, n)
			} else {
				copied := make([]*Node, len(n.From))
				copy(copied, n.From)
				q = append(q, copied...)
			}
		}
	}

	{
		q := make([]*Node, 0, len(nodes))
		for _, n := range nodes {
			if len(n.To) > 0 {
				q = append(q, n.To...)
			}
		}
		var n *Node
		for len(q) > 0 {
			n, q = q[0], q[1:]
			if _, ok := seen[n.ID]; ok {
				continue
			}
			seen[n.ID] = struct{}{}
			if len(n.To) > 0 {
				copied := make([]*Node, len(n.To))
				copy(copied, n.To)
				q = append(q, copied...)
			}
		}
	}

	return dump(w, c, g, selected, seen)
}

func dump(w io.Writer, c *Config, g *Graph, nodes []*Node, filter map[int]struct{}) error {
	pkgpath := c.PkgPath
	expand := c.ExpandAll

	rows := make([]*row, 0, len(nodes))
	sameIDRows := map[int][]*row{}

	parts := strings.Split(pkgpath, "/")
	prefix := strings.Join(parts[:len(parts)-1], "/") + "/"

	{
		sorted := g.SortedByFrom(nodes)
		sortedMap := make(map[int]int, len(sorted))
		for i, n := range sorted {
			sortedMap[n.ID] = i
		}
		g.Walk(func(n *Node) {
			if n.Value.Kind == KindObject {
				if len(n.To) > 0 {
					sort.SliceStable(n.To, func(i, j int) bool { return sortedMap[n.To[i].ID] < sortedMap[n.To[j].ID] })
				}
			}
		})
	}

	prevIndent := 0
	g.WalkPath(func(path []*Node) {
		node := path[len(path)-1]
		if filter != nil {
			if _, ok := filter[node.ID]; !ok {
				return
			}
			for _, x := range path[:len(path)-1] {
				if _, ok := filter[x.ID]; !ok {
					return
				}
			}
		}

		indent := len(path)
		if indent == 1 {
			if len(node.From) == 0 && c.NeedName(node.Name) && (node.Value.Recv == "" || c.NeedName(node.Value.Recv)) {
				if node.Value.Kind == KindObject && len(node.To) == 0 {
					return
				}

				text := strings.ReplaceAll(path[indent-1].Value.Object.String(), prefix, "")
				if c.TrimPrefix != "" {
					text = strings.ReplaceAll(text, c.TrimPrefix, "")
				}

				row := &row{indent: indent, name: node.Name, text: text, id: node.ID, kind: node.Value.Kind, hasChildren: len(node.To) > 0, isToplevel: true}
				rows = append(rows, row)
				sameIDRows[node.ID] = append(sameIDRows[node.ID], row)
				prevIndent = row.indent
			}
		} else {
			if (filter != nil || prevIndent == 0) && prevIndent < indent && indent-prevIndent > 1 { // for --only with sub nodes
				return
			}
			if c.NeedName(node.Name) && (node.Value.Recv == "" || c.NeedName(node.Value.Recv)) {
				text := strings.ReplaceAll(node.Value.Object.String(), prefix, "")
				if c.TrimPrefix != "" {
					text = strings.ReplaceAll(text, c.TrimPrefix, "")
				}

				isRecursive := false
				for _, x := range path[:len(path)-1] {
					if x.ID == node.ID {
						isRecursive = true
					}
				}
				row := &row{indent: indent, name: node.Name, text: text, id: node.ID, kind: node.Value.Kind, hasChildren: len(node.To) > 0, isRecursive: isRecursive}
				rows = append(rows, row)
				sameIDRows[node.ID] = append(sameIDRows[node.ID], row)
				prevIndent = row.indent
			}
		}
	}, nodes)

	if !c.skipHeader {
		fmt.Fprintf(w, "package %s\n", pkgpath)
	}

	seen := make(map[int][]int, len(sameIDRows))
	if expand {
		var dumpCache func(*row, int, int) int
		dumpCache = func(row *row, indent int, i int) int {
			idx := seen[row.id][0]
			st := rows[idx]
			seen[row.id] = append(seen[row.id], i)
			emit(w, c, indent, st)
			if c.Debug {
				fmt.Fprintf(w, "  // c *%d\n", st.id)
			} else {
				fmt.Fprintln(w, "")
			}
			idx++
			for {
				x := rows[idx]
				if x.indent <= st.indent {
					return idx
				}

				idt := indent + (x.indent - st.indent)
				if showID := len(sameIDRows[x.id]) > 1; showID {
					if x.isRecursive {
						seen[x.id] = append(seen[x.id], i)
						emit(w, c, idt, x)
						if c.Debug {
							fmt.Fprintf(w, "  // c *%d  recursion\n", x.id)
						} else {
							fmt.Fprintln(w, " // recursion")
						}
					} else {
						for _, j := range seen[x.id] {
							if i == j {
								return idx
							}
						}
						dumpCache(x, idt, i)
					}
				} else {
					seen[x.id] = append(seen[x.id], i)
					emit(w, c, idt, x)
					if c.Debug {
						fmt.Fprintf(w, "  // c *%d\n", x.id)
					} else {
						fmt.Fprintln(w, "")
					}
				}
				idx++
			}
		}

		scopeID := 0
		var scopeKind Kind
		for i, row := range rows {
			if row.isToplevel {
				fmt.Fprintln(w, "")
				scopeID = i
				scopeKind = row.kind
			}

			if row.indent == 2 && scopeKind == KindObject { // skip used method declaration
				hist, ok := seen[row.id]
				if ok && scopeID < hist[len(hist)-1] {
					continue
				}
			}

			if showID := len(sameIDRows[row.id]) > 1; showID {
				if len(seen[row.id]) == 0 {
					seen[row.id] = append(seen[row.id], i)
					emit(w, c, row.indent, row)
					if c.Debug {
						fmt.Fprintf(w, "  // &%d\n", row.id) // define reference
					} else {
						fmt.Fprintln(w, "")
					}
				} else if row.isRecursive {
					seen[row.id] = append(seen[row.id], i)
					emit(w, c, row.indent, row)
					if c.Debug {
						fmt.Fprintf(w, "  // *%d recursion\n", row.id) // define reference
					} else {
						fmt.Fprintln(w, "  // recursion")
					}
				} else {
					dumpCache(row, row.indent, i)
				}
			} else {
				seen[row.id] = append(seen[row.id], i)
				emit(w, c, row.indent, row)
				fmt.Fprintln(w, "")
			}
		}
	} else {
		for i, row := range rows {
			if row.isToplevel {
				fmt.Fprintln(w, "")
			}
			if showID := len(sameIDRows[row.id]) > 1; showID {
				if len(seen[row.id]) == 0 {
					emit(w, c, row.indent, row)
					fmt.Fprintf(w, "  // &%d\n", row.id) // define reference
				} else if row.isRecursive {
					emit(w, c, row.indent, row)
					fmt.Fprintf(w, "  // *%d recursion\n", row.id)
				} else {
					emit(w, c, row.indent, row)
					fmt.Fprintf(w, "  // *%d\n", row.id) // use reference
				}
				seen[row.id] = append(seen[row.id], i)
			} else {
				emit(w, c, row.indent, row)
				fmt.Fprintln(w, "")
				seen[row.id] = append(seen[row.id], i)
			}

		}
	}

	if c.Debug {
		fmt.Fprintln(os.Stderr)
		log.Printf("** rows of %s **", c.PkgPath)
		for i, row := range rows {
			log.Printf("%3d: %s%s", i, strings.Repeat("@", row.indent), row.text)
		}
	}
	return nil
}

func emit(w io.Writer, c *Config, indent int, row *row) {
	if c.Debug {
		fmt.Fprintf(w, "%3d: %s%s", indent, strings.Repeat(c.Padding, indent), row.text)
	} else {
		fmt.Fprintf(w, "%s%s", strings.Repeat(c.Padding, indent), row.text)
	}
}

type row struct {
	indent int
	name   string
	text   string
	id     int

	kind        Kind
	hasChildren bool
	isToplevel  bool
	isRecursive bool
}
