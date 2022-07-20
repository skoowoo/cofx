package cofunc

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
		return StatementTokensErrorf(ErrMapKVIllegal, ts)
	}
	k, delim, v := ts[0], ts[1], ts[2]
	if k.typ != _string_t {
		return TokenTypeErrorf(k, _string_t)
	}
	if delim.typ != _symbol_t {
		return TokenTypeErrorf(delim, _symbol_t)
	}
	if delim.String() != ":" {
		return TokenValueErrorf(delim, ":")
	}
	if v.typ != _string_t {
		return TokenTypeErrorf(k, _string_t)
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
		return StatementTokensErrorf(ErrListElemIllegal, ts)
	}
	t := ts[0]
	t.typ = l.etype
	l.lines = append(l.lines, newstm("element").Append(t))
	return nil
}
