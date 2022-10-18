package runcmd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRuncmd(t *testing.T) {
	wrap := Wrap{
		Name:         "echo",
		Args:         []string{"hello\nworld"},
		Split:        "",
		Extract:      []int{0},
		QueryColumns: []string{"c0"},
		QueryWhere:   "",
	}
	rows, err := wrap.Run(context.Background())
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Len(t, rows, 2)
	assert.Equal(t, "hello", rows[0][0])
	assert.Equal(t, "world", rows[1][0])
}
