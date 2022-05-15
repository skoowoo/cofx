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
	line = strings.TrimSpace(line)
	words := strings.Fields(line)

	// This is a new line
	if blockstore.parsingBlock != nil {
		blockstore.parsingBlock.parsingBlockLineNum += 1
	}

	for i, wd := range words {
		token := &Token{
			word:         wd,
			seqNumInLine: int8(i + 1),
		}

		switch blockstore.parsingBlockKind {
		case _block_load:
			// eg. load path
			//
			b := blockstore.parsingBlock
			token.BlockLineNum = b.parsingBlockLineNum
			b.tokens[token.BlockLineNum] = append(b.tokens[token.BlockLineNum], token)
		case _block_set:
			b := blockstore.parsingBlock
			token.BlockLineNum = b.parsingBlockLineNum
			b.tokens[token.BlockLineNum] = append(b.tokens[token.BlockLineNum], token)
		case _block_var:

		case _block_run:

		case _block_none:
			// In the global, outside block
			//
			var kind BlockKind
			switch wd {
			case "load":
				kind = _block_load
				token.keyword = true
			case "set":
				kind = _block_set
				token.keyword = true
			case "var":
				kind = _block_var
				token.keyword = true
			case "run":
				kind = _block_run
				token.keyword = true
			default:
				// TODO, extension
				//
				return errors.New("Invalid token word")
			}

			b := &Block{
				tokens:              make(map[BlockLineNum]TokenList),
				kind:                kind,
				parsingBlockLineNum: 1,
			}
			token.BlockLineNum = b.parsingBlockLineNum
			b.tokens[token.BlockLineNum] = append(b.tokens[token.BlockLineNum], token)
			blockstore.Store(b)
			blockstore.parsingBlock = b
			blockstore.parsingBlockKind = kind
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
		blockstore.parsingBlockKind = _block_none
		blockstore.parsingBlock = nil
	case _block_set:
		// block_set is a multi lines block
		if len(all) == 1 && all[0].word == "end" {
			blockstore.parsingBlockKind = _block_none
			blockstore.parsingBlock = nil
			return nil
		}
		// todo, check ...
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
	return errors.New("Invlaid character")
}
