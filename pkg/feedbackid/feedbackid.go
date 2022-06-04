package feedbackid

import "encoding/base64"

type ID interface {
	Value() string
	String() string
	Feedback(m string)
}

type DefaultID struct {
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
	return d.id
}

func (d *DefaultID) String() string {
	return d.id + ":" + d.message
}

func (d *DefaultID) Feedback(m string) {
	d.message = m
}
