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
			typ:   IntT,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}
	{
		tk := &Token{
			value: "0100",
			typ:   IntT,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			value: "100.0",
			typ:   IntT,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}

	// load
	{
		tk := &Token{
			value: "go:print",
			typ:   LoadT,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}
	{
		tk := &Token{
			value: "go1:print",
			typ:   LoadT,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}
	{
		tk := &Token{
			value: "go:/path/print:1.0",
			typ:   LoadT,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}

	{
		tk := &Token{
			value: "go:print/",
			typ:   LoadT,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			value: "go-:print",
			typ:   LoadT,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			value: "1go:print",
			typ:   LoadT,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}

	//mapkey
	{
		tk := &Token{
			value: "abcABC123-",
			typ:   MapKeyT,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}
	{
		tk := &Token{
			value: "===",
			typ:   MapKeyT,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}

	{
		tk := &Token{
			value: "abc:1",
			typ:   MapKeyT,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			value: "abc:",
			typ:   MapKeyT,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}

	// functionname
	{
		tk := &Token{
			value: "printPrint123-a_",
			typ:   FunctionNameT,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}

	{
		tk := &Token{
			value: "123print",
			typ:   FunctionNameT,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			value: "print.",
			typ:   FunctionNameT,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			value: "print/",
			typ:   FunctionNameT,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
}
