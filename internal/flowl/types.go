package flowl

import (
	"errors"
	"strings"
)

// Map
//
type FlMap struct {
	RawBody
	state parserStateL2
}

func (m *FlMap) ToMap() map[string]string {
	ret := make(map[string]string)
	for _, line := range m.Lines {
		k, v := line.Tokens[0].Value, line.Tokens[1].Value
		ret[k] = v
	}
	return ret
}

func (m *FlMap) Type() string {
	return "map"
}

func (m *FlMap) Append(o interface{}) error {
	const multiflag = "***"
	s := o.(string)
	if m.state == _statel2_multilines_started {
		if strings.HasSuffix(s, multiflag) {
			s = strings.TrimSuffix(s, multiflag)
			m.state = _statel2_unknow
		}
		t := m.LastStatement().LastToken()
		t.Value = t.Value + "\n" + s
	} else {
		idx := strings.Index(s, ":")
		if idx == -1 {
			return errors.New("invalid kv in map: " + s)
		}
		k := strings.TrimSpace(s[0:idx])
		v := strings.TrimSpace(s[idx+1:])
		if strings.HasPrefix(v, multiflag) {
			v = strings.TrimPrefix(v, multiflag)
			m.state = _statel2_multilines_started
			m.Lines = append(m.Lines, NewStatement(k, v))
		} else {
			m.Lines = append(m.Lines, NewStatement(k, v))
		}
	}
	return nil
}

// List
//
type FlList struct {
	RawBody
	EType TokenType
}

func (l *FlList) ToSlice() []string {
	var ret []string
	for _, line := range l.Lines {
		v := line.Tokens[0].Value
		ret = append(ret, v)
	}
	return ret
}

func (l *FlList) Type() string {
	return "list"
}

func (l *FlList) Append(o interface{}) error {
	s := o.(string)
	t := &Token{
		Value: s,
		Type:  l.EType,
	}
	l.Lines = append(l.Lines, NewStatementWithToken(t))
	return nil
}
