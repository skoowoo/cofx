package cofunc

import (
	"errors"
	"strings"
)

const (
	_map  = "map"
	_list = "list"
)

type FMap struct {
	plainbody
	state stateL2
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
	const multiline = "***"
	s := o.(string)
	if m.state == _l2_multilines_started {
		if strings.HasSuffix(s, multiline) {
			s = strings.TrimSuffix(s, multiline)
			m.state = _l2_unknow
		}
		t := m.Laststm().LastToken()
		t.str = t.str + "\n" + s
	} else {
		if s == "" {
			return nil
		}
		idx := strings.Index(s, ":")
		if idx == -1 {
			return errors.New("invalid kv in map: " + s)
		}
		k := strings.TrimSpace(s[0:idx])
		v := strings.TrimSpace(s[idx+1:])
		if strings.HasPrefix(v, multiline) {
			v = strings.TrimPrefix(v, multiline)
			m.state = _l2_multilines_started
			m.lines = append(m.lines, newstm("kv").Append(newToken(k, _mapkey_t)).Append(newToken(v, _text_t)))
		} else {
			m.lines = append(m.lines, newstm("kv").Append(newToken(k, _mapkey_t)).Append(newToken(v, _text_t)))
		}
	}
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
	s := o.(string)
	t := &Token{
		str: s,
		typ: l.etype,
	}
	l.lines = append(l.lines, newstm("element").Append(t))
	return nil
}
