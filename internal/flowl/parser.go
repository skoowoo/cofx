package flowl

import (
	"bufio"
	"errors"
	"io"
	"strings"
)

// Parse parse a 'flowl' file
func ParseFile(file string) error {
	return nil
}

func Parse(rd io.Reader) (runq *RunQueue, bs *BlockStore, err error) {
	if bs, err = ParseBlocks(rd); err != nil {
		return
	}
	runq = NewRunQueue()
	if err = runq.Generate(bs); err != nil {
		return
	}
	return
}

func ParseBlocks(rd io.Reader) (*BlockStore, error) {
	blockstore := NewBlockStore()

	scanner := bufio.NewScanner(rd)
	for {
		if !scanner.Scan() {
			break
		}
		err := processOneLine(blockstore, scanner.Text())
		if err != nil {
			return nil, err
		}
	}
	return blockstore, nil
}

func processOneLine(bs *BlockStore, line string) error {
	line = strings.TrimSpace(line)
	words := strings.Fields(line)
	if len(words) == 0 {
		// This is empty line, so skip it
		return nil
	}

	dir := &Directive{}
	for _, wd := range words {
		var subtext []string
		if strings.HasPrefix(wd, Keyword("@")) {
			subtext = []string{Keyword("@"), strings.TrimPrefix(wd, Keyword("@"))}
		} else {
			subtext = []string{wd}
		}
		t := &Token{
			text:    wd,
			subtext: subtext,
		}
		if IsKeyword(t.text) {
			t.SetKeyWord()
		}
		dir.Put(t)
	}
	if err := dir.Init(); err != nil {
		return err
	}

	var block *Block
	if dir.IsDefineBlock() {
		// new block
		block = &Block{}
		block.SetKind(dir.blockKind)
		if _, err := block.Put(dir); err != nil {
			return err
		}
		if err := bs.PutAndSetParsing(block); err != nil {
			return err
		}
	} else {
		// in a block
		block = bs.ParsingBlock()
		if block == nil {
			return errors.New("invalid directive: " + dir.name)
		}
		if _, err := block.Put(dir); err != nil {
			return err
		}
	}

	// The lind ends, check the state
	if bs.ParsingBlock().status == _block_status_end {
		bs.FinishParsingBlock()
	}
	return nil
}
