//go:generate stringer -type parserStateL1
//go:generate stringer -type parserStateL2
package flowl

import (
	"bufio"
	"errors"
	"io"
	"strings"
	"unicode"
)

// Parse parse a 'flowl' file
func ParseFile(file string) error {
	return nil
}

func Parse(rd io.Reader) (runq *RunQueue, bl *BlockList, err error) {
	if bl, err = ParseBlocks(rd); err != nil {
		return
	}
	runq = NewRunQueue()
	if err = runq.Generate(bl); err != nil {
		return
	}
	return
}

func ParseBlocks(rd io.Reader) (*BlockList, error) {
	bl := NewBlockList()
	num := 0
	scanner := bufio.NewScanner(rd)
	for {
		if !scanner.Scan() {
			break
		}
		num += 1
		err := scanToken(bl, scanner.Text(), num)
		if err != nil {
			return nil, err
		}
	}

	return bl, bl.Foreach(validateBlocks)
}

func validateBlocks(b *Block) error {
	if err := b.Kind.Validate(); err != nil {
		return err
	}
	if err := b.Target.Validate(); err != nil {
		return err
	}
	if err := b.Operator.Validate(); err != nil {
		return err
	}
	if err := b.TypeOrValue.Validate(); err != nil {
		return err
	}

	if b.BlockBody != nil {
		lines := b.BlockBody.Statements()
		for _, ln := range lines {
			for _, t := range ln.Tokens {
				if err := t.Validate(); err != nil {
					return err
				}
			}
		}
	}

	if len(b.Child) == 0 {
		return nil
	}
	for _, c := range b.Child {
		if err := validateBlocks(c); err != nil {
			return err
		}
	}
	return nil
}

type parserStateL1 int
type parserStateL2 int

const (
	_statel1_global parserStateL1 = iota
	_statel1_block_started
	_statel1_block_end
	_statel1_load_block_started
	_statel1_run_block_started
	_statel1_run_body_started
	_statel1_run_body_inside
	_statel1_fn_block_started
	_statel1_fn_body_started
	_statel1_fn_body_inside
	_statel1_args_started
	_statel1_args_body_started
	_statel1_args_body_inside
)

const (
	_statel2_unknow parserStateL2 = iota
	_statel2_multilines_started
	_statel2_word_stared
	_statel2_kind_started
	_statel2_kind_done
	_statel2_target_started
	_statel2_target_done
	_statel2_operator_started
	_statel2_operator_done
	_statel2_typeorvalue_started
	_statel2_typeorvalue_done
)

func scanToken(bl *BlockList, line string, linenum int) error {
	prestate := bl.prestate
	state := bl.state
	block := bl.parsing

	var startPos int

	finiteAutomata := func(last int, chr rune, newline string) error {
		switch state {
		case _statel1_global:
			if unicode.IsSpace(chr) {
				break
			}
			prestate = state
			state = _statel1_block_started
			startPos = last
		case _statel1_block_started:
			if !unicode.IsSpace(chr) && chr != '{' {
				break
			}
			var body BlockBody = nil
			word := newline[startPos:last]
			switch word {
			case "load":
				prestate = state
				state = _statel1_load_block_started
			case "fn":
				prestate = state
				state = _statel1_fn_block_started
			case "run":
				if chr == '{' {
					prestate = state
					state = _statel1_run_body_started
					body = &FlList{}
				} else {
					prestate = state
					state = _statel1_run_block_started
				}
			default:
				return errors.New("invalid block define: " + word)
			}
			block = &Block{
				Kind: Token{
					Value: word,
				},
				state:     _statel2_kind_done,
				Level:     LevelParent,
				BlockBody: body,
			}
			bl.l.PushBack(block)
			bl.parsing = block
		case _statel1_load_block_started:
			///
			// load go:sleep
			//
			if chr == '\n' {
				block.Target = Token{
					Value: strings.TrimSpace(newline[startPos:last]),
				}
				block.state = _statel2_typeorvalue_done
				prestate = state
				state = _statel1_global
				break
			}
			if unicode.IsSpace(chr) {
				break
			}
			if block.state != _statel2_typeorvalue_started {
				block.state = _statel2_typeorvalue_started
				startPos = last
			}
		case _statel1_run_block_started:
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
				block.Target = Token{
					Value: strings.TrimSpace(newline[startPos:last]),
					Type:  FunctionNameT,
				}
				block.state = _statel2_typeorvalue_done
				prestate = state
				state = _statel1_global
				break
			}

			if unicode.IsSpace(chr) {
				break
			}

			if chr == '{' {
				if block.state == _statel2_typeorvalue_started {
					/*
						run sleep {
							time: 1s
						}
					*/
					block.Target = Token{
						Value: strings.TrimSpace(newline[startPos:last]),
						Type:  FunctionNameT,
					}
					block.BlockBody = &FlMap{}
				} else {
					/*
						run {
							f1
							f2
						}
					*/
					block.BlockBody = &FlList{EType: FunctionNameT}
				}

				block.state = _statel2_typeorvalue_done
				prestate = state
				state = _statel1_run_body_started
				break
			}

			if block.state != _statel2_typeorvalue_started {
				block.state = _statel2_typeorvalue_started
				startPos = last
			}
		case _statel1_run_body_started:
			if chr == '\n' {
				prestate = state
				state = _statel1_run_body_inside
				break
			}
			if !unicode.IsSpace(chr) {
				return errors.New("invalid run block: " + newline)
			}
		case _statel1_run_body_inside:
			// 1. k: v
			// 2. f
			// 3. }
			if chr == '\n' {
				if newline == "}" {
					prestate = state
					state = _statel1_global
				} else if newline != "" {
					if err := block.BlockBody.Append(newline); err != nil {
						return err
					}
				}
			}
		case _statel1_fn_block_started:
			/*
				fn f = f {

				}
			*/
			switch block.state {
			case _statel2_kind_done:
				if unicode.IsSpace(chr) {
					break
				}
				block.state = _statel2_target_started
				startPos = last
			case _statel2_target_started:
				if unicode.IsSpace(chr) {
					block.state = _statel2_target_done
				} else if chr == '=' {
					block.state = _statel2_operator_started
				}
				s := newline[startPos:last]
				block.Target = Token{
					Value: s,
				}
			case _statel2_target_done:
				if unicode.IsSpace(chr) {
					break
				}
				if chr == '=' {
					block.state = _statel2_operator_started
				} else {
					return errors.New("invalid fn block: " + newline)
				}
			case _statel2_operator_started:
				if unicode.IsSpace(chr) {
					block.state = _statel2_operator_done
				} else {
					block.state = _statel2_typeorvalue_started
					startPos = last
				}
				block.Operator = Token{
					Value: "=",
				}
			case _statel2_operator_done:
				if unicode.IsSpace(chr) {
					break
				}
				block.state = _statel2_typeorvalue_started
				startPos = last
			case _statel2_typeorvalue_started:
				if unicode.IsSpace(chr) || chr == '{' {
					block.state = _statel2_typeorvalue_done
					s := newline[startPos:last]
					block.TypeOrValue = Token{
						Value: s,
					}
					if chr == '{' {
						prestate = state
						state = _statel1_fn_body_started
					}
				}
			case _statel2_typeorvalue_done:
				if unicode.IsSpace(chr) {
					break
				}
				if chr == '{' {
					prestate = state
					state = _statel1_fn_body_started
				} else {
					return errors.New("invalid fn block: " + newline)
				}
			}
		case _statel1_fn_body_started:
			if chr == '\n' {
				prestate = state
				state = _statel1_fn_body_inside
				break
			}
			if !unicode.IsSpace(chr) {
				return errors.New("invalid fn block: " + newline)
			}
		case _statel1_fn_body_inside:
			if block.state == _statel2_word_stared {
				if unicode.IsSpace(chr) || chr == '=' {
					block.state = _statel2_unknow
					s := newline[startPos:last]
					switch s {
					case "args":
						argsBlock := &Block{
							Kind: Token{
								Value: s,
							},
							state:     _statel2_kind_done,
							Level:     LevelChild,
							Parent:    block,
							BlockBody: &FlMap{},
						}
						block.Child = append(block.Child, argsBlock)
						block = argsBlock
						prestate = state
						state = _statel1_args_started
					default:
						return errors.New("invalid statement in fn block: " + newline)
					}
				}
			} else {
				// the right bracket of fn block body is appeared, so fn block should be closed
				if chr == '\n' && newline == "}" {
					prestate = state
					state = _statel1_global
					block.state = _statel2_unknow
					break
				}
				if unicode.IsSpace(chr) || chr == '}' {
					break
				}
				startPos = last
				block.state = _statel2_word_stared
			}
		case _statel1_args_started:
			switch block.state {
			case _statel2_kind_done:
				if unicode.IsSpace(chr) {
					break
				}
				if chr == '=' {
					block.state = _statel2_operator_started
				} else {
					return errors.New("invliad args block: " + newline)
				}
			case _statel2_operator_started:
				if chr == '{' || unicode.IsSpace(chr) {
					block.Operator = Token{
						Value: "=",
					}
					block.state = _statel2_operator_done
					if chr == '{' {
						prestate = state
						state = _statel1_args_body_started
					}
				} else {
					return errors.New("invalid args block: " + newline)
				}
			case _statel2_operator_done:
				if unicode.IsSpace(chr) {
					break
				}
				if chr == '{' {
					prestate = state
					state = _statel1_args_body_started
				} else {
					return errors.New("invalid args block: " + newline)
				}
			}
		case _statel1_args_body_started:
			if chr == '\n' {
				prestate = state
				state = _statel1_args_body_inside
				break
			}
			if !unicode.IsSpace(chr) {
				return errors.New("invalid args block: " + newline)
			}
		case _statel1_args_body_inside:
			if chr == '\n' {
				if newline == "}" {
					prestate = state
					state = _statel1_fn_body_inside
					block = block.Parent
					block.state = _statel2_unknow
				} else {
					if err := block.BlockBody.Append(newline); err != nil {
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
	bl.prestate = prestate
	bl.state = state
	bl.parsing = block
	return nil
}
