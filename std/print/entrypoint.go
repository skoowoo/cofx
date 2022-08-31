package print

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/cofunclabs/cofunc/functiondriver/go/spec"
	"github.com/cofunclabs/cofunc/manifest"
)

var _manifest = manifest.Manifest{
	Name:   "print",
	Driver: "go",
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	var slice []string
	for k, v := range args {
		if strings.HasPrefix(k, "_") {
			slice = append(slice, v)
		} else {
			slice = append(slice, k+": "+v)
		}
	}
	sort.Strings(slice)
	for _, s := range slice {
		fmt.Fprintln(bundle.Logger, s)
	}
	return map[string]string{
		"status": "ok",
	}, nil
}
