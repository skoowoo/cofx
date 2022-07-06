//go:generate stringer -type stateL1
//go:generate stringer -type stateL2
package cofunc

import (
	"bufio"
	"errors"
	"io"
	"math"
	"strings"
	"unicode"

	"github.com/cofunclabs/cofunc/pkg/is"
)

func ParseAST(rd io.Reader) (*AST, error) {
	lx := newLexer()
	buff := bufio.NewReader(rd)
	for n := 1; ; n += 1 {
		line, err := buff.ReadString('\n')
		if err == io.EOF {
			if len(line) != 0 {
				if err := lx.split(line, n); err != nil {
					return nil, err
				}
			}
			break
		}
		if err != nil {
			return nil, err
		}
		if err := lx.split(line, n); err != nil {
			return nil, err
		}
	}

	ast := newAST()
	if err := ast.scanToken(lx); err != nil {
		return nil, err
	}

	return ast, ast.Foreach(func(b *Block) error {
		if err := doBlockHeader(b); err != nil {
			return err
		}

		if err := doBlockBody(b); err != nil {
			return err
		}
		return nil
	})
}

func doBlockBody(b *Block) error {
	if b.bbody == nil {
		return nil
	}
	lines := b.bbody.List()
	for _, l := range lines {
		// handle tokens
		for _, t := range l.tokens {
			t.setblock(b)
			if err := t.extractVar(); err != nil {
				return err
			}
			if err := t.validate(); err != nil {
				return err
			}
		}

		if err := buildVarGraph(b, l); err != nil {
			return err
		}
	}
	return nil
}

func buildVarGraph(b *Block, stm *Statement) error {
	if stm.desc != "var" {
		return nil
	}
	name := stm.tokens[0].String()
	v := &_var{
		segments: []struct {
			str   string
			isvar bool
		}{},
		child: []*_var{},
	}
	if len(stm.tokens) == 2 {
		vt := stm.tokens[1]
		if !vt.HasVar() {
			v.v = vt.String()
			v.cached = true
		} else {
			v.segments = vt.Segments()
			for _, seg := range v.segments {
				if !seg.isvar {
					continue
				}
				vname := seg.str
				chld, ok := b.GetVar(vname)
				if ok {
					v.child = append(v.child, chld)
				}
			}
		}
	}
	if err := b.PutVar(name, v); err != nil {
		return err
	}
	return nil
}

func doBlockHeader(b *Block) error {
	ts := []*Token{
		&b.kind,
		&b.target,
		&b.operator,
		&b.typevalue,
	}
	for _, t := range ts {
		t.setblock(b)
		if err := t.extractVar(); err != nil {
			return err
		}
		if err := t.validate(); err != nil {
			return err
		}
	}
	return nil
}

const (
	_l1_global stateL1 = iota
	_l1_keyword
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
	_l1_var_started
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

var statementPatterns = map[string]struct {
	min     int
	max     int
	types   []TokenType
	values  []string
	uptypes []TokenType
	newbody func() bbody
}{
	"load": {
		2, 2,
		[]TokenType{_word_t, _string_t},
		[]string{"", ""},
		[]TokenType{_keyword_t, _load_t},
		nil,
	},
	"fn": {
		5, 5,
		[]TokenType{_word_t, _word_t, _operator_t, _word_t, _symbol_t},
		[]string{"", "", "=", "", "{"},
		[]TokenType{_keyword_t, _functionname_t, _operator_t, _functionname_t, _symbol_t},
		func() bbody { return &plainbody{} },
	},
	"run1": {
		2, 2,
		[]TokenType{_word_t, _word_t},
		[]string{"", ""},
		[]TokenType{_keyword_t, _functionname_t},
		nil,
	},
	"run1+": {
		3, 3,
		[]TokenType{_word_t, _word_t, _symbol_t},
		[]string{"", "", "{"},
		[]TokenType{_keyword_t, _functionname_t, _symbol_t},
		func() bbody { return &FMap{} },
	},
	"run2": {
		2, 2,
		[]TokenType{_word_t, _symbol_t},
		[]string{"", "{"},
		[]TokenType{_keyword_t, _symbol_t},
		func() bbody { return &FList{etype: _functionname_t} },
	},
	"var": {
		2, math.MaxInt,
		[]TokenType{_word_t, _word_t, _operator_t},
		[]string{"", "", "="},
		[]TokenType{_keyword_t, _varname_t, _operator_t},
		nil,
	},
	"args": {
		3, 3,
		[]TokenType{_word_t, _operator_t, _symbol_t},
		[]string{"", "=", "{"},
		[]TokenType{_keyword_t, _operator_t, _symbol_t},
		func() bbody { return &FMap{} },
	},
	"kv": {
		3, 3,
		[]TokenType{_string_t, _symbol_t, _string_t},
		[]string{"", ":", ""},
		[]TokenType{_mapkey_t, _symbol_t, _string_t},
		nil,
	},
	"element": {
		1, 1,
		[]TokenType{_string_t},
		[]string{""},
		[]TokenType{_string_t},
		nil,
	},
	"closed": {
		1, 1,
		[]TokenType{_symbol_t},
		[]string{"}"},
		[]TokenType{_symbol_t},
		nil,
	},
}

func preparse(k string, tokens []*Token, b *Block) (bbody, error) {
	pattern := statementPatterns[k]

	if l := len(tokens); l < pattern.min || l > pattern.max {
		return nil, errors.New("invalid statement: " + k)
	}

	min := len(tokens)
	if l := len(pattern.types); min > l {
		min = l
	}

	for i := 0; i < min; i++ {
		t := tokens[i]
		p := pattern.types[i]
		v := pattern.values[i]

		if p != t.typ {
			return nil, errors.New("invalid statement: " + k)
		}
		if v != "" && v != t.String() {
			return nil, errors.New("invalid statement: " + k)
		}

		up := pattern.uptypes[i]
		t.typ = up
	}

	for _, t := range tokens {
		t._b = b
	}

	var body bbody
	if pattern.newbody != nil {
		body = pattern.newbody()
	}

	return body, nil
}

func (ast *AST) scanToken(lx *lexer) error {
	block := &ast.global

	lx.foreachLine(func(num int, line []*Token) error {
		if len(line) == 0 {
			return nil
		}
		switch ast.phase() {
		case _l1_global:
			kind := line[0]
			switch kind.String() {
			case "load":
				newb := &Block{
					child:    []*Block{},
					parent:   block,
					variable: vsys{vars: make(map[string]*_var)},
				}
				body, err := preparse("load", line, newb)
				if err != nil {
					return err
				}
				newb.bbody = body
				newb.kind = *kind
				newb.target = *line[1]

				block.child = append(block.child, newb)
			case "fn":
				newb := &Block{
					child:    []*Block{},
					parent:   block,
					variable: vsys{vars: make(map[string]*_var)},
					bbody:    &plainbody{},
				}
				body, err := preparse("fn", line, newb)
				if err != nil {
					return err
				}
				newb.bbody = body

				target, op, tv := line[1], line[2], line[3]

				newb.kind = *kind
				newb.target = *target
				newb.operator = *op
				newb.typevalue = *tv

				block.child = append(block.child, newb)

				block = newb
				ast.transfer(_l1_fn_body_started)
			case "run":
				newb := &Block{
					child:    []*Block{},
					parent:   block,
					variable: vsys{vars: make(map[string]*_var)},
					bbody:    nil,
				}

				var (
					body bbody
					err  error
				)
				keys := []string{"run1", "run1+", "run2"}
				for _, k := range keys {
					body, err = preparse(k, line, newb)
					if err == nil {
						newb.kind = *kind
						if k == "run1" || k == "run1+" {
							newb.target = *line[1]
						}
						newb.bbody = body
						break
					}
				}
				if err != nil {
					return err
				}

				block.child = append(block.child, newb)
				if newb.bbody != nil {
					block = newb
					ast.transfer(_l1_run_body_started)
				}
			case "var":
				if _, err := preparse("var", line, block); err != nil {
					return err
				}
				name := line[1]
				val := &Token{
					str: "",
					typ: _string_t,
					_b:  block,
				}
				if len(line) == 4 {
					val.str = line[3].String()
				}

				block.bbody.Append(newstm("var"))
				body := block.bbody.(*plainbody)
				body.Laststm().Append(name).Append(val)
			default:
				return errors.New("invalid block define: " + kind.String())
			}
		case _l1_fn_body_started:

		case _l1_run_body_started:
			if _, err := preparse("closed", line, block); err == nil {
				block = block.parent
				ast.transfer(_l1_global)
				break
			}

		}

		ast.parsing = block
		return nil
	})

	finiteAutomata := func(last int, current rune, newline string, rd *bufio.Reader) error {
		switch ast.phase() {
		case _l1_global:

		case _l1_run_body_inside:
			// 1. k: v
			// 2. f
			// 3. }
			if is.EOL(current) {
				if newline == "}" {
					ast.transfer(_l1_global)
					block = block.parent
				} else if newline != "" {
					if err := block.bbody.Append(newline); err != nil {
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
				if is.LB(current) {
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
						newb := &Block{
							kind:      Token{str: s, typ: _word_t},
							target:    Token{},
							operator:  Token{},
							typevalue: Token{},
							state:     _l2_kind_done,
							child:     []*Block{},
							parent:    block,
							variable:  vsys{},
							bbody:     &FMap{},
						}
						block.child = append(block.child, newb)
						block = newb
						ast.transfer(_l1_args_started)
					default:
						return errors.New("invalid statement in fn block: " + newline)
					}
				}
			} else {
				// the right bracket of fn block body is appeared, so fn block should be closed
				if current == '\n' && newline == "}" {
					block.state = _l2_unknow
					ast.transfer(_l1_global)
					block = block.parent
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
					if err := block.bbody.Append(newline); err != nil {
						return err
					}
				}
			}
		case _l1_var_started:
			// var a = 1
			// var a = $(b)
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
				// 1. var a
				if is.EOL(current) {
					s := newline[start:last]
					stm := block.bbody.(*plainbody).Laststm()
					stm.Append(&Token{
						str: strings.TrimSpace(s),
						typ: _varname_t,
						_b:  block,
					})
					block.state = _l2_unknow
					ast.transfer(_l1_global)
					break
				}

				// 2. var a = 1
				if is.Eq(current) {
					s := newline[start:last]
					stm := block.bbody.(*plainbody).Laststm()
					stm.Append(&Token{
						str: strings.TrimSpace(s),
						typ: _varname_t,
						_b:  block,
					})
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
				// not space, transfer
				start = last
				block.state = _l2_typevalue_started
			case _l2_typevalue_started: // from '{word}'
				// keep

				// transfer
				if is.EOL(current) {
					s := newline[start:last]
					stm := block.bbody.(*plainbody).Laststm()
					stm.Append(&Token{
						str: strings.TrimSpace(s),
						typ: _text_t,
						_b:  block,
					})
					block.state = _l2_unknow
					ast.transfer(_l1_global)
					break
				}
				// error
			}

		default:
		}
		return nil
	}

	line = strings.TrimSpace(line)
	buff := bufio.NewReader(strings.NewReader(line + "\n"))
	for i := 0; ; i += 1 {
		r, _, err := buff.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		if err := finiteAutomata(i, r, line, buff); err != nil {
			return err
		}
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
	ast := &AST{
		global: Block{
			kind:      Token{},
			target:    Token{},
			operator:  Token{},
			typevalue: Token{},
			state:     _l2_unknow,
			child:     make([]*Block, 0),
			parent:    nil,
			variable:  vsys{vars: make(map[string]*_var)},
			bbody:     &plainbody{},
		},
		_FA: _FA{
			state: _l1_global,
		},
	}
	ast._FA.parsing = &ast.global
	return ast
}

func deepwalk(b *Block, do func(*Block) error) error {
	if err := do(b); err != nil {
		return err
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

func (f *_FA) Back() {
}

func (f *_FA) phase() stateL1 {
	return f.state
}
