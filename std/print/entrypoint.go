package print

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/cofunclabs/cofunc/manifest"
)

var _manifest = manifest.Manifest{
	Name:   "print",
	Driver: "go",
}

func New() (*manifest.Manifest, manifest.EntrypointFunc) {
	return &_manifest, Entrypoint
}

func Entrypoint(ctx context.Context, out io.Writer, version string, args map[string]string) (map[string]string, error) {
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
		fmt.Fprintln(out, s)
	}
	return map[string]string{
		"status": "ok",
	}, nil
}
