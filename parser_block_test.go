package cofunc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateToken(t *testing.T) {
	// int
	{
		tk := &Token{
			value: "100",
			typ:   _int_t,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}
	{
		tk := &Token{
			value: "0100",
			typ:   _int_t,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			value: "100.0",
			typ:   _int_t,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}

	// load
	{
		tk := &Token{
			value: "go:print",
			typ:   _load_t,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}
	{
		tk := &Token{
			value: "go1:print",
			typ:   _load_t,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}
	{
		tk := &Token{
			value: "go:/path/print:1.0",
			typ:   _load_t,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}

	{
		tk := &Token{
			value: "go:print/",
			typ:   _load_t,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			value: "go-:print",
			typ:   _load_t,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			value: "1go:print",
			typ:   _load_t,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}

	//mapkey
	{
		tk := &Token{
			value: "abcABC123-",
			typ:   _mapkey_t,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}
	{
		tk := &Token{
			value: "===",
			typ:   _mapkey_t,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}

	{
		tk := &Token{
			value: "abc:1",
			typ:   _mapkey_t,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			value: "abc:",
			typ:   _mapkey_t,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}

	// functionname
	{
		tk := &Token{
			value: "printPrint123-a_",
			typ:   _functionname_t,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}

	{
		tk := &Token{
			value: "123print",
			typ:   _functionname_t,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			value: "print.",
			typ:   _functionname_t,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			value: "print/",
			typ:   _functionname_t,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
}
