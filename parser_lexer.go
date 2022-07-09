package cofunc

import (
	"strings"

	"github.com/cofunclabs/cofunc/pkg/is"
)

type ststate int

const (
	_lx_unknow ststate = iota
	_lx_identifier
	_lx_symbol
	_lx_string
	_lx_string_backslash
	_lx_var_directuse1
	_lx_var_directuse2
)

type lexer struct {
	tt        map[int][]*Token
	nums      []int
	state     ststate
	buf       strings.Builder
	stringNum int
}

func newLexer() *lexer {
	return &lexer{
		tt:    make(map[int][]*Token),
		state: _lx_unknow,
	}
}

func (l *lexer) saveRune(r rune) {
	l.buf.WriteRune(r)
}

func (l *lexer) exportString() string {
	s := l.buf.String()
	l.buf.Reset()
	return s
}

func (l *lexer) insert(num int, t *Token) {
	_, ok := l.tt[num]
	if !ok {
		l.tt[num] = make([]*Token, 0)
	}
	l.tt[num] = append(l.tt[num], t)
}

func (l *lexer) get(num int) []*Token {
	ts, ok := l.tt[num]
	if ok {
		return ts
	}
	return nil
}

func (l *lexer) _goto(s ststate) {
	l.state = s
}

func (l *lexer) split(line string, ln int) error {
	l.nums = append(l.nums, ln)

	for _, c := range line {
		switch l.state {
		case _lx_unknow:
			if is.Space(c) || is.EOL(c) {
				break
			}
			if is.Identifier(c) {
				l.saveRune(c)
				l._goto(_lx_identifier)
				break
			}
			if is.Symbol(c) {
				l.saveRune(c)
				l._goto(_lx_symbol)
				break
			}
			// string
			if is.Quotation(c) {
				l.stringNum = ln
				l._goto(_lx_string)
				break
			}
			// var direct use
			if is.Dollar(c) {
				l.saveRune(c)
				l._goto(_lx_var_directuse1)
				break
			}
			return lexerErr().New(line, ln, c, l.state)
		case _lx_symbol:
			if is.Symbol(c) {
				l.saveRune(c)
				break
			}
			l.insert(ln, &Token{
				str: l.exportString(),
				typ: _symbol_t,
			})
			// Here is special handling of comments, because one comment can contain unicode character
			if ts := l.get(ln); ts != nil {
				if len(ts) == 1 && ts[0].String() == "//" {
					// skip the remaining characters on the current line
					// We don't save them into a token, Maybe someday We will save them.
					l._goto(_lx_unknow)
					return nil
				}
			}

			if is.Space(c) || is.EOL(c) {
				l._goto(_lx_unknow)
				break
			}
			if is.Identifier(c) {
				l.saveRune(c)
				l._goto(_lx_identifier)
				break
			}
			if is.Quotation(c) {
				l.stringNum = ln
				l._goto(_lx_string)
				break
			}
			return lexerErr().New(line, ln, c, l.state)
		case _lx_identifier:
			if is.Identifier(c) {
				l.saveRune(c)
				break
			}
			l.insert(ln, &Token{
				str: l.exportString(),
				typ: _identifier_t,
			})

			if is.Space(c) || is.EOL(c) {
				l._goto(_lx_unknow)
				break
			}
			if is.Symbol(c) {
				l.saveRune(c)
				l._goto(_lx_symbol)
				break
			}
			return lexerErr().New(line, ln, c, l.state)
		case _lx_string:
			if is.BackSlash(c) {
				l._goto(_lx_string_backslash)
				break
			}
			if is.Quotation(c) {
				// use l.stringNum to replace num, aim to support multi line string
				l.insert(l.stringNum, &Token{
					str: l.exportString(),
					typ: _string_t,
				})
				l._goto(_lx_unknow)
				break
			}
			l.saveRune(c)
		case _lx_string_backslash:
			if !is.Quotation(c) {
				l.saveRune('\\')
			}
			l.saveRune(c)
			l._goto(_lx_string)
		case _lx_var_directuse1:
			if c == '(' {
				l.saveRune(c)
				l._goto(_lx_var_directuse2)
				break
			}
			return lexerErr().New(line, ln, c, l.state)
		case _lx_var_directuse2:
			if is.Identifier(c) {
				l.saveRune(c)
				break
			}
			if c == ')' {
				l.saveRune(c)
				l.insert(ln, &Token{
					str: l.exportString(),
					typ: _string_t,
				})
				l._goto(_lx_unknow)
				break
			}
			return lexerErr().New(line, ln, c, l.state)
		}
	}
	return nil
}

func (l *lexer) foreachLine(do func(num int, line []*Token) error) error {
	for n := range l.nums {
		line, ok := l.tt[n]
		if !ok {
			continue
		}
		if err := do(n, line); err != nil {
			return err
		}
	}
	return nil
}