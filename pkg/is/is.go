package is

import "unicode"

func Space(x rune) bool {
	return x == ' ' || x == '\t'
}

func SpaceOrEOL(x rune) bool {
	return unicode.IsSpace(x)
}

func EOL(x rune) bool {
	return x == '\n'
}

func LB(x rune) bool {
	return x == '{'
}

func RB(x rune) bool {
	return x == '}'
}

func Colon(x rune) bool {
	return x == ':'
}

func Quotation(x rune) bool {
	return x == '"'
}

func Dollar(x rune) bool {
	return x == '$'
}

func BackSlash(x rune) bool {
	return x == '\\'
}

func EQ(x rune) bool {
	return x == '='
}

func Symbol(x rune) bool {
	return LB(x) || RB(x) || x == ':' || EQ(x) || x == '+' || x == '-' || x == '*' || x == '/'
}

func Word(x rune) bool {
	if x >= 'a' && x <= 'z' {
		return true
	}
	if x >= 'A' && x <= 'Z' {
		return true
	}
	if x >= '0' && x <= '9' {
		return true
	}
	if x == '_' || x == '-' {
		return true
	}
	return false
}
