//go:generate stringer -type stateL1
//go:generate stringer -type stateL2
package cofunc

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/cofunclabs/cofunc/pkg/is"
)

func ParseAST(rd io.Reader) (*AST, error) {
	ast := newAST()
	num := 0
	scanner := bufio.NewScanner(rd)
	for {
		if !scanner.Scan() {
			break
		}
		num += 1
		err := scanToken(ast, scanner.Text(), num)
		if err != nil {
			return nil, err
		}
	}

	return ast, ast.Foreach(func(b *Block) error {
		if err := b.kind.extractVar(); err != nil {
			return err
		}
		if err := b.kind.validate(); err != nil {
			return err
		}

		if err := b.target.extractVar(); err != nil {
			return err
		}
		if err := b.target.validate(); err != nil {
			return err
		}

		if err := b.operator.extractVar(); err != nil {
			return err
		}
		if err := b.operator.validate(); err != nil {
			return err
		}

		if err := b.typevalue.extractVar(); err != nil {
			return err
		}
		if err := b.typevalue.validate(); err != nil {
			return err
		}

		if b.blockBody != nil {
			lines := b.blockBody.GetStatements()
			for _, ln := range lines {
				for _, t := range ln.tokens {
					if err := t.extractVar(); err != nil {
						return err
					}
					if err := t.validate(); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
}

const (
	_l1_global stateL1 = iota
	_l1_block_kind
	_l1_load_started
	_l1_run_started
	_l1_run_body_started
	_l1_run_body_inside
	_l1_fn_started
	_l1_fn_body_started
	_l1_fn_body_inside
	_l1_args_started
	_l1_args_body_started
	_l1_args_body_inside
)

const (
	_l2_unknow stateL2 = iota
	_l2_multilines_started
	_l2_word_started
	_l2_kind_started
	_l2_kind_done
	_l2_target_started
	_l2_target_done
	_l2_operator_started
	_l2_operator_done
	_l2_typevalue_started
)

func scanToken(ast *AST, line string, linenum int) error {
	block := ast.parsing

	var start int

	finiteAutomata := func(last int, current rune, newline string) error {
		switch ast.phase() {
		case _l1_global:
			// skip
			if is.SpaceOrEOL(current) {
				break
			}
			// transfer
			if is.Word(current) {
				start = last
				ast.transfer(_l1_block_kind)
				break
			}
			// error
			return errors.New("contain invalid character: " + newline)
		case _l1_block_kind:
			// keep
			if is.Word(current) {
				break
			}
			// transfer
			if is.Space(current) || is.LeftBracket(current) {
				var body blockBody = nil
				word := newline[start:last]
				switch word {
				case "load":
					if is.LeftBracket(current) {
						return errors.New("contain invalid character: " + newline)
					}
					ast.transfer(_l1_load_started)
				case "fn":
					if is.LeftBracket(current) {
						return errors.New("contain invalid character: " + newline)
					}
					ast.transfer(_l1_fn_started)
				case "run":
					if is.LeftBracket(current) {
						body = &FList{etype: _functionname_t}
						ast.transfer(_l1_run_body_started)
					} else {
						ast.transfer(_l1_run_started)
					}
				default:
					return errors.New("invalid block define: " + word)
				}
				block = &Block{
					kind: Token{
						str: word,
						typ: _word_t,
					},
					state:     _l2_kind_done,
					parent:    &ast.global,
					blockBody: body,
				}
				ast.global.child = append(ast.global.child, block)
				ast.parsing = block
				break
			}
			// error
			return errors.New("contain invalid character: " + newline)
		case _l1_load_started:
			///
			// load go:sleep
			//
			switch block.state {
			case _l2_kind_done:
				// skip
				if is.Space(current) {
					break
				}
				// transfer
				if is.Word(current) {
					start = last
					block.state = _l2_target_started
					break
				}
				// error
				return errors.New("contain invalid character: " + newline)
			case _l2_target_started:
				// transfer
				if is.EOL(current) {
					block.target = Token{
						str: strings.TrimSpace(newline[start:last]),
						typ: _load_t,
					}
					block.state = _l2_target_done
					ast.transfer(_l1_global)
					break
				}
			case _l2_target_done:

			}
		case _l1_run_started:
			/**
			 1. run sleep
			 2. run {
			 		f1
					f2
			 	}
			3. run sleep {
				time: 1s
			}
			*/
			switch block.state {
			case _l2_kind_done:
				// skip
				if is.Space(current) {
					break
				}
				// transfer 1
				if is.Word(current) {
					start = last
					block.state = _l2_target_started
					break
				}
				if is.LeftBracket(current) {
					/*
						run {
							f1
							f2
						}
					*/
					block.blockBody = &FList{etype: _functionname_t}
					ast.transfer(_l1_run_body_started)
					break
				}
				// error
				return errors.New("contain invalid character: " + newline)
			case _l2_target_started:
				// keep
				if is.Word(current) {
					break
				}
				// 1. transfer - run sleep{
				if is.LeftBracket(current) {
					block.target = Token{
						str: strings.TrimSpace(newline[start:last]),
						typ: _functionname_t,
					}
					block.blockBody = &FMap{}
					block.state = _l2_unknow
					ast.transfer(_l1_run_body_started)
					break
				}
				// 2. transfer - run sleep {  or run sleep
				if is.Space(current) {
					block.target = Token{
						str: strings.TrimSpace(newline[start:last]),
						typ: _functionname_t,
					}
					block.state = _l2_target_done
					break
				}
				// 3. transfer run sleep
				if is.EOL(current) {
					block.target = Token{
						str: strings.TrimSpace(newline[start:last]),
						typ: _functionname_t,
					}
					block.state = _l2_unknow
					ast.transfer(_l1_global)
					break
				}
				return errors.New("contain invalid character: " + newline)
			case _l2_target_done:
				// transfer
				if is.EOL(current) {
					ast.transfer(_l1_global)
					break
				}
				if is.LeftBracket(current) {
					block.blockBody = &FMap{}
					ast.transfer(_l1_run_body_started)
					break
				}
				// skip
				if is.Space(current) {
					break
				}
				// error
				return errors.New("contain invalid character: " + newline)
			}

		case _l1_run_body_started:
			// transfer
			if is.EOL(current) {
				ast.transfer(_l1_run_body_inside)
				break
			}
			// skip
			if is.Space(current) {
				break
			}
			return errors.New("invalid run block: " + newline + fmt.Sprintf(" (%c)", current))
		case _l1_run_body_inside:
			// 1. k: v
			// 2. f
			// 3. }
			if is.EOL(current) {
				if newline == "}" {
					ast.transfer(_l1_global)
				} else if newline != "" {
					if err := block.blockBody.Append(newline); err != nil {
						return err
					}
				}
			}

		case _l1_fn_started:
			/*
				fn f1 = f {

				}
			*/
			switch block.state {
			case _l2_kind_done:
				// skip
				if is.Space(current) {
					break
				}
				// transfer
				if is.Word(current) {
					start = last
					block.state = _l2_target_started
					break
				}
				// error
				return errInvalidChar(byte(current), newline)
			case _l2_target_started: // from '{word}'
				// keep
				if is.Word(current) {
					break
				}
				if is.Space(current) {
					break
				}
				// transfer
				if is.Eq(current) {
					s := newline[start:last]
					block.target = Token{
						str: strings.TrimSpace(s),
						typ: _word_t,
					}
					block.operator = Token{
						str: "=",
						typ: _operator_t,
					}
					block.state = _l2_operator_started
					break
				}
				// error
				return errInvalidChar(byte(current), newline)
			case _l2_operator_started: // from '='
				// skip
				if is.Space(current) {
					break
				}
				// transfer
				if is.Word(current) {
					start = last
					block.state = _l2_typevalue_started
					break
				}
				// error
				return errInvalidChar(byte(current), newline)
			case _l2_typevalue_started: // from '{word}'
				// keep
				if is.Word(current) {
					break
				}
				if is.Space(current) {
					break
				}
				// transfer
				if is.LeftBracket(current) {
					s := newline[start:last]
					block.typevalue = Token{
						str: strings.TrimSpace(s),
						typ: _functionname_t,
					}
					block.state = _l2_unknow
					ast.transfer(_l1_fn_body_started)
					break
				}
				// error
				return errInvalidChar(byte(current), newline)
			}
		case _l1_fn_body_started: // from '{'
			// skip
			if is.Space(current) {
				break
			}
			// transfer
			if is.EOL(current) {
				block.state = _l2_unknow
				ast.transfer(_l1_fn_body_inside)
				break
			}
			// error
			return errInvalidChar(byte(current), newline)
		case _l1_fn_body_inside: // from '\n'
			if block.state == _l2_word_started {
				if unicode.IsSpace(current) || current == '=' {
					block.state = _l2_unknow
					s := newline[start:last]
					switch s {
					case "args":
						argsBlock := &Block{
							kind: Token{
								str: s,
								typ: _word_t,
							},
							state:     _l2_kind_done,
							parent:    block,
							blockBody: &FMap{},
						}
						block.child = append(block.child, argsBlock)
						block = argsBlock
						ast.transfer(_l1_args_started)
					default:
						return errors.New("invalid statement in fn block: " + newline)
					}
				}
			} else {
				// the right bracket of fn block body is appeared, so fn block should be closed
				if current == '\n' && newline == "}" {
					ast.transfer(_l1_global)
					block.state = _l2_unknow
					break
				}
				if unicode.IsSpace(current) || current == '}' {
					break
				}
				start = last
				block.state = _l2_word_started
			}
		case _l1_args_started:
			switch block.state {
			case _l2_kind_done:
				if unicode.IsSpace(current) {
					break
				}
				if current == '=' {
					block.state = _l2_operator_started
				} else {
					return errors.New("invliad args block: " + newline)
				}
			case _l2_operator_started:
				if current == '{' || unicode.IsSpace(current) {
					block.operator = Token{
						str: "=",
						typ: _operator_t,
					}
					block.state = _l2_operator_done
					if current == '{' {
						ast.transfer(_l1_args_body_started)
					}
				} else {
					return errors.New("invalid args block: " + newline)
				}
			case _l2_operator_done:
				if unicode.IsSpace(current) {
					break
				}
				if current == '{' {
					ast.transfer(_l1_args_body_started)
				} else {
					return errors.New("invalid args block: " + newline)
				}
			}
		case _l1_args_body_started:
			if current == '\n' {
				ast.transfer(_l1_args_body_inside)
				break
			}
			if !unicode.IsSpace(current) {
				return errors.New("invalid args block: " + newline)
			}
		case _l1_args_body_inside:
			if current == '\n' {
				if newline == "}" {
					block = block.parent
					block.state = _l2_unknow
					ast.transfer(_l1_fn_body_inside)
				} else {
					if err := block.blockBody.Append(newline); err != nil {
						return err
					}
				}
			}
		default:
		}
		return nil
	}

	line = strings.TrimSpace(line)
	for i, c := range line {
		if err := finiteAutomata(i, c, line); err != nil {
			return err
		}
	}
	// todo, comment
	if err := finiteAutomata(len(line), '\n', line); err != nil {
		return err
	}
	ast.parsing = block
	return nil
}

// AST store all blocks in the flowl
//
type AST struct {
	global Block

	// for parsing
	_FA
}

func newAST() *AST {
	return &AST{
		global: Block{
			child: make([]*Block, 0),
		},
		_FA: _FA{
			parsing: nil,
			state:   _l1_global,
		},
	}
}

func deepwalk(b *Block, do func(*Block) error) error {
	// skip the global block
	if b.parent != nil {
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

type stateL1 int
type stateL2 int

type _FA struct {
	parsing  *Block
	state    stateL1
	prestate stateL1
}

func (f *_FA) transfer(s stateL1) {
	f.prestate = f.state
	f.state = s
}

func (f *_FA) phase() stateL1 {
	return f.state
}
