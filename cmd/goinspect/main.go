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
	"path/filepath"
	"strings"

	"github.com/podhmo/flagstruct"
	"github.com/podhmo/goinspect"
	"golang.org/x/tools/go/packages"
)

type Options struct {
	IncludeUnexported bool `flag:"include-unexported"`
	ExpandAll         bool `flag:"expand-all"`
	Short             bool `flag:"short"`
	Debug             bool `flag:"debug"`

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

	if strings.HasSuffix(c.PkgPath, "/...") {
		c.OtherPackages = append(c.OtherPackages, c.PkgPath)
		c.PkgPath = strings.TrimSuffix(c.PkgPath, "/...")
	}

	cfg := &packages.Config{
		Fset: fset,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedImports | packages.NeedTypes | packages.NeedTypesSizes | packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedDeps,
	}
	pkgs, err := packages.Load(cfg, append([]string{c.PkgPath}, c.OtherPackages...)...)
	if err != nil {
		return fmt.Errorf("load packages: %w", err)
	}

	{
		goMod, err := goMod()
		if err != nil {
			log.Println("go mod failed, %w", err)
		}

		mods, err := modInfo(goMod)
		if err == nil && len(mods) > 0 {
			modInfo := mods[0]
			// handling --short
			if options.Short {
				c.TrimPrefix = modInfo.Path
			}

			// detect fullpath from relative path
			if strings.HasPrefix(c.PkgPath, ".") {
				if err := func() error {
					cwd, err := os.Getwd()
					if err != nil {
						return fmt.Errorf("getwd: %w", err)
					}
					pkgpath, err := fullPkgPath(c.PkgPath, cwd, modInfo, pkgs, options.Debug)
					if err != nil {
						return err
					}
					log.Printf("detect package path, %q -> %q", c.PkgPath, pkgpath)
					c.PkgPath = pkgpath
					return nil
				}(); err != nil {
					return fmt.Errorf("detect package path is failed: %w", err)
				}
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

// modInfo determines the go mod info in the current directory
// by invoking the go command. It should only be used in module mode,
// when vendor mode isn't on.
//
// See https://golang.org/cmd/go/#hdr-The_main_module_and_the_build_list.
func modInfo(goMod string) ([]mod, error) {
	if goMod == os.DevNull {
		// Empty build list.
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

func fullPkgPath(pkgpath string, cwd string, modInfo mod, pkgs []*packages.Package, debug bool) (string, error) {
	pkgpath = strings.TrimSuffix(pkgpath, "/")

	var prefix, suffix string
	parts := strings.Split(pkgpath, "/")
	for i, x := range parts {
		switch x {
		case ".":
			continue
		case "..":
			prefix += "../"
		default:
			suffix = strings.Join(parts[i-1:], "/")
		}
	}

	// e.g.
	// modInfo.Dir :: ~/go/github.com/<user>/<name>
	// modInfo.Path :: /github.com/<user>/<name>
	// pkgpath :: ../xxx/yyy
	// cwd :: ~/go/github.com/<user>/<name>/foo/bar
	// prefix :: ..
	// suffix :: xxx/yyy
	if debug {
		log.Printf("* modInfo: %+v", modInfo)
		log.Printf("* cwd: %v", cwd)
		log.Printf("* prefix: %v", prefix)
		log.Printf("* suffix: %v", suffix)
	}

	// pkgpath is github.com/<user>/<name>/foo/xxx/yyy
	abspath, err := filepath.Abs(filepath.Join(cwd, prefix))
	if err != nil {
		return "", fmt.Errorf("abspath: %w", err)
	}
	rel, err := filepath.Rel(modInfo.Dir, abspath)
	if err != nil {
		return "", fmt.Errorf("rel: %w", err)
	}
	if debug {
		log.Printf("* rel: %v", rel)
	}

	xs := make([]string, 0, 3)
	for _, x := range []string{modInfo.Path, rel, suffix} {
		if x == "." || x == "" {
			continue
		}
		xs = append(xs, strings.TrimSuffix(strings.TrimPrefix(x, "/"), "/"))
	}
	fullpkgpath := strings.Join(xs, "/")

	if debug {
		log.Printf("* pkgpath: %v", fullpkgpath)
	}
	for _, pkg := range pkgs {
		if debug {
			log.Printf("\t* pkg check %5v: pkgpath == %v", pkg.PkgPath == fullpkgpath, pkg.PkgPath)
		}
		if pkg.PkgPath == fullpkgpath {
			return fullpkgpath, nil
		}
	}
	return "", fmt.Errorf("pkg is not found, %q", pkgpath)
}
