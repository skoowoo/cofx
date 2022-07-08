package cofunc

import (
	"errors"
)

const (
	_map  = "map"
	_list = "list"
)

type FMap struct {
	plainbody
}

func (m *FMap) ToMap() map[string]string {
	ret := make(map[string]string)
	for _, ln := range m.lines {
		k, v := ln.tokens[0].Value(), ln.tokens[1].Value()
		ret[k] = v
	}
	return ret
}

func (m *FMap) Type() string {
	return _map
}

func (m *FMap) Append(o interface{}) error {
	ts := o.([]*Token)
	if len(ts) != 3 {
		return errors.New("invalid kv in map")
	}
	k, delim, v := ts[0], ts[1], ts[2]
	if k.typ != _string_t || delim.typ != _symbol_t || delim.String() != ":" || v.typ != _string_t {
		return errors.New("invalid kv in map")
	}
	m.lines = append(m.lines, newstm("kv").Append(k).Append(v))

	return nil
}

type FList struct {
	plainbody
	etype TokenType
}

func (l *FList) ToSlice() []string {
	var ret []string
	for _, ln := range l.lines {
		v := ln.tokens[0].Value()
		ret = append(ret, v)
	}
	return ret
}

func (l *FList) Type() string {
	return _list
}

func (l *FList) Append(o interface{}) error {
	ts := o.([]*Token)
	if len(ts) != 1 {
		return errors.New("invalid list element")
	}
	t := ts[0]
	t.typ = l.etype
	l.lines = append(l.lines, newstm("element").Append(t))
	return nil
}
