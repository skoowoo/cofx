package print

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/manifest"
)

var _manifest = manifest.Manifest{
	Name:        "print",
	Driver:      "go",
	Description: "Print to stdout",
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
		fmt.Fprintln(bundle.Resources.Logwriter, s)
	}
	return nil, nil
}
