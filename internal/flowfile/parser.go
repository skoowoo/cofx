package flowfile

import (
	"bufio"
	"bytes"
	"container/list"
	"io"
)

type NameType string
type LoadPathType string
type BlockType string

const (
	LOAD_BLOCK BlockType = "LOAD"
	SET_BLOCK  BlockType = "SET"
	RUN_BLOCK  BlockType = "RUN"
	VAR_BLOCK  BlockType = "VAR"
	NONE_BLOCK BlockType = "NONE"
)

type ActionConfig struct {
	name      NameType
	aliasName NameType
	path      LoadPathType
}

// Parse parse a 'flowfile'
func Parse(rd io.Reader) (runq *list.List, actions map[NameType]*ActionConfig, err error) {
	// run queue of a flow
	runq = list.New()
	if runq == nil {
		panic("new list error")
	}
	// save action's configs of a flow in Flowfile
	actions = make(map[NameType]*ActionConfig)

	currentBlock := NONE_BLOCK

	scanner := bufio.NewScanner(rd)
	for {
		if !scanner.Scan() {
			break
		}
		currentBlock = parseLine(runq, actions, currentBlock, scanner.Bytes())
	}

	return
}

func parseLine(runq *list.List, actions map[NameType]*ActionConfig, block BlockType, line []byte) BlockType {
	line = bytes.TrimSpace(line)
	words := bytes.Fields(line)
	if len(words) == 0 {
		return block
	}
	for _, wd := range words {
		switch block {
		case LOAD_BLOCK:
			// find the action's load path, and parse it
			name, path := parseLoadPath(string(w))

		case SET_BLOCK:
		case RUN_BLOCK:
		case NONE_BLOCK:
			// In the global, try to find the keywords
			//
			keyword := string(wd)

			switch keyword {
			case "load":
				block = LOAD_BLOCK
			case "set":
				block = SET_BLOCK
			case "run":
				block = RUN_BLOCK
			case "var":
				block = VAR_BLOCK
			default:
				// todo, error
			}
		}
	}
}

// load path utils
func parseLoadPath(path string) (name NameType, path LoadPathType, err error) {
	return
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
	return newInvalidCharacter()
}
