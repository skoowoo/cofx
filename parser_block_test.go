package cofunc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractAndCalcVar(t *testing.T) {
	get := func(b *Block, name string) (string, bool) {
		return name, true
	}
	{
		text := `hello word\n`
		tk := Token{
			str: text,
			typ: _text_t,
			get: get,
		}
		err := tk.extractVar()
		assert.NoError(t, err)
		assert.Len(t, tk.segments, 1)

		vl := tk.Value()
		assert.Equal(t, text, vl)
	}
	{
		vs := "$(co)"
		text := vs + `hello word\n`
		tk := Token{
			str: text,
			typ: _text_t,
			get: get,
		}
		err := tk.extractVar()
		assert.NoError(t, err)
		assert.Len(t, tk.segments, 2)

		vl := tk.Value()
		assert.Equal(t, `cohello word\n`, vl)
	}
	{
		vs := "$(co)"
		text := `123456789\n` + vs
		tk := Token{
			str: text,
			typ: _text_t,
			get: get,
		}
		err := tk.extractVar()
		assert.NoError(t, err)
		assert.Len(t, tk.segments, 2)

		vl := tk.Value()
		assert.Equal(t, `123456789\nco`, vl)
	}
	{
		vs := "$(co)"
		text := "123456" + vs + "word\n"
		tk := Token{
			str: text,
			typ: _text_t,
			get: get,
		}
		err := tk.extractVar()
		assert.NoError(t, err)
		assert.Len(t, tk.segments, 3)

		vl := tk.Value()
		assert.Equal(t, "123456coword\n", vl)
	}
	{
		vs1 := "$(co1)"
		vs2 := "$(co2)"
		text := "123456" + vs1 + vs2 + "word\n"
		tk := Token{
			str: text,
			typ: _text_t,
			get: get,
		}
		err := tk.extractVar()
		assert.NoError(t, err)
		assert.Len(t, tk.segments, 4)

		vl := tk.Value()
		assert.Equal(t, "123456co1co2word\n", vl)
	}
	{
		vs1 := "$(co1)"
		fake := `\$(co2)`
		text := "123456" + vs1 + fake + "word\n"
		tk := Token{
			str: text,
			typ: _text_t,
			get: get,
		}
		err := tk.extractVar()
		assert.NoError(t, err)
		assert.Len(t, tk.segments, 3)

		vl := tk.Value()
		assert.Equal(t, "123456co1$(co2)word\n", vl)
	}
	{
		vs1 := "$(co1"
		text := "123456" + vs1 + "word\n"
		tk := Token{
			str: text,
			typ: _text_t,
			get: get,
		}
		err := tk.extractVar()
		assert.NoError(t, err)

		vl := tk.Value()
		assert.Equal(t, "123456$(co1word\n", vl)
	}
	{
		vs1 := "$(co1"
		text := "123456" + vs1 + "word"
		tk := Token{
			str: text,
			typ: _text_t,
			get: get,
		}
		err := tk.extractVar()
		assert.NoError(t, err)

		vl := tk.Value()
		assert.Equal(t, "123456$(co1word", vl)
	}
}

func TestValidateToken(t *testing.T) {
	// int
	{
		tk := &Token{
			str: "100",
			typ: _int_t,
		}
		err := tk.validate()
		assert.NoError(t, err)
	}
	{
		tk := &Token{
			str: "0100",
			typ: _int_t,
		}
		err := tk.validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			str: "100.0",
			typ: _int_t,
		}
		err := tk.validate()
		assert.Error(t, err)
	}

	// load
	{
		tk := &Token{
			str: "go:print",
			typ: _load_t,
		}
		err := tk.validate()
		assert.NoError(t, err)
	}
	{
		tk := &Token{
			str: "go1:print",
			typ: _load_t,
		}
		err := tk.validate()
		assert.NoError(t, err)
	}
	{
		tk := &Token{
			str: "go:/path/print:1.0",
			typ: _load_t,
		}
		err := tk.validate()
		assert.NoError(t, err)
	}

	{
		tk := &Token{
			str: "go:print/",
			typ: _load_t,
		}
		err := tk.validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			str: "go-:print",
			typ: _load_t,
		}
		err := tk.validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			str: "1go:print",
			typ: _load_t,
		}
		err := tk.validate()
		assert.Error(t, err)
	}

	//mapkey
	{
		tk := &Token{
			str: "abcABC123-",
			typ: _mapkey_t,
		}
		err := tk.validate()
		assert.NoError(t, err)
	}
	{
		tk := &Token{
			str: "===",
			typ: _mapkey_t,
		}
		err := tk.validate()
		assert.NoError(t, err)
	}

	{
		tk := &Token{
			str: "abc:1",
			typ: _mapkey_t,
		}
		err := tk.validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			str: "abc:",
			typ: _mapkey_t,
		}
		err := tk.validate()
		assert.Error(t, err)
	}

	// functionname
	{
		tk := &Token{
			str: "printPrint123-a_",
			typ: _functionname_t,
		}
		err := tk.validate()
		assert.NoError(t, err)
	}

	{
		tk := &Token{
			str: "123print",
			typ: _functionname_t,
		}
		err := tk.validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			str: "print.",
			typ: _functionname_t,
		}
		err := tk.validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			str: "print/",
			typ: _functionname_t,
		}
		err := tk.validate()
		assert.Error(t, err)
	}
}
