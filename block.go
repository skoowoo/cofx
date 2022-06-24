//go:generate stringer -type TokenType
//go:generate stringer -type BlockLevel
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
	UnknowT TokenType = iota
	IntT
	TextT
	MapKeyT
	OperatorT
	FunctionNameT
	LoadT
)

var tokenPatterns = map[TokenType]*regexp.Regexp{
	UnknowT:       regexp.MustCompile(`^*$`),
	IntT:          regexp.MustCompile(`^[1-9][0-9]*$`),
	TextT:         regexp.MustCompile(`^*$`),
	MapKeyT:       regexp.MustCompile(`^[^:]+$`), // not contain ":"
	OperatorT:     regexp.MustCompile(`^=$`),
	LoadT:         regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9]*:.*[a-zA-Z0-9]$`),
	FunctionNameT: regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_\-]*$`),
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

func NewTextToken(s string) *Token {
	return NewToken(s, TextT)
}

func NewToken(s string, typ TokenType) *Token {
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

// TODO: when running
func (t *Token) AssignVar(b *Block) error {
	return nil
}

func (t *Token) HasVar() bool {
	return len(t.vars) != 0
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

func NewStatement(ss ...string) *Statement {
	stm := &Statement{}
	for _, s := range ss {
		stm.tokens = append(stm.tokens, NewTextToken(s))
	}
	return stm
}

func NewStatementWithToken(ts ...*Token) *Statement {
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
type BlockBody interface {
	Type() string
	Append(o interface{}) error
	Statements() []*Statement
	Len() int
}

type RawBody struct {
	lines []*Statement
}

func (r *RawBody) Len() int {
	return len(r.lines)
}

func (r *RawBody) Statements() []*Statement {
	return r.lines
}

func (r *RawBody) Type() string {
	return "raw"
}

func (r *RawBody) Append(o interface{}) error {
	stm := o.(*Statement)
	r.lines = append(r.lines, stm)
	return nil
}

func (r *RawBody) LastStatement() *Statement {
	l := len(r.lines)
	if l == 0 {
		panic("not found statement")
	}
	return r.lines[l-1]
}

type BlockLevel int

const (
	LevelGlobal BlockLevel = iota
	LevelParent
	LevelChild
)

type Block struct {
	kind      Token
	target    Token
	operator  Token
	typevalue Token
	state     parserStateL2
	level     BlockLevel
	child     []*Block
	parent    *Block
	BlockBody
}

func (b *Block) String() string {
	if b.BlockBody != nil {
		return fmt.Sprintf(`kind="%s", target="%s", operator="%s", tov="%s", bodylen="%d"`, &b.kind, &b.target, &b.operator, &b.typevalue, b.BlockBody.Len())
	} else {
		return fmt.Sprintf(`kind="%s", target="%s", operator="%s", tov="%s"`, &b.kind, &b.target, &b.operator, &b.typevalue)
	}
}

// TODO: Call InputVar at running, not parsing
func (b *Block) InputVar() error {
	return nil
}

// AST store all blocks in the flowl
//
type AST struct {
	global Block

	parsing  *Block
	state    parserStateL1
	prestate parserStateL1
}

func NewAST() *AST {
	return &AST{
		global: Block{
			child: make([]*Block, 0),
			level: LevelGlobal,
		},
		parsing: nil,
		state:   _statel1_global,
	}
}

func deepwalk(b *Block, do func(*Block) error) error {
	// skip the global block
	// if b.Parent != nil {
	// 	if err := do(b); err != nil {
	// 		return err
	// 	}
	// }
	if b.level != LevelGlobal {
		if err := do(b); err != nil {
			return err
		}
	}
	for _, c := range b.child {
		if err := deepwalk(c, do); err != nil {
			return err
		}
	}
	return nil
}

func (a *AST) Foreach(do func(*Block) error) error {
	return deepwalk(&a.global, do)
}

func (a *AST) String() string {
	return ""
}
