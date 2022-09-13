package gobuild

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindModAndMainPkg(t *testing.T) {
	mods, err := findMods(".", []string{"."})
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Len(t, mods, 1)
	assert.Equal(t, "testing/testingmain", mods["testingdata"].mainpkgs[0])
}
