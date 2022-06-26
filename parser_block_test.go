package cofunc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseVar(t *testing.T) {
	{
		text := `hello word\n`
		tk := Token{
			value: text,
			typ:   _text_t,
		}
		err := tk.extractVar()
		assert.NoError(t, err)
		assert.Len(t, tk.vars, 0)
	}
	{
		vs := "$(co)"
		name := "co"
		text := vs + `hello word\n`
		tk := Token{
			value: text,
			typ:   _text_t,
		}
		err := tk.extractVar()
		assert.NoError(t, err)
		assert.Len(t, tk.vars, 1)

		v := tk.vars[0]
		assert.Equal(t, name, v.n)
		assert.Equal(t, 0, v.s)
		assert.Equal(t, v.s+len(vs), v.e)
	}
	{
		vs := "$(co)"
		name := "co"
		text := `123456789\n` + vs
		tk := Token{
			value: text,
			typ:   _text_t,
		}
		err := tk.extractVar()
		assert.NoError(t, err)
		assert.Len(t, tk.vars, 1)

		v := tk.vars[0]
		assert.Equal(t, name, v.n)
		assert.Equal(t, len(text)-len(vs), v.s)
		assert.Equal(t, len(text), v.e)
	}
	{
		vs := "$(co)"
		name := "co"
		text := "123456" + vs + "word\n"
		tk := Token{
			value: text,
			typ:   _text_t,
		}
		err := tk.extractVar()
		assert.NoError(t, err)
		assert.Len(t, tk.vars, 1)

		v := tk.vars[0]
		assert.Equal(t, name, v.n)
		assert.Equal(t, 6, v.s)
		assert.Equal(t, 6+len(vs), v.e)
	}
	{
		vs1 := "$(co1)"
		vs2 := "$(co2)"
		text := "123456" + vs1 + vs2 + "word\n"
		tk := Token{
			value: text,
			typ:   _text_t,
		}
		err := tk.extractVar()
		assert.NoError(t, err)
		assert.Len(t, tk.vars, 2)

		v1 := tk.vars[0]
		v2 := tk.vars[1]
		assert.Equal(t, "co1", v1.n)
		assert.Equal(t, "co2", v2.n)
	}
	{
		vs1 := "$(co1)"
		fake := "\\$(co2)"
		text := "123456" + vs1 + fake + "word\n"
		tk := Token{
			value: text,
			typ:   _text_t,
		}
		err := tk.extractVar()
		assert.NoError(t, err)
		assert.Len(t, tk.vars, 1)

		v1 := tk.vars[0]
		assert.Equal(t, "co1", v1.n)
	}
	{
		vs1 := "$(co1"
		text := "123456" + vs1 + "word\n"
		tk := Token{
			value: text,
			typ:   _text_t,
		}
		err := tk.extractVar()
		assert.NoError(t, err)
	}
	{
		vs1 := "$(co1"
		text := "123456" + vs1 + "word"
		tk := Token{
			value: text,
			typ:   _text_t,
		}
		err := tk.extractVar()
		assert.NoError(t, err)
	}
}

func TestValidateToken(t *testing.T) {
	// int
	{
		tk := &Token{
			value: "100",
			typ:   _int_t,
		}
		err := tk.validate()
		assert.NoError(t, err)
	}
	{
		tk := &Token{
			value: "0100",
			typ:   _int_t,
		}
		err := tk.validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			value: "100.0",
			typ:   _int_t,
		}
		err := tk.validate()
		assert.Error(t, err)
	}

	// load
	{
		tk := &Token{
			value: "go:print",
			typ:   _load_t,
		}
		err := tk.validate()
		assert.NoError(t, err)
	}
	{
		tk := &Token{
			value: "go1:print",
			typ:   _load_t,
		}
		err := tk.validate()
		assert.NoError(t, err)
	}
	{
		tk := &Token{
			value: "go:/path/print:1.0",
			typ:   _load_t,
		}
		err := tk.validate()
		assert.NoError(t, err)
	}

	{
		tk := &Token{
			value: "go:print/",
			typ:   _load_t,
		}
		err := tk.validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			value: "go-:print",
			typ:   _load_t,
		}
		err := tk.validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			value: "1go:print",
			typ:   _load_t,
		}
		err := tk.validate()
		assert.Error(t, err)
	}

	//mapkey
	{
		tk := &Token{
			value: "abcABC123-",
			typ:   _mapkey_t,
		}
		err := tk.validate()
		assert.NoError(t, err)
	}
	{
		tk := &Token{
			value: "===",
			typ:   _mapkey_t,
		}
		err := tk.validate()
		assert.NoError(t, err)
	}

	{
		tk := &Token{
			value: "abc:1",
			typ:   _mapkey_t,
		}
		err := tk.validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			value: "abc:",
			typ:   _mapkey_t,
		}
		err := tk.validate()
		assert.Error(t, err)
	}

	// functionname
	{
		tk := &Token{
			value: "printPrint123-a_",
			typ:   _functionname_t,
		}
		err := tk.validate()
		assert.NoError(t, err)
	}

	{
		tk := &Token{
			value: "123print",
			typ:   _functionname_t,
		}
		err := tk.validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			value: "print.",
			typ:   _functionname_t,
		}
		err := tk.validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			value: "print/",
			typ:   _functionname_t,
		}
		err := tk.validate()
		assert.Error(t, err)
	}
}
