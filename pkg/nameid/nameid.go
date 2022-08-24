package nameid

import (
	"crypto/md5"
	"fmt"

	co "github.com/cofunclabs/cofunc"
)

type ID interface {
	Name() string
	Value() string
	ShortID() string
}

type NameID struct {
	name string
	id   string
}

func New(s string) *NameID {
	encoded := md5.Sum([]byte(s))
	return &NameID{
		name: co.TruncFlowl(s),
		id:   fmt.Sprintf("%x", encoded),
	}
}

func WrapID(id string) *NameID {
	return &NameID{
		id: id,
	}
}

func (d *NameID) Name() string {
	return d.name
}

func (d *NameID) Value() string {
	return d.id
}

func (d *NameID) ShortID() string {
	return d.id[:8]
}
