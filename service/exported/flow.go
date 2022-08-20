package exported

import (
	"encoding/json"
	"io"
	"time"
)

type NodeInsight struct {
	Seq       int    `json:"seq"`
	Step      int    `json:"step"`
	Name      string `json:"name"`
	LastError error  `json:"last_error"`
	Status    string `json:"status"`
}

type FlowInsight struct {
	Name      string        `json:"name"`
	ID        string        `json:"id"`
	Status    string        `json:"status"`
	LastError error         `json:"last_error"`
	Begin     time.Time     `json:"begin_time"`
	End       time.Time     `json:"end_time"`
	Total     int           `json:"total"`
	Running   int           `json:"running"`
	Done      int           `json:"done"`
	Nodes     []NodeInsight `json:"nodes"`
}

func (f FlowInsight) JsonWrite(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(f)
}
