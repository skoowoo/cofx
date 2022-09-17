package exported

import (
	"encoding/json"
	"io"

	"github.com/cofxlabs/cofx/manifest"
)

type ListStdFunctions struct {
	Category string `json:"category"`
	Name     string `json:"name"`
	Desc     string `json:"desc"`
}

func (l ListStdFunctions) JsonWrite(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(l)
}

type InspectStdFunction manifest.Manifest

func (i InspectStdFunction) JsonWrite(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(i)
}
