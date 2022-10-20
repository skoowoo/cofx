package outcome

import (
	"context"

	"github.com/cofxlabs/cofx/functiondriver/go/spec"
	"github.com/cofxlabs/cofx/manifest"
	"github.com/cofxlabs/cofx/pkg/textparse"
)

var _manifest = manifest.Manifest{
	Name:        "outcome",
	Driver:      "go",
	Description: "Save outcome of the functions to a database",
}

func New() (*manifest.Manifest, spec.EntrypointFunc, spec.CreateCustomFunc) {
	return &_manifest, Entrypoint, nil
}

func Entrypoint(ctx context.Context, bundle spec.EntrypointBundle, args spec.EntrypointArgs) (map[string]string, error) {
	fid := bundle.Resources.Labels.GetFlowID()
	seq := bundle.Resources.Labels.GetNodeSeq()
	name := bundle.Resources.Labels.GetNodeName()

	columns := []string{"flow_id", "node_seq", "node_name", "key", "value"}
	values := []string{fid, seq, name}

	for k, v := range args {
		vs := textparse.String2Slice(v)
		for _, s := range vs {
			rv := append(values, k, s)
			if err := bundle.Resources.Outcome.Insert(ctx, columns, rv); err != nil {
				return nil, err
			}
		}
	}
	return nil, nil
}
