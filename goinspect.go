package goinspect

import (
	"fmt"
	"go/token"
	"io"
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

	parts := strings.Split(pkgpath, "/")
	prefix := strings.Join(parts[:len(parts)-1], "/") + "/"

	sections := make(map[int]*section, len(nodes))
	order := make([]*section, 0, len(nodes))
	counter := make(map[int]int, len(nodes))

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

				if _, ok := sections[node.ID]; !ok {
					text := strings.ReplaceAll(node.Value.Object.String(), prefix, "")
					if c.TrimPrefix != "" {
						text = strings.ReplaceAll(text, c.TrimPrefix, "")
					}
					s := &section{
						node: node,
						text: text,
					}
					sections[node.ID] = s
					order = append(order, s)
				}
				counter[node.ID]++
			}
		} else {
			if c.NeedName(node.Name) && (node.Value.Recv == "" || c.NeedName(node.Value.Recv)) {
				parent := path[len(path)-2]
				p, ok := sections[parent.ID]
				if !ok {
					text := strings.ReplaceAll(parent.Value.Object.String(), prefix, "")
					if c.TrimPrefix != "" {
						text = strings.ReplaceAll(text, c.TrimPrefix, "")
					}
					p = &section{
						node: parent,
						text: text,
					}
					sections[parent.ID] = p
					order = append(order, p)
				}

				s, ok := sections[node.ID]
				if !ok {
					text := strings.ReplaceAll(node.Value.Object.String(), prefix, "")
					if c.TrimPrefix != "" {
						text = strings.ReplaceAll(text, c.TrimPrefix, "")
					}
					s = &section{
						node: node,
						text: text,
					}
					sections[node.ID] = s
					order = append(order, s)
				}
				p.body = append(p.body, s)
				counter[node.ID]++
			}
		}
	}, g.SortedByFrom(nodes))

	seen := make(map[int]bool, len(nodes))
	var stack []*section
	var walk func(*section, int)
	walk = func(s *section, indent int) {
		recursive := false
		for _, x := range stack {
			if x.node.ID == s.node.ID {
				recursive = true
			}
		}

		suffix := ""
		if recursive {
			suffix = " recursion"
		}
		if counter[s.node.ID] == 1 {
			fmt.Fprintf(w, "%3d: %s %s\n", indent, strings.Repeat(c.Padding, indent), s.text)
		} else if _, visited := seen[s.node.ID]; !visited {
			fmt.Fprintf(w, "%3d: %s %s  // &%d%s\n", indent, strings.Repeat(c.Padding, indent), s.text, s.node.ID, suffix)
		} else {
			fmt.Fprintf(w, "%3d: %s %s  // *%d%s\n", indent, strings.Repeat(c.Padding, indent), s.text, s.node.ID, suffix)
		}
		if recursive {
			return
		}
		seen[s.node.ID] = true
		stack = append(stack, s) // push
		for _, child := range s.body {
			walk(child, indent+1)
		}
		stack = stack[:len(stack)-1] // pop
	}
	for _, s := range order {
		if _, visited := seen[s.node.ID]; visited {
			continue
		}
		fmt.Fprintln(w, "")
		walk(s, 0)
	}
	return nil
}

type section struct {
	node        *Node
	text        string
	body        []*section
	isRecursive bool
}
