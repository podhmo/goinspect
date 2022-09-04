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
	ID     string
	Object types.Object
}

type Graph = graph.Graph[string, *Subject]
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
		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch spec := spec.(type) {
				case *ast.TypeSpec:
					if err := s.scanTypeSpec(pkg, f, spec); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (s *Scanner) Need(name string) bool {
	return s.Config.IncludeUnexported || token.IsExported(name)
}

func (s *Scanner) scanFuncDecl(pkg *packages.Package, f *file, decl *ast.FuncDecl) error {
	// func <name>(...) ... { ... }

	ob := pkg.TypesInfo.Defs[decl.Name]
	subject := &Subject{ID: pkg.ID + "." + decl.Name.Name, Object: ob}
	node := s.g.Madd(subject)
	node.Name = decl.Name.Name

	if decl.Recv != nil {
		if sig, ok := ob.Type().(*types.Signature); ok {
			recv := sig.Recv()
			recvType := recv.Type()
			if t, ok := recvType.(*types.Pointer); ok {
				recvType = t.Elem()
			}
			if named, ok := recvType.(*types.Named); ok {
				typob := named.Obj()
				parent := s.g.Madd(&Subject{ID: pkg.ID + "." + typob.Name(), Object: typob})
				parent.Name = typob.Name()
				s.g.LinkTo(parent, node)
			}
		}
	}

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
							subject := &Subject{Object: ob, ID: impkg.ID + "." + sym.Sel.Name}
							child := s.g.Madd(subject)
							child.Name = sym.Sel.Name
							s.g.LinkTo(node, child)
						}
					}
				}
			case *ast.Ident:
				if ob, ok := pkg.TypesInfo.Uses[sym]; ok {
					subject := &Subject{ID: pkg.ID + "." + sym.Name, Object: ob}
					child := s.g.Madd(subject)
					child.Name = sym.Name
					s.g.LinkTo(node, child)
				}
			}
		}
		return true
	})
	return nil
}

func (s *Scanner) scanTypeSpec(pkg *packages.Package, f *file, spec *ast.TypeSpec) error {
	// type <name> = <type>
	// type <name> <type>
	// type <name> struct { ... }
	// type <name> interface { ... }

	ob := pkg.TypesInfo.Defs[spec.Name]
	subject := &Subject{ID: pkg.ID + "." + spec.Name.Name, Object: ob}
	node := s.g.Madd(subject)
	node.Name = spec.Name.Name

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
