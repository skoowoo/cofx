package feedbackid

import (
	"crypto/md5"
	"fmt"
)

type ID interface {
	Value() string
	Short() string
}

type DefaultID struct {
	id string
}

func NewID(s string) *DefaultID {
	encoded := md5.Sum([]byte(s))
	return &DefaultID{
		id: fmt.Sprintf("%x", encoded),
	}
}

func WrapID(id string) *DefaultID {
	return &DefaultID{
		id: id,
	}
}

func (d *DefaultID) Value() string {
	return d.id
}

func (d *DefaultID) Short() string {
	return d.id[:8]
}
