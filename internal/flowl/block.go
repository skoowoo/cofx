package flowl

import (
	"container/list"
	"errors"
	"fmt"
	"strconv"
)

// Token
//
type Token struct {
	text     string
	subtext  []string
	keyword  bool
	operator bool
}

func (t *Token) ToNum() (int, bool) {
	v, err := strconv.Atoi(t.text)
	if err != nil {
		return 0, false
	}
	return v, true
}

func (t *Token) SetKeyWord() *Token {
	t.keyword = true
	t.operator = false
	return t
}

func (t *Token) SetOperator() *Token {
	t.operator = true
	t.keyword = false
	return t
}

// Directive
//
type DirectiveKind string

const (
	_directive_define_block DirectiveKind = "define"
	_directive_finish_block DirectiveKind = "finish"
	_directive_in_block     DirectiveKind = "in"
)

type Directive struct {
	name       string // Defined
	_tokensMin int8   // Defined, Include the directive token
	_tokensMax int8   // Defined, Include the directive token

	tokens    []*Token
	kind      DirectiveKind // Defined
	blockKind BlockKind
	Verify    func(*Directive) error
}

func (d *Directive) Init() error {
	name := d.First().subtext[0]
	def, ok := directiveDefines[name]
	if !ok {
		return errors.New("init the directive failed: " + name)
	}
	d.name = name
	d._tokensMin = def._tokensMin
	d._tokensMax = def._tokensMax
	d.kind = def.kind
	d.blockKind = def.blockKind
	d.Verify = def.Verify

	if err := d.Verify(d); err != nil {
		return err
	}
	return nil
}

func (d *Directive) Put(t *Token) int {
	d.tokens = append(d.tokens, t)
	return len(d.tokens)
}

func (d *Directive) SetBlockKind(k BlockKind) {
	d.blockKind = k
}

func (d *Directive) IsDefineBlock() bool {
	return d.kind == _directive_define_block
}

func (d *Directive) Last() *Token {
	l := len(d.tokens)
	if l == 0 {
		return nil
	}
	return d.tokens[l-1]
}

func (d *Directive) First() *Token {
	l := len(d.tokens)
	if l == 0 {
		return nil
	}
	return d.tokens[0]
}

func (d *Directive) String() string {
	return fmt.Sprintf("directive - name:%s, kind:%s, bkind:%s, min:%d, max: %d, len:%d", d.name, d.kind, d.blockKind, d._tokensMin, d._tokensMax, len(d.tokens))
}

var directiveDefines = map[string]Directive{
	"load":  {"load", 2, 2, nil, _directive_define_block, _block_load, verifyLoad},
	"set":   {"set", 3, 3, nil, _directive_define_block, _block_set, verifySet},
	"run":   {"run", 2, 2, nil, _directive_define_block, _block_run, verifyRun},
	"input": {"input", 3, 3, nil, _directive_in_block, _block_set, verifyInput},
	"loop":  {"loop", 3, 3, nil, _directive_in_block, _block_set, verifyLoop},
	"}":     {"}", 1, 1, nil, _directive_finish_block, _block_none, verifyEndBlock},
	"@":     {"@", 1, 1, nil, _directive_in_block, _block_run, verifyAt},
}

func IsKeyword(s string) bool {
	_, ok := directiveDefines[s]
	return ok
}

// Maybe panic()
func Keyword(s string) string {
	if !IsKeyword(s) {
		panic("not a keyword: " + s)
	}
	return s
}

func verifybase(dir *Directive) error {
	l := len(dir.tokens)
	if l < int(dir._tokensMin) || l > int(dir._tokensMax) {
		return errors.New("too many tokens: " + dir.name)
	}
	return nil
}

func verifyLoad(dir *Directive) error {
	if err := verifybase(dir); err != nil {
		return err
	}
	return nil
}

func verifySet(dir *Directive) error {
	if err := verifybase(dir); err != nil {
		return err
	}
	return nil
}

func verifyRun(dir *Directive) error {
	if err := verifybase(dir); err != nil {
		return err
	}
	return nil
}

func verifyInput(dir *Directive) error {
	if err := verifybase(dir); err != nil {
		return err
	}
	return nil
}

func verifyLoop(dir *Directive) error {
	if err := verifybase(dir); err != nil {
		return err
	}
	return nil
}

func verifyEndBlock(dir *Directive) error {
	if err := verifybase(dir); err != nil {
		return err
	}
	return nil
}

func verifyAt(dir *Directive) error {
	if err := verifybase(dir); err != nil {
		return err
	}
	return nil
}

// Block
//
type BlockKind string
type BlockStatus string

const (
	_block_load BlockKind = "load"
	_block_set  BlockKind = "set"
	_block_run  BlockKind = "run"
	_block_var  BlockKind = "var"
	_block_none BlockKind = "none"
)

const (
	_block_status_begin  BlockStatus = "begin"
	_block_status_end    BlockStatus = "end"
	_block_status_unknow BlockStatus = "unknow"
)

type Block struct {
	directives []*Directive
	kind       BlockKind
	bracket    int
	status     BlockStatus
}

func Newblock(k BlockKind) *Block {
	return &Block{
		kind:    k,
		bracket: 0,
		status:  _block_status_unknow,
	}
}

func (b *Block) Put(dir *Directive) (BlockStatus, error) {
	if dir.blockKind != _block_none && dir.blockKind != b.kind {
		return b.status, fmt.Errorf("directive(%s) can't appear in the block", dir.name)
	}
	first, last := dir.First(), dir.Last()
	if first != nil && last != nil {
		if last.text == "{" {
			b.bracket += 1
		}
		if last.text == "}" && first.text == "}" {
			b.bracket -= 1
		}
	}
	if b.bracket > 0 {
		b.status = _block_status_begin
	} else if b.bracket == 0 {
		b.status = _block_status_end
	} else if b.bracket < 0 {
		return b.status, errors.New("invalid syntax: too many }")
	}
	b.directives = append(b.directives, dir)
	return b.status, nil
}

func (b *Block) SetKind(k BlockKind) {
	b.kind = k
}

func (b *Block) GetKind() BlockKind {
	return b.kind
}

func (b *Block) String() string {
	return fmt.Sprintf("block - kind:%s, bracket:%d, status:%s, len:%d", b.kind, b.bracket, b.status, len(b.directives))
}

// Blocklist store all blocks in the flowl
//
type BlockStore struct {
	l       *list.List
	parsing *Block
}

func NewBlockStore() *BlockStore {
	return &BlockStore{
		l:       list.New(),
		parsing: nil,
	}
}

func (bs *BlockStore) PutAndSetParsing(b *Block) error {
	if bs.parsing != nil {
		return errors.New("cann't put a block into blockstore, because the parsing block is not finished")
	}
	bs.l.PushBack(b)
	bs.parsing = b
	return nil
}

func (bs *BlockStore) ParsingBlock() *Block {
	return bs.parsing
}

func (bs *BlockStore) FinishParsingBlock() error {
	bs.parsing = nil
	return nil
}

func (bs *BlockStore) ParsingBlockIsFinished() bool {
	return bs.parsing == nil
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
