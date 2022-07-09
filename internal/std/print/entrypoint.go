package print

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/cofunclabs/cofunc/pkg/manifest"
	"github.com/sirupsen/logrus"
)

func New() manifest.Manifester {
	return &printer{}
}

// Be used to test the go driver
type printer struct{}

func (p *printer) Manifest() manifest.Manifest {
	return manifest.Manifest{
		Driver:         "go",
		EntrypointFunc: p.Entrypoint,
	}
}

func (p *printer) Name() string {
	return "print"
}

func (p *printer) Entrypoint(ctx context.Context, args map[string]string) (map[string]string, error) {
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
