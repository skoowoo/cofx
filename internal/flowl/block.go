//go:generate stringer -type TokenType
//go:generate stringer -type BlockLevel
package flowl

import (
	"container/list"
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

var tokenPatterns = map[TokenType]string{
	UnknowT:       `^*$`,
	IntT:          `^*$`,
	TextT:         `^*$`,
	MapKeyT:       `^*$`,
	OperatorT:     `^*$`,
	LoadT:         `^*$`,
	FunctionNameT: `^[a-zA-Z][a-zA-Z0-9]*$`,
}

type Token struct {
	Value string
	Type  TokenType
}

func NewTextToken(s string) *Token {
	return NewToken(s, TextT)
}

func NewToken(s string, typ TokenType) *Token {
	return &Token{
		Value: s,
		Type:  typ,
	}
}

func (t *Token) String() string {
	return t.Value
}

func (t *Token) IsEmpty() bool {
	return len(t.Value) == 0
}

func (t *Token) Validate() error {
	pattern := tokenPatterns[t.Type]
	match, err := regexp.MatchString(pattern, t.Value)
	if err != nil {
		return errors.Wrapf(err, "validate token")
	}
	if !match {
		return errors.Errorf("not match: %s:%s", t.Value, pattern)
	}
	return nil
}

// Statement
//
type Statement struct {
	LineNum int
	Tokens  []*Token
}

func NewStatement(ss ...string) *Statement {
	stm := &Statement{}
	for _, s := range ss {
		stm.Tokens = append(stm.Tokens, NewTextToken(s))
	}
	return stm
}

func NewStatementWithToken(ts ...*Token) *Statement {
	stm := &Statement{}
	stm.Tokens = append(stm.Tokens, ts...)
	return stm
}

func (s *Statement) LastToken() *Token {
	l := len(s.Tokens)
	if l == 0 {
		return nil
	}
	return s.Tokens[l-1]
}

// Block
//
type BlockBody interface {
	Type() string
	Append(o interface{}) error
	Statements() []*Statement
}

type RawBody struct {
	Lines []*Statement
}

func (r *RawBody) Statements() []*Statement {
	return r.Lines
}

func (r *RawBody) Type() string {
	return "raw"
}

func (r *RawBody) Append(o interface{}) error {
	stm := o.(*Statement)
	r.Lines = append(r.Lines, stm)
	return nil
}

func (r *RawBody) LastStatement() *Statement {
	l := len(r.Lines)
	if l == 0 {
		panic("not found statement")
	}
	return r.Lines[l-1]
}

type BlockLevel int

const (
	LevelGlobal BlockLevel = iota
	LevelParent
	LevelChild
)

type Block struct {
	Kind        Token
	Target      Token
	Operator    Token
	TypeOrValue Token
	state       parserStateL2
	Level       BlockLevel
	Child       []*Block
	Parent      *Block
	BlockBody
}

func (b *Block) String() string {
	return fmt.Sprintf("kind:%s, receriver:%s, symbol:%s, object:%s", &b.Kind, &b.Target, &b.Operator, &b.TypeOrValue)
}

// Blocklist store all blocks in the flowl
//
type BlockStore struct {
	l        *list.List
	parsing  *Block
	state    parserStateL1
	prestate parserStateL1
}

func NewBlockStore() *BlockStore {
	return &BlockStore{
		l:       list.New(),
		parsing: nil,
		state:   _statel1_global,
	}
}

func (bs *BlockStore) Foreach(do func(*Block) error) error {
	l := bs.l
	for e := l.Front(); e != nil; e = e.Next() {
		b := e.Value.(*Block)
		if err := do(b); err != nil {
			return err
		}
	}
	return nil
}

func (bs *BlockStore) String() string {
	return ""
}
