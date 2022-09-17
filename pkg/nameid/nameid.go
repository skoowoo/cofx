package nameid

import (
	"crypto/md5"
	"errors"
	"fmt"

	co "github.com/cofxlabs/cofx"
)

type NameOrID string

func (n NameOrID) String() string {
	return string(n)
}

type ID interface {
	Name() string
	ID() string
	String() string
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

func Wrap(name string, id string) *NameID {
	return &NameID{
		name: name,
		id:   id,
	}
}

func Guess(s NameOrID, guessID func(string) *NameID, guessName ...func(string) *NameID) (*NameID, error) {
	if id := guessID(s.String()); id != nil {
		return id, nil
	}
	if len(guessName) != 0 {
		gn := guessName[0]
		if id := gn(s.String()); id != nil {
			return id, nil
		}
	} else {
		if id := guessID(New(s.String()).ID()); id != nil {
			return id, nil
		}
	}
	return nil, errors.New("not a valid name or id: " + s.String())
}

func (d *NameID) Name() string {
	return d.name
}

func (d *NameID) ID() string {
	return d.id
}

func (d *NameID) ShortID() string {
	return d.id[:10]
}

func (d *NameID) String() string {
	return d.name + " " + d.id
}
