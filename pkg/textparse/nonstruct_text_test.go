package textparse

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNonStructParse(t *testing.T) {
	{
		text := `
module github.com/skoowoo/cofx

	`

		nst, err := New("gomod", "", []int{0, 1})
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		if err := nst.ParseLine(context.Background(), text); err != nil {
			assert.FailNow(t, err.Error())
		}
		v, err := nst.String(context.Background(), "c1", "c0 = 'module'")
		assert.NoError(t, err)
		assert.Equal(t, "github.com/skoowoo/cofx", v)

		// clear
		err = nst.Clear(context.Background())
		assert.NoError(t, err)
	}

	{
		text := `go 1.18

require (
	github.com/charmbracelet/bubbletea v0.22.0
	github.com/charmbracelet/lipgloss v0.5.0
	github.com/glebarez/go-sqlite v1.18.2
	github.com/lucasb-eyer/go-colorful v1.2.0
	github.com/mitchellh/go-homedir v1.1.0
)`

		nst, err := New("gomod", "/", []int{0, 1, 2})
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		lines := strings.Split(text, "\n")
		for _, l := range lines {
			if err := nst.ParseLine(context.Background(), l); err != nil {
				assert.FailNow(t, err.Error())
			}
		}

		{
			rows, err := nst.Query(context.Background(), []string{"c1"})
			assert.NoError(t, err)
			assert.Equal(t, 9, len(rows))
		}

		{
			rows, err := nst.Query(context.Background(), []string{"c1"}, "c0 = 'github.com'")
			assert.NoError(t, err)
			assert.Equal(t, 5, len(rows))
		}

		// clear
		err = nst.Clear(context.Background())
		assert.NoError(t, err)
	}
}
