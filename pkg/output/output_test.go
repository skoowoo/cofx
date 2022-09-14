package output

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOutput(t *testing.T) {
	{
		expected := map[string]struct{}{
			"hello\n":      {},
			"world\n":      {},
			"helloworld\n": {},
			"foo":          {},
		}
		data1 := "hello\nworld\n"
		data2 := "hello"
		data3 := "world\n"
		data4 := "foo"

		lines := 0
		out := &Output{
			W: os.Stdout,
			HandleFunc: func(line []byte) {
				lines++
				_, ok := expected[string(line)]
				assert.Equal(t, true, ok)
			},
		}
		out.Write([]byte(data1))
		out.Write([]byte(data2))
		out.Write([]byte(data3))
		out.Write([]byte(data4))
		out.Close()

		assert.Equal(t, 4, lines)
	}
}

func TestOutput2(t *testing.T) {
	{
		expected := map[string]struct{}{
			"hello\n":      {},
			"world\n":      {},
			"helloworld\n": {},
			"foo":          {},
		}
		data1 := "hello\n"

		var buf bytes.Buffer
		lines := 0
		out := &Output{
			W: &buf,
			HandleFunc: func(line []byte) {
				lines++
				_, ok := expected[string(line)]
				assert.Equal(t, true, ok)
			},
		}
		out.Write([]byte(data1))
		out.Close()

		assert.Equal(t, 1, lines)
		assert.Equal(t, "hello", strings.TrimSpace(buf.String()))
	}
}

func TestColumnFunc(t *testing.T) {
	testingdata := `
origin  https://github.com/skoo87/cofunc.git (fetch)
origin  https://github.com/skoo87/cofunc.git (push)
upstream        https://github.com/cofunclabs/cofunc.git (fetch)
upstream        https://github.com/cofunclabs/cofunc.git (push)	
`

	var (
		rows [][]string
		sep  = " "
	)
	out := &Output{
		W: nil,
		HandleFunc: ColumnFunc(sep, func(fields []string) {
			if fields[0] == "upstream" && strings.Contains(fields[2], "fetch") {
				rows = append(rows, fields)
			}
		}, 0, 1, 2),
	}
	out.Write([]byte(testingdata))
	out.Close()

	assert.Len(t, rows, 1)
}
