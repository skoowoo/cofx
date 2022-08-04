package parser

import "strings"

type Statement struct {
	desc   string
	tokens []*Token
}

func NewStatement(desc string) *Statement {
	stm := &Statement{desc: desc}
	return stm
}

func (s *Statement) FormatString() string {
	var builder strings.Builder
	for _, t := range s.tokens {
		builder.WriteString(t.String())
		builder.WriteString(" ")
	}
	return builder.String()
}

func (s *Statement) LastToken() *Token {
	l := len(s.tokens)
	if l == 0 {
		return nil
	}
	return s.tokens[l-1]
}

func (s *Statement) Append(t *Token) *Statement {
	s.tokens = append(s.tokens, t)
	return s
}

func (s *Statement) Copy() *Statement {
	stm := &Statement{
		desc: s.desc,
	}
	for _, t := range s.tokens {
		nt := &Token{
			str:       t.str,
			typ:       t.typ,
			ln:        t.ln,
			_b:        t._b,
			_segments: t.CopySegments(),
			_get:      t._get,
		}
		stm.tokens = append(stm.tokens, nt)
	}
	return stm
}
