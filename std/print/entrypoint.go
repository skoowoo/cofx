package print

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/cofunclabs/cofunc/manifest"
	"github.com/sirupsen/logrus"
)

var _manifest = manifest.Manifest{
	Name:           "print",
	Driver:         "go",
	EntrypointFunc: Entrypoint,
}

func New() *manifest.Manifest {
	return &_manifest
}

func Entrypoint(ctx context.Context, version string, args map[string]string) (map[string]string, error) {
	logrus.Debugf("function print: args=%v\n", args)
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
		fmt.Fprintln(os.Stdout, s)
	}
	return map[string]string{
		"status": "ok",
	}, nil
}
