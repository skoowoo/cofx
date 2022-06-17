package feedbackid

import (
	"encoding/base64"
	"sync"
)

type ID interface {
	Value() string
	String() string
	Feedback(m string)
}

type DefaultID struct {
	sync.RWMutex
	id      string
	message string
}

func NewDefaultID(s string) *DefaultID {
	encoded := base64.StdEncoding.EncodeToString([]byte(s))
	return &DefaultID{
		id: encoded,
	}
}

func (d *DefaultID) Value() string {
	d.RLock()
	defer d.RUnlock()
	return d.id
}

func (d *DefaultID) String() string {
	d.RLock()
	defer d.RUnlock()
	return d.id + ":" + d.message
}

func (d *DefaultID) Feedback(m string) {
	d.Lock()
	defer d.Unlock()
	d.message = m
}
