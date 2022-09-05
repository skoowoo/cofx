package std

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStdInit(t *testing.T) {
	assert.Equal(t, len(builtin), len(entrypoints))
	for k, m := range builtin {
		assert.Equal(t, k, m.Name)

		ep := m.Entrypoint
		_, ok := entrypoints[ep]
		assert.Equal(t, true, ok)
	}
}
