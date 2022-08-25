package exported

import (
	"encoding/json"
	"io"
	"time"
)

type NodeRunningInsight struct {
	Seq       int    `json:"seq"`
	Step      int    `json:"step"`
	Name      string `json:"name"`
	Function  string `json:"function"`
	Driver    string `json:"driver"`
	LastError error  `json:"last_error"`
	Status    string `json:"status"`
	Runs      int    `json:"runs"`
	Duration  int64  `json:"duration"`
}

type FlowRunningInsight struct {
	Name      string               `json:"name"`
	ID        string               `json:"id"`
	Status    string               `json:"status"`
	LastError error                `json:"last_error"`
	Begin     time.Time            `json:"begin_time"`
	Duration  int64                `json:"duration"`
	Total     int                  `json:"total"`
	Running   int                  `json:"running"`
	Done      int                  `json:"done"`
	Nodes     []NodeRunningInsight `json:"nodes"`
}

func (f FlowRunningInsight) JsonWrite(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(f)
}

type FlowMetaInsight struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Total  int    `json:"total"`
	Source string `json:"source"`
	Desc   string `json:"desc"`
}

func (f FlowMetaInsight) JsonWrite(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(f)
}
