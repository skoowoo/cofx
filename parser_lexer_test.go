package cofunc

import (
	"bufio"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func loadTestingdataForLexer(testingdata string) (*lexer, error) {
	lx := &lexer{
		tt:    make(map[int][]*Token),
		state: _lx_unknow,
	}

	buff := bufio.NewReader(strings.NewReader(testingdata))
	for i := 1; ; i += 1 {
		line, err := buff.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				err = lx.split(line, i)
				return lx, err
			}
			return nil, err
		}
		if err := lx.split(line, i); err != nil {
			return nil, err
		}
	}
}

func TestLexerLoad(t *testing.T) {
	testingdata := `
	load "cmd:root/function1"
	`
	_, err := loadTestingdataForLexer(testingdata)
	assert.NoError(t, err)
}

func TestLexerVar(t *testing.T) {
	testingdata := `
	var a
	`
	_, err := loadTestingdataForLexer(testingdata)
	assert.NoError(t, err)
}

func TestLexer(t *testing.T) {
	testingdata := `load "go:print"

co print {
	"STARTING": "build \"cofunc\""
}
fn gobuild = command {
	"hello
world"`

	lx, err := loadTestingdataForLexer(testingdata)
	assert.NoError(t, err)

	assert.Len(t, lx.nums, 8)
	assert.Len(t, lx.tt, 6)

	err = lx.foreachLine(func(n int, line []*Token) error {
		// load "go:print"
		if n == 1 {
			assert.Len(t, line, 2)
			assert.Equal(t, "load", line[0].String())
			assert.Equal(t, _identifier_t, line[0].typ)

			assert.Equal(t, "go:print", line[1].String())
			assert.Equal(t, _string_t, line[1].typ)
		}

		// co print {
		if n == 3 {
			assert.Len(t, line, 3)
			assert.Equal(t, "co", line[0].String())
			assert.Equal(t, _identifier_t, line[0].typ)

			assert.Equal(t, "print", line[1].String())
			assert.Equal(t, _identifier_t, line[1].typ)

			assert.Equal(t, "{", line[2].String())
			assert.Equal(t, _symbol_t, line[2].typ)
		}
		// "STARTING": "build \"cofunc\""
		if n == 4 {
			assert.Len(t, line, 3)
			assert.Equal(t, "STARTING", line[0].String())
			assert.Equal(t, _string_t, line[0].typ)

			assert.Equal(t, ":", line[1].String())
			assert.Equal(t, _symbol_t, line[1].typ)

			assert.Equal(t, `build "cofunc"`, line[2].String())
			assert.Equal(t, _string_t, line[2].typ)
		}
		// }
		if n == 5 {
			assert.Len(t, line, 1)
			assert.Equal(t, "}", line[0].String())
			assert.Equal(t, _symbol_t, line[0].typ)
		}
		// fn gobuild = command {
		if n == 6 {
			assert.Len(t, line, 5)
			assert.Equal(t, "fn", line[0].String())
			assert.Equal(t, _identifier_t, line[0].typ)

			assert.Equal(t, "gobuild", line[1].String())
			assert.Equal(t, _identifier_t, line[1].typ)

			assert.Equal(t, "=", line[2].String())
			assert.Equal(t, _symbol_t, line[2].typ)

			assert.Equal(t, "command", line[3].String())
			assert.Equal(t, _identifier_t, line[3].typ)

			assert.Equal(t, "{", line[4].String())
			assert.Equal(t, _symbol_t, line[4].typ)
		}

		if n == 8 {
			assert.Len(t, line, 1)
			assert.Equal(t, "hello\nworld", line[0].String())
			assert.Equal(t, _string_t, line[0].typ)
		}
		return nil
	})
	assert.NoError(t, err)
}
