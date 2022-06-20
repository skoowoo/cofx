//go:generate stringer -type TokenType
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
	FunctionNodeNameT
	LoadT
)

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
	pattern := `*`
	switch t.Type {
	case FunctionNodeNameT:
		pattern = `^[a-zA-Z][a-zA-Z0-9]*$`
	default:
		return nil
	}
	match, err := regexp.MatchString(pattern, t.Value)
	if err != nil {
		return errors.Wrapf(err, "not match: %s:%s", t.Value, pattern)
	}
	if !match {
		return errors.Errorf("not match: %s:%s", t.Value, pattern)
	}
	return nil
}

// Block
//
type BlockLevel int

const (
	LevelGlobal BlockLevel = iota
	LevelParent
	LevelChild
)

type Block struct {
	Kind     Token
	Receiver Token
	Symbol   Token
	Object   Token
	state    parserStateL2
	Level    BlockLevel
	Child    []*Block
	Parent   *Block
	BlockBody
}

func (b *Block) String() string {
	return fmt.Sprintf("kind:%s, receriver:%s, symbol:%s, object:%s", &b.Kind, &b.Receiver, &b.Symbol, &b.Object)
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
