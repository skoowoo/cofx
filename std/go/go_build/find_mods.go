package gobuild

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/cofunclabs/cofunc/pkg/textline"
)

type modinfo struct {
	dir      string
	name     string
	mainpkgs []string
}

// findMods finds all modules in the given directory.
func findMods(rootDir string, mainDirs []string) (map[string]*modinfo, error) {
	mods := make(map[string]*modinfo)
	err := filepath.Walk(rootDir, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("%w: access path '%s'", err, path)
		}
		if !info.IsDir() {
			if info.Name() == "go.mod" {
				module, err := textline.FindFileLine(path, splitSpace, getPrefix)
				if err != nil {
					return err
				}
				dir := filepath.Dir(path)
				mods[dir] = &modinfo{
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
	for _, dir := range mainDirs {
		if err := findMainPkg(mods, dir); err != nil {
			return nil, err
		}
	}
	return mods, nil
}

// findMainPkg finds the main package in the given directory automatically.
func findMainPkg(mods map[string]*modinfo, dir string) error {
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
			if len(pkgs) == 0 {
				return nil
			}
			for name := range pkgs {
				if name != "main" {
					continue
				}
				for p := path; ; p = filepath.Dir(p) {
					if m, ok := mods[p]; ok {
						path = strings.TrimPrefix(path, m.dir)
						m.mainpkgs = append(m.mainpkgs, filepath.Join(m.name, path))
						return nil
					}
					if p == "." {
						return nil
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("%w: find main package", err)
	}
	return nil
}
