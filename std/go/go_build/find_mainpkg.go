package gobuild

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
)

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
