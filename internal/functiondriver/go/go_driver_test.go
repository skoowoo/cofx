package godriver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	dr := New("go:print")
	if dr == nil {
		t.FailNow()
	}
	assert.Equal(t, "print", dr.funcName)
	assert.Equal(t, "print", dr.location)
	err := dr.Load(map[string]string{
		"k1": "v1",
		"k2": "v2",
		"k3": "v3",
	})
	assert.NoError(t, err)
	assert.Len(t, dr.manifest.Args, 3)
	assert.NotNil(t, dr.fn)
}

func TestRun(t *testing.T) {
	dr := New("go:print")
	if dr == nil {
		t.FailNow()
	}
	err := dr.Load(map[string]string{
		"k1": "v1",
		"k2": "v2",
		"k3": "v3",
	})
	assert.NoError(t, err)
	out, err := dr.Run()
	assert.NoError(t, err)
	assert.Equal(t, "ok", out["status"])
}
