package gobuild

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"

	"github.com/cofunclabs/cofunc/pkg/textline"
)

type mod struct {
	dir      string
	name     string
	mainpkgs []string
}

func findMods(rootDir string) (map[string]mod, error) {
	mods := make(map[string]mod)
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("%w: access path '%s'", err, path)
		}
		if info.IsDir() {
			if info.Name() == "go.mod" {
				module, err := textline.FindFileLine(path, splitSpace, getPrefix)
				if err != nil {
					return err
				}
				dir := filepath.Dir(path)
				mods[dir] = mod{
					dir:  filepath.Dir(path),
					name: module,
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: find go.mod", err)
	}
	return mods, nil
}

// findMainPkg finds the main package in the given directory automatically.
func findMainPkg(module string, dir string) ([]string, error) {
	var mains []string
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("%w: access path '%s'", err, path)
		}
		if info.IsDir() {
			fset := token.NewFileSet()
			pkgs, err := parser.ParseDir(fset, path, nil, parser.PackageClauseOnly)
			if err != nil {
				return fmt.Errorf("%w: find main package", err)
			}
			if len(pkgs) > 0 {
				for name := range pkgs {
					if name == "main" {
						mains = append(mains, filepath.Join(module, path))
						break
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%w: find main package", err)
	}
	return mains, nil
}
