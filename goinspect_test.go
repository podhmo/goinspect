package goinspect

import (
	"go/token"
	"testing"
)

func TestParse(t *testing.T) {
	pkg := "github.com/podhmo/goinspect/internal/x"
	fset := token.NewFileSet()
	cfg := &Config{
		Fset:    fset,
		PkgPath: pkg,
		OtherPackages: []string{
			"github.com/podhmo/goinspect/internal/x/sub",
		},

		IncludeUnexported: true,
	}

	result, err := Scan(cfg)
	if err != nil {
		t.Errorf("unexpected error: %+v", err)
	}

	if err := Dump(result); err != nil {
		t.Errorf("unexpected error: %+v", err)
	}
}
