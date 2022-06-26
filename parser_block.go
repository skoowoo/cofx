//go:generate stringer -type TokenType
package cofunc

import (
	"fmt"
	"regexp"

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
)

var tokenPatterns = map[TokenType]*regexp.Regexp{
	_unknow_t:       regexp.MustCompile(`^*$`),
	_int_t:          regexp.MustCompile(`^[1-9][0-9]*$`),
	_text_t:         regexp.MustCompile(`^*$`),
	_mapkey_t:       regexp.MustCompile(`^[^:]+$`), // not contain ":"
	_operator_t:     regexp.MustCompile(`^=$`),
	_load_t:         regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*:.*[a-zA-Z0-9]$`),
	_functionname_t: regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_\-]*$`),
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

func newTextToken(s string) *Token {
	return newToken(s, _text_t)
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

// TODO: when running
func (t *Token) assignVar(b *Block) error {
	return nil
}

func (t *Token) Validate() error {
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

func newStatement(ss ...string) *Statement {
	stm := &Statement{}
	for _, s := range ss {
		stm.tokens = append(stm.tokens, newTextToken(s))
	}
	return stm
}

func newStatementWithToken(ts ...*Token) *Statement {
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
