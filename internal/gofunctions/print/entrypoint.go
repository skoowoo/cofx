package print

import (
	"fmt"
	"os"
	"sort"

	"github.com/cofunclabs/cofunc/pkg/manifest"
)

func New() manifest.Manifester {
	return &printer{}
}

// Be used to test the go driver
type printer struct {
}

func (p *printer) Manifest() manifest.Manifest {
	return manifest.Manifest{
		Driver:         "go",
		EntryPointFunc: p.EntryPoint,
	}
}

func (p *printer) Name() string {
	return "print"
}

func (p *printer) EntryPoint(args map[string]string) (map[string]string, error) {
	var slice []string
	for k, v := range args {
		slice = append(slice, k+" = "+v)
	}
	sort.Strings(slice)
	for _, s := range slice {
		fmt.Fprintln(os.Stdout, s)
	}
	return map[string]string{
		"status": "ok",
	}, nil
}
