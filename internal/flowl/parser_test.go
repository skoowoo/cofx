package flowl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func loadTestingdata(data string) ([]*Block, error) {
	rd := strings.NewReader(data)
	bs, err := ParseBlocks(rd)
	if err != nil {
		return nil, err
	}
	var blocks []*Block
	bs.Foreach(func(b *Block) error {
		blocks = append(blocks, b)
		return nil
	})
	return blocks, nil
}

func TestParseBlocksFull(t *testing.T) {
	const testingdata string = `
	load cmd:root/function1
	load cmd:url/function2
	load cmd:path/function3
	load go:function4
	 
	set @function1 {
		input k1 v1
		input k3 v3
		input k $v
	
		loop 5 2
	}
	
	set @function2 {
		input k $v
	
		input function1_out $out1
	}
	
	run @function1
	run	@function2
	run	@function3
	run @function4
	`
	blocks, err := loadTestingdata(testingdata)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	_ = blocks
}

// Only load part
func TestParseBlocksOnlyLoad(t *testing.T) {
	const testingdata string = `
load file:///root/function1
  load http://localhost:8080/function2

load https://github.com/path/function3

	load 	function4
	`
	blocks, err := loadTestingdata(testingdata)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	check := func(b *Block, path string) {
		assert.Equal(t, _block_load, b.kind)
		assert.Len(t, b.directives, 1)
		tokens := b.directives[0].tokens
		assert.Len(t, tokens, 2)

		assert.Equal(t, "load", tokens[0].text)
		assert.Equal(t, path, tokens[1].text)

		assert.True(t, tokens[0].keyword)
		assert.False(t, tokens[1].keyword)
	}
	check(blocks[0], "file:///root/function1")
	check(blocks[1], "http://localhost:8080/function2")
	check(blocks[2], "https://github.com/path/function3")
	check(blocks[3], "function4")
}

func TestParseBlocksOnlySet(t *testing.T) {
	const testingdata string = `
	set @function1 {
	input k1 v1
	input k3 v3
	input k $v

	loop 5 2
	}

set @function2 { 
	input k $v
	
	input function1_out $out1
}
	`
	blocks, err := loadTestingdata(testingdata)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	_ = blocks
}

func TestParseBlocksSetWithError(t *testing.T) {
	// testingdata is an error data
	const testingdata1 string = `
set @function1 {
	input k1 v1
	input k3 v3
	input k $v

	loop 5 2


set @function2 {
	input k $v
	
	input function1_out $out1
}
	`
	_, err := loadTestingdata(testingdata1)
	assert.Error(t, err)

	const testingdata2 string = `
	set @function1 {
		input k1 v1
		input k3 v3
		input k $v
	
		loop 5 2
	}

	}
	
	set @function2  {
		input k $v
		
		input function1_out $out1
	}
	`
	_, err = loadTestingdata(testingdata2)
	assert.Error(t, err)

}

func TestParseBlocksOnlyRun(t *testing.T) {
	const testingdata string = `
	run @function1
	run @function2

run @function3

	`
	blocks, err := loadTestingdata(testingdata)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	check := func(b *Block, arg string) {
		assert.Equal(t, _block_run, b.kind)
		assert.Len(t, b.directives, 1)
		tokens := b.directives[0].tokens
		assert.Len(t, tokens, 2)

		assert.Equal(t, "run", tokens[0].text)
		assert.Equal(t, arg, tokens[1].text)

		assert.True(t, tokens[0].keyword)
		assert.False(t, tokens[1].keyword)
	}
	check(blocks[0], "@function1")
	check(blocks[1], "@function2")
	check(blocks[2], "@function3")
}

// Parallel run testing
func TestParseBlocksOnlyRun2(t *testing.T) {
	const testingdata string = `
run {
	@function1
	@function2

	@function3
}
	`
	blocks, err := loadTestingdata(testingdata)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	assert.Len(t, blocks, 1)
}

func TestParseBlocksOnlyRun2WithError(t *testing.T) {
	{
		const testingdata string = `
run 3 {
	@function1
	load file:///root/function1
	@function2

	@function3
}
	`
		_, err := loadTestingdata(testingdata)
		assert.Error(t, err)
	}

	{
		const testingdata string = `
run 3 {
	@function1
	run @function2

	@function3
}
	`
		_, err := loadTestingdata(testingdata)
		assert.Error(t, err)
	}

	{
		const testingdata string = `
run {
	@function1
	input k v

	@function3
}
	`
		_, err := loadTestingdata(testingdata)
		assert.Error(t, err)
	}

	{
		const testingdata string = `
run xyz {
	@function1

	@function3
}
	`
		_, err := loadTestingdata(testingdata)
		assert.Error(t, err)
	}

	{
		const testingdata string = `
run 3 {
	@function1
	@function3
}
	`
		_, err := loadTestingdata(testingdata)
		assert.Error(t, err)
	}

}

//
// Test Run queue and blockstore
func loadTestingdata2(data string) ([]*Block, *BlockStore, *RunQueue, error) {
	rd := strings.NewReader(data)
	rq, bs, err := Parse(rd)
	if err != nil {
		return nil, nil, nil, err
	}
	var blocks []*Block
	bs.Foreach(func(b *Block) error {
		blocks = append(blocks, b)
		return nil
	})
	return blocks, bs, rq, nil
}

func TestParseFull(t *testing.T) {
	const testingdata string = `
	load go:function1
	load go:function2
	load cmd:/tmp/function3
	load cmd:/tmp/function4
	load cmd:/tmp/function5

	set function1 {
		input k1 v1
		input k2 v2
	}

	run @function1
	run	@function2
	run	@function3
	run {
		@function4
		@function5
	}
	`

	blocks, bs, rq, err := loadTestingdata2(testingdata)
	assert.NoError(t, err)
	assert.NotNil(t, blocks)
	assert.NotNil(t, bs)
	assert.NotNil(t, rq)

	assert.Len(t, rq.Functions, 5)

	assert.Equal(t, "function1", rq.Functions["function1"].Name)
	assert.Equal(t, "function3", rq.Functions["function3"].Name)

	assert.Len(t, rq.Functions["function1"].Args(), 2)
	assert.Equal(t, 4, rq.Queue.Len())
}
