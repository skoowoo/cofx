package flowfile

import "container/list"

type Token struct {
	word         string
	keyword      bool
	operator     bool
	fileLineNum  int
	BlockLineNum BlockLineNum
	seqNumInLine int8
}

type BlockLineNum int8
type BlockKind string
type TokenList []*Token

const (
	_block_load   BlockKind = "LOAD"
	_block_set    BlockKind = "SET"
	_block_run    BlockKind = "RUN"
	_block_var    BlockKind = "VAR"
	_block_none   BlockKind = "NONE"
	_block_finish BlockKind = "FINISH"
)

type Block struct {
	tokens map[BlockLineNum]TokenList
	kind   BlockKind

	// for parsing
	parsingBlockLineNum BlockLineNum
}

// Blocklist store all blocks in the flowfile
//
type BlockStore struct {
	l                *list.List
	parsingBlock     *Block
	parsingBlockKind BlockKind
}

func NewBlockStore() *BlockStore {
	return &BlockStore{
		l:                list.New(),
		parsingBlock:     nil,
		parsingBlockKind: _block_none,
	}
}

func (bs *BlockStore) Store(b *Block) {
	bs.l.PushBack(b)
}

func (bs *BlockStore) Done() {
	bs.parsingBlock = nil
	bs.parsingBlockKind = _block_none
}

func (bs *BlockStore) BlockNum() int {
	return bs.l.Len()
}

func (bs *BlockStore) Foreach(do func(*Block)) {
	l := bs.l
	for e := l.Front(); e != nil; e = e.Next() {
		b := e.Value.(*Block)
		do(b)
	}
}

func (bs *BlockStore) String() string {
	return ""
}
