//go:generate stringer -type stateL1
//go:generate stringer -type stateL2
package cofunc

import (
	"bufio"
	"errors"
	"io"
	"math"
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
	if err := ast.scan(lx); err != nil {
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
		if err := t.extractVar(); err != nil {
			return err
		}
		if err := t.validate(); err != nil {
			return err
		}
	}
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
	return ast
}

func (a *AST) Foreach(do func(*Block) error) error {
	return deepwalk(&a.global, do)
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

type stateL1 int
type stateL2 int

type _FA struct {
	state    stateL1
	prestate stateL1
}

func (f *_FA) _goto(s stateL1) {
	f.prestate = f.state
	f.state = s
}

func (f *_FA) phase() stateL1 {
	return f.state
}

const (
	_l1_global stateL1 = iota
	_l1_run_body
	_l1_fn_body
	_l1_args_body
)

const (
	_l2_unknow stateL2 = iota
	_l2_multilines
	_l2_word
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
		[]TokenType{_word_t, _word_t, _symbol_t, _word_t, _symbol_t},
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
		[]TokenType{_word_t, _word_t, _symbol_t},
		[]string{"", "", "="},
		[]TokenType{_keyword_t, _varname_t, _operator_t},
		nil,
	},
	"args": {
		3, 3,
		[]TokenType{_word_t, _symbol_t, _symbol_t},
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

func (ast *AST) preparse(k string, tokens []*Token, b *Block) (bbody, error) {
	pattern := statementPatterns[k]

	if l := len(tokens); l < pattern.min || l > pattern.max {
		return nil, errors.New("invalid statement(token number): " + k)
	}

	min := len(tokens)
	if l := len(pattern.types); min > l {
		min = l
	}

	for i := 0; i < min; i++ {
		t := tokens[i]
		pt := pattern.types[i]
		pv := pattern.values[i]

		if pt != t.typ {
			return nil, errors.New("invalid statement(token type): " + k)
		}
		if pv != "" && pv != t.String() {
			return nil, errors.New("invalid statement(token value): " + k)
		}
	}

	for i := 0; i < min; i++ {
		t := tokens[i]
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

func (ast *AST) scan(lx *lexer) error {
	parsingblock := &ast.global

	return lx.foreachLine(func(num int, line []*Token) error {
		if len(line) == 0 {
			return nil
		}
		switch ast.phase() {
		case _l1_global:
			kind := line[0]
			switch kind.String() {
			case "//":
				return nil
			case "load":
				nb := &Block{
					child:    []*Block{},
					parent:   parsingblock,
					variable: vsys{vars: make(map[string]*_var)},
				}
				body, err := ast.preparse("load", line, nb)
				if err != nil {
					return err
				}
				nb.bbody = body
				nb.kind = *kind
				nb.target = *line[1]

				parsingblock.child = append(parsingblock.child, nb)
			case "fn":
				nb := &Block{
					child:    []*Block{},
					parent:   parsingblock,
					variable: vsys{vars: make(map[string]*_var)},
					bbody:    &plainbody{},
				}
				body, err := ast.preparse("fn", line, nb)
				if err != nil {
					return err
				}
				nb.bbody = body

				target, op, tv := line[1], line[2], line[3]

				nb.kind = *kind
				nb.target = *target
				nb.operator = *op
				nb.typevalue = *tv

				parsingblock.child = append(parsingblock.child, nb)

				parsingblock = nb
				ast._goto(_l1_fn_body)
			case "run":
				nb := &Block{
					child:    []*Block{},
					parent:   parsingblock,
					variable: vsys{vars: make(map[string]*_var)},
					bbody:    nil,
				}

				var (
					body bbody
					err  error
				)
				keys := []string{"run1", "run1+", "run2"}
				for _, k := range keys {
					body, err = ast.preparse(k, line, nb)
					if err == nil {
						nb.kind = *kind
						if k == "run1" || k == "run1+" {
							nb.target = *line[1]
						}
						nb.bbody = body
						break
					}
				}
				if err != nil {
					return err
				}

				parsingblock.child = append(parsingblock.child, nb)
				if nb.bbody != nil {
					parsingblock = nb
					ast._goto(_l1_run_body)
				}
			case "var":
				if _, err := ast.preparse("var", line, parsingblock); err != nil {
					return err
				}
				name := line[1]
				var val *Token
				if len(line) == 4 {
					val = line[3]
				}

				parsingblock.bbody.Append(newstm("var"))
				body := parsingblock.bbody.(*plainbody)
				body.Laststm().Append(name)
				if val != nil {
					body.Laststm().Append(val)
				}
			default:
				return errors.New("invalid block define: " + kind.String())
			}
		case _l1_fn_body:
			if _, err := ast.preparse("closed", line, parsingblock); err == nil {
				parsingblock = parsingblock.parent
				ast._goto(_l1_global)
				break
			}

			kind := line[0]
			switch kind.String() {
			case "args":
				nb := &Block{
					child:    []*Block{},
					parent:   parsingblock,
					variable: vsys{vars: make(map[string]*_var)},
				}
				body, err := ast.preparse("args", line, nb)
				if err != nil {
					return err
				}
				nb.bbody = body
				nb.kind = *kind

				parsingblock.child = append(parsingblock.child, nb)
				parsingblock = nb
				ast._goto(_l1_args_body)
			default:
				return errors.New("invalid statement: " + kind.String())
			}

		case _l1_args_body:
			if _, err := ast.preparse("closed", line, parsingblock); err == nil {
				parsingblock = parsingblock.parent
				ast._goto(_l1_fn_body)
				break
			}
			for _, t := range line {
				t._b = parsingblock
			}
			if err := parsingblock.bbody.Append(line); err != nil {
				return err
			}
		case _l1_run_body:
			if _, err := ast.preparse("closed", line, parsingblock); err == nil {
				parsingblock = parsingblock.parent
				ast._goto(_l1_global)
				break
			}

			for _, t := range line {
				t._b = parsingblock
			}
			if err := parsingblock.bbody.Append(line); err != nil {
				return err
			}
		}

		return nil
	})
}
