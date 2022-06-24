package cofunc

import (
	"errors"
	"strings"
)

// Map
//
type FMap struct {
	RawBody
	state parserStateL2
}

func (m *FMap) ToMap() map[string]string {
	ret := make(map[string]string)
	for _, line := range m.lines {
		k, v := line.tokens[0].value, line.tokens[1].value
		ret[k] = v
	}
	return ret
}

func (m *FMap) Type() string {
	return "map"
}

func (m *FMap) Append(o interface{}) error {
	const multiline = "***"
	s := o.(string)
	if m.state == _statel2_multilines_started {
		if strings.HasSuffix(s, multiline) {
			s = strings.TrimSuffix(s, multiline)
			m.state = _statel2_unknow
		}
		t := m.LastStatement().LastToken()
		t.value = t.value + "\n" + s
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
			m.state = _statel2_multilines_started
			m.lines = append(m.lines, NewStatement(k, v))
		} else {
			m.lines = append(m.lines, NewStatement(k, v))
		}
	}
	return nil
}

// List
//
type FList struct {
	RawBody
	etype TokenType
}

func (l *FList) ToSlice() []string {
	var ret []string
	for _, line := range l.lines {
		v := line.tokens[0].value
		ret = append(ret, v)
	}
	return ret
}

func (l *FList) Type() string {
	return "list"
}

func (l *FList) Append(o interface{}) error {
	s := o.(string)
	t := &Token{
		value: s,
		typ:   l.etype,
	}
	l.lines = append(l.lines, NewStatementWithToken(t))
	return nil
}
