package cofunc

import (
	"fmt"
)

type lexerError struct{}

func lexerErr() lexerError {
	return lexerError{}
}
func (e lexerError) New(line string, ln int, current rune, state lexstate) error {
	return fmt.Errorf("%d: %s, parsing '%c' in %s: illegal character", ln, line, current, state)
}

type parseTokenTypeError struct{}

func parseTokenTypeErr() parseTokenTypeError {
	return parseTokenTypeError{}
}

func (e parseTokenTypeError) New(line []*Token, ln int, t *Token, expect TokenType) error {
	return fmt.Errorf("%d: '%s', actual type '%s', expect type '%s': token type not match", ln, t.String(), t.typ, expect)
}

type parseTokenValError struct{}

func parseTokenValErr() parseTokenValError {
	return parseTokenValError{}
}

func (e parseTokenValError) New(line []*Token, ln int, t *Token, expect string) error {
	return fmt.Errorf("%d: '%s', actual value '%s', expect value '%s': token value not match", ln, t.String(), t.String(), expect)
}

type generatorError struct{}

func generatorErr() generatorError {
	return generatorError{}
}

func (e generatorError) New() error {
	return nil
}
