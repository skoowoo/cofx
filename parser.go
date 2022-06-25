//go:generate stringer -type parserStateL1
//go:generate stringer -type parserStateL2
package cofunc

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"unicode"
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
		if err := b.kind.Validate(); err != nil {
			return err
		}
		if err := b.target.Validate(); err != nil {
			return err
		}
		if err := b.operator.Validate(); err != nil {
			return err
		}
		if err := b.typevalue.Validate(); err != nil {
			return err
		}

		if b.blockBody != nil {
			lines := b.blockBody.GetStatements()
			for _, ln := range lines {
				for _, t := range ln.tokens {
					if err := t.Validate(); err != nil {
						return err
					}
				}
			}
		}
		return nil
	})
}

type stateL1 int
type stateL2 int

const (
	_l1_global stateL1 = iota
	_l1_block_started
	_l1_block_end
	_l1_load_block_started
	_l1_run_block_started
	_l1_run_body_started
	_l1_run_body_inside
	_l1_fn_block_started
	_l1_fn_body_started
	_l1_fn_body_inside
	_l1_args_started
	_l1_args_body_started
	_l1_args_body_inside
)

const (
	_l2_unknow stateL2 = iota
	_l2_multilines_started
	_l2_word_stared
	_l2_kind_started
	_l2_kind_done
	_l2_target_started
	_l2_target_done
	_l2_operator_started
	_l2_operator_done
	_l2_typevalue_started
	_l2_typevalue_done
)

func scanToken(ast *AST, line string, linenum int) error {
	prestate := ast.prestate
	state := ast.state
	block := ast.parsing

	var start int

	finiteAutomata := func(last int, chr rune, newline string) error {
		switch state {
		case _l1_global:
			if unicode.IsSpace(chr) {
				break
			}
			prestate = state
			state = _l1_block_started
			start = last
		case _l1_block_started:
			if !unicode.IsSpace(chr) && chr != '{' {
				break
			}
			var body blockBody = nil
			word := newline[start:last]
			switch word {
			case "load":
				prestate = state
				state = _l1_load_block_started
			case "fn":
				prestate = state
				state = _l1_fn_block_started
			case "run":
				if chr == '{' {
					prestate = state
					state = _l1_run_body_started
					body = &FList{}
				} else {
					prestate = state
					state = _l1_run_block_started
				}
			default:
				return errors.New("invalid block define: " + word)
			}
			block = &Block{
				kind: Token{
					value: word,
				},
				state:     _l2_kind_done,
				level:     _level_parent,
				blockBody: body,
			}
			ast.global.child = append(ast.global.child, block)
			ast.parsing = block
		case _l1_load_block_started:
			///
			// load go:sleep
			//
			if chr == '\n' {
				block.target = Token{
					value: strings.TrimSpace(newline[start:last]),
				}
				block.state = _l2_typevalue_done
				prestate = state
				state = _l1_global
				break
			}
			if unicode.IsSpace(chr) {
				break
			}
			if block.state != _l2_typevalue_started {
				block.state = _l2_typevalue_started
				start = last
			}
		case _l1_run_block_started:
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
			if chr == '\n' {
				// run function
				block.target = Token{
					value: strings.TrimSpace(newline[start:last]),
					typ:   _functionname_t,
				}
				block.state = _l2_typevalue_done
				prestate = state
				state = _l1_global
				break
			}

			if unicode.IsSpace(chr) {
				break
			}

			if chr == '{' {
				if block.state == _l2_typevalue_started {
					/*
						run sleep {
							time: 1s
						}
					*/
					block.target = Token{
						value: strings.TrimSpace(newline[start:last]),
						typ:   _functionname_t,
					}
					block.blockBody = &FMap{}
				} else {
					/*
						run {
							f1
							f2
						}
					*/
					block.blockBody = &FList{etype: _functionname_t}
				}

				block.state = _l2_typevalue_done
				prestate = state
				state = _l1_run_body_started
				break
			}

			if block.state != _l2_typevalue_started {
				block.state = _l2_typevalue_started
				start = last
			}
		case _l1_run_body_started:
			if chr == '\n' {
				prestate = state
				state = _l1_run_body_inside
				break
			}
			if !unicode.IsSpace(chr) {
				return errors.New("invalid run block: " + newline)
			}
		case _l1_run_body_inside:
			// 1. k: v
			// 2. f
			// 3. }
			if chr == '\n' {
				if newline == "}" {
					prestate = state
					state = _l1_global
				} else if newline != "" {
					if err := block.blockBody.Append(newline); err != nil {
						return err
					}
				}
			}
		case _l1_fn_block_started:
			/*
				fn f = f {

				}
			*/
			switch block.state {
			case _l2_kind_done:
				if unicode.IsSpace(chr) {
					break
				}
				block.state = _l2_target_started
				start = last
			case _l2_target_started:
				if unicode.IsSpace(chr) {
					block.state = _l2_target_done
				} else if chr == '=' {
					block.state = _l2_operator_started
				}
				s := newline[start:last]
				block.target = Token{
					value: s,
				}
			case _l2_target_done:
				if unicode.IsSpace(chr) {
					break
				}
				if chr == '=' {
					block.state = _l2_operator_started
				} else {
					return errors.New("invalid fn block: " + newline)
				}
			case _l2_operator_started:
				if unicode.IsSpace(chr) {
					block.state = _l2_operator_done
				} else {
					block.state = _l2_typevalue_started
					start = last
				}
				block.operator = Token{
					value: "=",
				}
			case _l2_operator_done:
				if unicode.IsSpace(chr) {
					break
				}
				block.state = _l2_typevalue_started
				start = last
			case _l2_typevalue_started:
				if unicode.IsSpace(chr) || chr == '{' {
					block.state = _l2_typevalue_done
					s := newline[start:last]
					block.typevalue = Token{
						value: s,
					}
					if chr == '{' {
						prestate = state
						state = _l1_fn_body_started
					}
				}
			case _l2_typevalue_done:
				if unicode.IsSpace(chr) {
					break
				}
				if chr == '{' {
					prestate = state
					state = _l1_fn_body_started
				} else {
					return errors.New("invalid fn block: " + newline)
				}
			}
		case _l1_fn_body_started:
			if chr == '\n' {
				prestate = state
				state = _l1_fn_body_inside
				break
			}
			if !unicode.IsSpace(chr) {
				return errors.New("invalid fn block: " + newline)
			}
		case _l1_fn_body_inside:
			if block.state == _l2_word_stared {
				if unicode.IsSpace(chr) || chr == '=' {
					block.state = _l2_unknow
					s := newline[start:last]
					switch s {
					case "args":
						argsBlock := &Block{
							kind: Token{
								value: s,
							},
							state:     _l2_kind_done,
							level:     _level_child,
							parent:    block,
							blockBody: &FMap{},
						}
						block.child = append(block.child, argsBlock)
						block = argsBlock
						prestate = state
						state = _l1_args_started
					default:
						return errors.New("invalid statement in fn block: " + newline)
					}
				}
			} else {
				// the right bracket of fn block body is appeared, so fn block should be closed
				if chr == '\n' && newline == "}" {
					prestate = state
					state = _l1_global
					block.state = _l2_unknow
					break
				}
				if unicode.IsSpace(chr) || chr == '}' {
					break
				}
				start = last
				block.state = _l2_word_stared
			}
		case _l1_args_started:
			switch block.state {
			case _l2_kind_done:
				if unicode.IsSpace(chr) {
					break
				}
				if chr == '=' {
					block.state = _l2_operator_started
				} else {
					return errors.New("invliad args block: " + newline)
				}
			case _l2_operator_started:
				if chr == '{' || unicode.IsSpace(chr) {
					block.operator = Token{
						value: "=",
					}
					block.state = _l2_operator_done
					if chr == '{' {
						prestate = state
						state = _l1_args_body_started
					}
				} else {
					return errors.New("invalid args block: " + newline)
				}
			case _l2_operator_done:
				if unicode.IsSpace(chr) {
					break
				}
				if chr == '{' {
					prestate = state
					state = _l1_args_body_started
				} else {
					return errors.New("invalid args block: " + newline)
				}
			}
		case _l1_args_body_started:
			if chr == '\n' {
				prestate = state
				state = _l1_args_body_inside
				break
			}
			if !unicode.IsSpace(chr) {
				return errors.New("invalid args block: " + newline)
			}
		case _l1_args_body_inside:
			if chr == '\n' {
				if newline == "}" {
					prestate = state
					state = _l1_fn_body_inside
					block = block.parent
					block.state = _l2_unknow
				} else {
					if err := block.blockBody.Append(newline); err != nil {
						return err
					}
				}
			}
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
	ast.prestate = prestate
	ast.state = state
	ast.parsing = block
	return nil
}
