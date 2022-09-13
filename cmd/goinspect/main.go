package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/token"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/podhmo/flagstruct"
	"github.com/podhmo/goinspect"
	"golang.org/x/tools/go/packages"
)

type Options struct {
	IncludeUnexported bool `flag:"include-unexported"`
	ExpandAll         bool `flag:"expand-all"`
	Short             bool `flag:"short"`

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

	if options.Short {
		if mods, err := modInfo(); err == nil && len(mods) > 0 {
			c.TrimPrefix = mods[0].Path
		}
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

// from: golang.org/x/tools/cmd/godoc/main.go

// goMod returns the go env GOMOD value in the current directory
// by invoking the go command.
//
// GOMOD is documented at https://golang.org/cmd/go/#hdr-Environment_variables:
//
//	The absolute path to the go.mod of the main module,
//	or the empty string if not using modules.
func goMod() (string, error) {
	out, err := exec.Command("go", "env", "-json", "GOMOD").Output()
	if ee := (*exec.ExitError)(nil); errors.As(err, &ee) {
		return "", fmt.Errorf("go command exited unsuccessfully: %v\n%s", ee.ProcessState.String(), ee.Stderr)
	} else if err != nil {
		return "", err
	}
	var env struct {
		GoMod string
	}
	err = json.Unmarshal(out, &env)
	if err != nil {
		return "", err
	}
	return env.GoMod, nil
}

type mod struct {
	Path string // Module path.
	Dir  string // Directory holding files for this module, if any.
}

// buildList determines the build list in the current directory
// by invoking the go command. It should only be used in module mode,
// when vendor mode isn't on.
//
// See https://golang.org/cmd/go/#hdr-The_main_module_and_the_build_list.
func modInfo() ([]mod, error) {
	_, err := goMod()
	if err != nil {
		return nil, nil
	}

	out, err := exec.Command("go", "list", "-m", "-json").Output()
	if ee := (*exec.ExitError)(nil); errors.As(err, &ee) {
		return nil, fmt.Errorf("go command exited unsuccessfully: %v\n%s", ee.ProcessState.String(), ee.Stderr)
	} else if err != nil {
		return nil, err
	}
	var mods []mod
	for dec := json.NewDecoder(bytes.NewReader(out)); ; {
		var m mod
		err := dec.Decode(&m)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		mods = append(mods, m)
	}
	return mods, nil
}
