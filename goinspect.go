package goinspect

import (
	"go/ast"
	"go/token"
	"go/types"

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
	g                 *Graph
	IncludeUnexported bool
}

func (s *Scanner) Scan(pkg *packages.Package, f *ast.File) error {
	for _, decl := range f.Decls {
		switch decl := decl.(type) {
		case *ast.FuncDecl:
			if err := s.scanFuncDecl(pkg, decl); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Scanner) Need(name string) bool {
	return s.IncludeUnexported || token.IsExported(name)
}

func (s *Scanner) scanFuncDecl(pkg *packages.Package, decl *ast.FuncDecl) error {
	ob := pkg.TypesInfo.Defs[decl.Name]
	subject := &Subject{ID: decl.Name, Object: ob}
	node := s.g.Madd(subject)
	node.Name = subject.ID.Name

	ast.Inspect(decl.Body, func(t ast.Node) bool {
		switch t := t.(type) {
		case *ast.CallExpr:
			switch sym := t.Fun.(type) {
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
