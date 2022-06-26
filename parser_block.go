//go:generate stringer -type TokenType
package cofunc

import (
	"fmt"
	"regexp"

	"github.com/cofunclabs/cofunc/pkg/is"
	"github.com/pkg/errors"
)

// Token
//
type TokenType int

const (
	_unknow_t TokenType = iota
	_int_t
	_text_t
	_mapkey_t
	_operator_t
	_functionname_t
	_load_t
	_word_t
)

var tokenPatterns = map[TokenType]*regexp.Regexp{
	_unknow_t:       regexp.MustCompile(`^*$`),
	_int_t:          regexp.MustCompile(`^[1-9][0-9]*$`),
	_text_t:         regexp.MustCompile(`^*$`),
	_mapkey_t:       regexp.MustCompile(`^[^:]+$`), // not contain ":"
	_operator_t:     regexp.MustCompile(`^=$`),
	_load_t:         regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*:.*[a-zA-Z0-9]$`),
	_functionname_t: regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_\-]*$`),
	_word_t:         regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_\-]*$`),
}

type Token struct {
	value string
	typ   TokenType
	vars  []*struct {
		n    string // var's name
		v    string // var's value, need to read from others
		s, e int    // S is var start position in 'Token.Value', E is end position
	}
}

func newToken(s string, typ TokenType) *Token {
	return &Token{
		value: s,
		typ:   typ,
	}
}

func (t *Token) String() string {
	return t.value
}

func (t *Token) IsEmpty() bool {
	return len(t.value) == 0
}

func (t *Token) HasVar() bool {
	return len(t.vars) != 0
}

func (t *Token) extractVar() error {
	// $(var)
	if t.typ != _text_t {
		return nil
	}
	var (
		pre   rune
		start int
		state stateL2
	)
	l := len(t.value)
	next := func(i int) byte {
		i += 1
		if i >= l {
			return 'x'
		}
		return t.value[i]
	}
	for i, c := range t.value {
		switch state {
		case _l2_unknow:
			// skip
			// transfer
			if pre != '\\' && c == '$' && next(i) == '(' {
				start = i
				state = _l2_word_started
			}
		case _l2_word_started: // from '$'
			// keep
			if is.Word(c) || c == '(' {
				break
			}
			// transfer
			if c == ')' {
				name := t.value[start+2 : i] // start +2: skip "$("
				if name == "" {
					return errors.New("contain invalid var: " + t.value)
				}
				t.vars = append(t.vars, &struct {
					n string
					v string
					s int
					e int
				}{n: name, s: start, e: i + 1}) // currently i is ')'

				state = _l2_unknow
			}
		}
		pre = c
	}
	return nil
}

func (t *Token) validate() error {
	if pattern := tokenPatterns[t.typ]; !pattern.MatchString(t.value) {
		return errors.Errorf("not match: %s:%s", t.value, pattern)
	}
	return nil
}

// Statement
//
type Statement struct {
	lineNum int
	tokens  []*Token
}

func newStatement(ts ...*Token) *Statement {
	stm := &Statement{}
	stm.tokens = append(stm.tokens, ts...)
	return stm
}

func (s *Statement) LastToken() *Token {
	l := len(s.tokens)
	if l == 0 {
		return nil
	}
	return s.tokens[l-1]
}

// Block
//
type blockBody interface {
	Type() string
	Append(o interface{}) error
	GetStatements() []*Statement
	Len() int
}

type rawbody struct {
	lines []*Statement
}

func (r *rawbody) Len() int {
	return len(r.lines)
}

func (r *rawbody) GetStatements() []*Statement {
	return r.lines
}

func (r *rawbody) Type() string {
	return "raw"
}

func (r *rawbody) append(o interface{}) error {
	stm := o.(*Statement)
	r.lines = append(r.lines, stm)
	return nil
}

func (r *rawbody) LastStatement() *Statement {
	l := len(r.lines)
	if l == 0 {
		panic("not found statement")
	}
	return r.lines[l-1]
}

type Block struct {
	kind      Token
	target    Token
	operator  Token
	typevalue Token
	state     stateL2
	child     []*Block
	parent    *Block
	blockBody
}

func (b *Block) String() string {
	if b.blockBody != nil {
		return fmt.Sprintf(`kind="%s", target="%s", operator="%s", tov="%s", bodylen="%d"`, &b.kind, &b.target, &b.operator, &b.typevalue, b.blockBody.Len())
	} else {
		return fmt.Sprintf(`kind="%s", target="%s", operator="%s", tov="%s"`, &b.kind, &b.target, &b.operator, &b.typevalue)
	}
}

// TODO: Call InputVar at running, not parsing
func (b *Block) InputVar() error {
	return nil
}
