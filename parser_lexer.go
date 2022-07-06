package cofunc

import (
	"errors"
	"strings"

	"github.com/cofunclabs/cofunc/pkg/is"
)

type ststate int

const (
	_lx_unknow ststate = iota
	_lx_word
	_lx_string
	_lx_string_backslash
	_lx_var_directuse1
	_lx_var_directuse2
)

type lexer struct {
	tt    map[int][]*Token
	nums  []int
	state ststate
	buf   strings.Builder
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

func (l *lexer) _goto(s ststate) {
	l.state = s
}

func (l *lexer) split(line string, num int) error {
	l.nums = append(l.nums, num)

	for _, c := range line {
		switch l.state {
		case _lx_unknow:
			if is.Space(c) || is.EOL(c) {
				break
			}
			if is.Word(c) {
				l.saveRune(c)
				l._goto(_lx_word)
				break
			}
			//symbol
			if is.LB(c) || is.RB(c) || is.Colon(c) || is.Eq(c) {
				l.saveRune(c)
				l.insert(num, &Token{
					str: l.exportString(),
					typ: _symbol_t,
				})
				l._goto(_lx_unknow)
				break
			}
			// string
			if is.Quotation(c) {
				l._goto(_lx_string)
				break
			}
			// var direct use
			if is.Dollar(c) {
				l.saveRune(c)
				l._goto(_lx_var_directuse1)
				break
			}
			return errors.New("contain invalid character: " + line)
		case _lx_word:
			if is.Word(c) {
				l.saveRune(c)
				break
			}
			l.insert(num, &Token{
				str: l.exportString(),
				typ: _word_t,
			})

			if is.Space(c) {
				l._goto(_lx_unknow)
				break
			}
			//symbol
			if is.LB(c) || is.RB(c) || is.Colon(c) || is.Eq(c) {
				l.saveRune(c)
				l.insert(num, &Token{
					str: l.exportString(),
					typ: _symbol_t,
				})
				l._goto(_lx_unknow)
				break
			}
			return errors.New("contain invalid character: " + line)
		case _lx_string:
			if is.BackSlash(c) {
				l._goto(_lx_string_backslash)
				break
			}
			if is.Quotation(c) {
				l.insert(num, &Token{
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
			return errors.New("contain invalid character: " + line)
		case _lx_var_directuse2:
			if is.Word(c) {
				l.saveRune(c)
				break
			}
			if c == ')' {
				l.saveRune(c)
				l.insert(num, &Token{
					str: l.exportString(),
					typ: _string_t,
				})
				l._goto(_lx_unknow)
				break
			}
			return errors.New("contain invalid character: " + line)
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
