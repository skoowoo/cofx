package flowfile

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

// Parse parse a 'flowfile'
func Parse(rd io.Reader) error {
	return nil
}

func ParseBlocks(rd io.Reader) (*BlockStore, error) {
	blockstore := NewBlockStore()

	scanner := bufio.NewScanner(rd)
	for {
		if !scanner.Scan() {
			break
		}
		err := splitLineToTokens(blockstore, scanner.Text())
		if err != nil {
			return nil, err
		}
	}
	return blockstore, nil
}

func splitLineToTokens(blockstore *BlockStore, line string) error {
	if (blockstore.parsingBlockKind == _block_none && blockstore.parsingBlock != nil) || (blockstore.parsingBlockKind != _block_none && blockstore.parsingBlock == nil) {
		panic("Fatal error")
	}

	line = strings.TrimSpace(line)
	words := strings.Fields(line)
	if len(words) == 0 {
		// This is empty line, so skip it
		return nil
	}

	// This is a new line
	if blockstore.parsingBlock != nil {
		blockstore.parsingBlock.parsingBlockLineNum += 1
	}

	for i, wd := range words {
		token := &Token{
			word:         wd,
			seqNumInLine: int8(i + 1),
		}
		var (
			kind     BlockKind
			newblock bool
		)
		switch wd {
		case "load":
			kind = _block_load
			newblock = true
		case "set":
			kind = _block_set
			newblock = true
		case "var":
			kind = _block_var
			newblock = true
		case "run":
			kind = _block_run
			newblock = true
		}

		if newblock {
			if !blockstore.ParsingBlockIsFinished() {
				return errors.New("parsing block is not finished, can't create a new block")
			}

			b := &Block{
				tokens:              make(map[BlockLineNum]TokenList),
				kind:                kind,
				parsingBlockLineNum: 1,
			}
			blockstore.Store(b)
			blockstore.parsingBlock = b
			blockstore.parsingBlockKind = kind

			token.BlockLineNum = b.parsingBlockLineNum
			token.keyword = true
			blockstore.AppendToParsingBlock(token)
			continue
		}

		switch blockstore.parsingBlockKind {
		case _block_load:
			// eg. load path
			//
			token.BlockLineNum = blockstore.parsingBlock.parsingBlockLineNum
			blockstore.AppendToParsingBlock(token)
		case _block_set:
			token.BlockLineNum = blockstore.parsingBlock.parsingBlockLineNum
			blockstore.AppendToParsingBlock(token)
		case _block_var:

		case _block_run:
			token.BlockLineNum = blockstore.parsingBlock.parsingBlockLineNum
			blockstore.AppendToParsingBlock(token)
		case _block_none:
			return errors.New("token not in a block: " + wd)
		}
	}

	// Check if all tokens of the line are valid
	//
	var all TokenList
	if b := blockstore.parsingBlock; b != nil {
		all = b.tokens[b.parsingBlockLineNum]
		if err := verifyAllTokensInLine(all); err != nil {
			return err
		}
	}

	// The lind ends, check the state
	switch blockstore.parsingBlockKind {
	case _block_load:
		// block_load is a single line block, close it
		blockstore.FinishParsingBlock()
	case _block_run:
		blockstore.FinishParsingBlock()
	case _block_set:
		// block_set is a multi lines block
		if len(all) == 1 && all[0].word == "end" {
			blockstore.FinishParsingBlock()
			return nil
		}
	}
	return nil
}

// keywords utils
//

// character utils
//
func IsBlank(b byte) bool {
	return b == '\r' || b == '\n' || b == '\t' || b == ' '
}

func IsSpace(b byte) bool {
	return b == '\t' || b == ' '
}

func IsNewLine(b byte) bool {
	return b == '\n'
}

func IsInvalid(b byte) error {
	if IsBlank(b) {
		return nil
	}
	switch {
	case b >= 'a' && b <= 'z':
		return nil
	case b >= 'A' && b <= 'Z':
		return nil
	case b == '@':
		return nil
	case b == '$':
		return nil
	}
	return errors.New("invlaid character")
}
