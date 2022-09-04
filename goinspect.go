package goinspect

import (
	"go/ast"
	"go/token"
	"go/types"
	"strconv"
	"strings"
	"sync"

	"github.com/podhmo/goinspect/graph"
	"golang.org/x/tools/go/packages"
)

type Subject struct {
	ID     *ast.Ident
	Object types.Object
}

type Graph = graph.Graph[*ast.Ident, *Subject]
type Node = graph.Node[*Subject]

type Scanner struct {
	g      *Graph
	pkgMap map[string]*packages.Package

	Config struct {
		IncludeUnexported bool
		// IncludeOtherPackage bool
	}
}

func (s *Scanner) Scan(pkg *packages.Package, t *ast.File) error {
	f := &file{t: t}
	for _, decl := range t.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if err := s.scanFuncDecl(pkg, f, decl); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Scanner) Need(name string) bool {
	return s.Config.IncludeUnexported || token.IsExported(name)
}

func (s *Scanner) scanFuncDecl(pkg *packages.Package, f *file, decl *ast.FuncDecl) error {
	ob := pkg.TypesInfo.Defs[decl.Name]
	subject := &Subject{ID: decl.Name, Object: ob}
	node := s.g.Madd(subject)
	node.Name = subject.ID.Name

	ast.Inspect(decl.Body, func(t ast.Node) bool {
		switch t := t.(type) {
		case *ast.CallExpr:
			switch sym := t.Fun.(type) {
			case *ast.SelectorExpr:
				// <x>.<sel>
				// fmt.Println(sym, sym.X, sym.Sel)

				switch x := sym.X.(type) {
				case *ast.Ident:
					if impath, ok := f.ImportPath(x.Name); ok {
						if impkg, ok := s.pkgMap[impath]; ok {
							ob := impkg.Types.Scope().Lookup(sym.Sel.Name)
							var id *ast.Ident
							for k := range impkg.TypesInfo.Defs {
								if k.Name == sym.Sel.Name {
									id = k
									break
								}
							}
							subject := &Subject{Object: ob, ID: id}
							child := s.g.Madd(subject)
							child.Name = subject.ID.Name
							s.g.LinkTo(node, child)
						}
					}
				}
			case *ast.Ident:
				if sym.Obj != nil {
					id := sym.Obj.Decl.(*ast.FuncDecl).Name
					if ob, ok := pkg.TypesInfo.Defs[id]; ok {
						subject := &Subject{ID: id, Object: ob}
						child := s.g.Madd(subject)
						child.Name = subject.ID.Name
						s.g.LinkTo(node, child)
					}
				}
			}
		}
		return true
	})
	return nil
}

type file struct {
	t       *ast.File
	imports map[string]string // name -> path
	sync.Once
}

func (f *file) ImportPath(name string) (string, bool) {
	f.Once.Do(func() {
		imports := make(map[string]string, len(f.t.Imports))
		for _, im := range f.t.Imports {
			path, _ := strconv.Unquote(im.Path.Value)
			name := ""
			if im.Name != nil {
				name = im.Name.Name
			} else {
				parts := strings.Split(path, "/") // TODO: this has bug (e.g. go-sqlite)
				name = parts[len(parts)-1]
			}
			imports[name] = path
		}
		f.imports = imports
	})
	path, ok := f.imports[name]
	return path, ok
}
