package exported

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type FlowInsight struct {
	Status    string    `json:"status"`
	LastError error     `json:"last_error"`
	Begin     time.Time `json:"begin_time"`
	End       time.Time `json:"end_time"`
	Total     int       `json:"total"`
	Running   int       `json:"running"`
	Done      int       `json:"done"`
	Nodes     []struct {
		Seq       int    `json:"seq"`
		Step      int    `json:"step"`
		Name      string `json:"name"`
		LastError error  `json:"last_error"`
		Status    string `json:"status"`
	} `json:"nodes"`
}

func (f FlowInsight) PrettyPrint(w io.Writer) error {
	fmt.Fprintf(w, "%s %d %d %d\n", f.Status, f.Total, f.Running, f.Done)
	return nil
}

func (f FlowInsight) JsonPrint(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(f)
}
