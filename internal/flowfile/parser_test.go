package flowfile

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Only load part
func TestParseBlocksOnlyLoad(t *testing.T) {
	const testingdataForLoadPart string = `
load file:///root/action1
  load http://localhost:8080/action2

load https://github.com/path/action3

	load 	action4
	`
	rd := strings.NewReader(testingdataForLoadPart)
	bs, err := ParseBlocks(rd)
	if err != nil {
		assert.FailNow(t, err.Error())
	}

	assert.Equal(t, 4, bs.BlockNum())

	var blocks []*Block
	bs.Foreach(func(b *Block) {
		blocks = append(blocks, b)
	})
	check := func(b *Block, path string) {
		assert.Equal(t, _block_load, b.kind)
		assert.Len(t, b.tokens, 1)
		line := b.tokens[1]
		assert.Len(t, line, 2)

		assert.Equal(t, "load", line[0].word)
		assert.Equal(t, path, line[1].word)

		assert.True(t, line[0].keyword)
		assert.False(t, line[1].keyword)
	}
	check(blocks[0], "file:///root/action1")
	check(blocks[1], "http://localhost:8080/action2")
	check(blocks[2], "https://github.com/path/action3")
	check(blocks[3], "action4")
}
