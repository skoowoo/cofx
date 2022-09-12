package gobuild

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindMainPkg(t *testing.T) {
	pkgs, err := findMainPkg("github.com/prefix", ".")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Len(t, pkgs, 1)
	assert.Equal(t, "github.com/prefix/testingdata/testingmain", pkgs[0])
}
