package flowl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateToken(t *testing.T) {
	// int
	{
		tk := &Token{
			Value: "100",
			Type:  IntT,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}
	{
		tk := &Token{
			Value: "0100",
			Type:  IntT,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			Value: "100.0",
			Type:  IntT,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}

	// load
	{
		tk := &Token{
			Value: "go:print",
			Type:  LoadT,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}
	{
		tk := &Token{
			Value: "go1:print",
			Type:  LoadT,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}
	{
		tk := &Token{
			Value: "go:/path/print:1.0",
			Type:  LoadT,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}

	{
		tk := &Token{
			Value: "go:print/",
			Type:  LoadT,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			Value: "go-:print",
			Type:  LoadT,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			Value: "1go:print",
			Type:  LoadT,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}

	//mapkey
	{
		tk := &Token{
			Value: "abcABC123-",
			Type:  MapKeyT,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}
	{
		tk := &Token{
			Value: "===",
			Type:  MapKeyT,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}

	{
		tk := &Token{
			Value: "abc:1",
			Type:  MapKeyT,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			Value: "abc:",
			Type:  MapKeyT,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}

	// functionname
	{
		tk := &Token{
			Value: "printPrint123-a_",
			Type:  FunctionNameT,
		}
		err := tk.Validate()
		assert.NoError(t, err)
	}

	{
		tk := &Token{
			Value: "123print",
			Type:  FunctionNameT,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			Value: "print.",
			Type:  FunctionNameT,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
	{
		tk := &Token{
			Value: "print/",
			Type:  FunctionNameT,
		}
		err := tk.Validate()
		assert.Error(t, err)
	}
}
