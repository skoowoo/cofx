package exported

import (
	"encoding/json"
	"io"
)

type SimpleError struct {
	Error string   `json:"error"`
	Desc  []string `json:"desc"`
}

func (s SimpleError) JsonWrite(w io.Writer) error {
	return json.NewEncoder(w).Encode(s)
}

type SimpleSucceed struct {
	Message string   `json:"message"`
	Desc    []string `json:"desc"`
}

func (s SimpleSucceed) JsonWrite(w io.Writer) error {
	return json.NewEncoder(w).Encode(s)
}
