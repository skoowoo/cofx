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
	bs.Foreach(func(b *Block) {
		blocks = append(blocks, b)
	})
	return blocks, nil
}

func TestParseBlocksFull(t *testing.T) {
	const testingdata string = `
	load file:///root/action1
	load http://url/action2
	load https://github.com/path/action3
	load action4
	 
	set @action1 {
		input k1 v1
		input k3 v3
		input k $v
	
		loop 5 2
	}
	
	set @action2 {
		input k $v
	
		input action1_out $out1
	}
	
	run @action1
	run	@action2
	run	@action3
	run @action4
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
load file:///root/action1
  load http://localhost:8080/action2

load https://github.com/path/action3

	load 	action4
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

		assert.Equal(t, "load", tokens[0].value)
		assert.Equal(t, path, tokens[1].value)

		assert.True(t, tokens[0].keyword)
		assert.False(t, tokens[1].keyword)
	}
	check(blocks[0], "file:///root/action1")
	check(blocks[1], "http://localhost:8080/action2")
	check(blocks[2], "https://github.com/path/action3")
	check(blocks[3], "action4")
}

func TestParseBlocksOnlySet(t *testing.T) {
	const testingdata string = `
	set @action1 {
	input k1 v1
	input k3 v3
	input k $v

	loop 5 2
	}

set @action2 { 
	input k $v
	
	input action1_out $out1
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
set @action1 {
	input k1 v1
	input k3 v3
	input k $v

	loop 5 2


set @action2 {
	input k $v
	
	input action1_out $out1
}
	`
	_, err := loadTestingdata(testingdata1)
	assert.Error(t, err)

	const testingdata2 string = `
	set @action1 {
		input k1 v1
		input k3 v3
		input k $v
	
		loop 5 2
	}

	}
	
	set @action2  {
		input k $v
		
		input action1_out $out1
	}
	`
	_, err = loadTestingdata(testingdata2)
	assert.Error(t, err)

}

func TestParseBlocksOnlyRun(t *testing.T) {
	const testingdata string = `
	run @action1
	run @action2

run @action3

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

		assert.Equal(t, "run", tokens[0].value)
		assert.Equal(t, arg, tokens[1].value)

		assert.True(t, tokens[0].keyword)
		assert.False(t, tokens[1].keyword)
	}
	check(blocks[0], "@action1")
	check(blocks[1], "@action2")
	check(blocks[2], "@action3")
}

// Parallel run testing
func TestParseBlocksOnlyRun2(t *testing.T) {
	const testingdata string = `
run {
	@action1
	@action2

	@action3
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
	@action1
	load file:///root/action1
	@action2

	@action3
}
	`
		_, err := loadTestingdata(testingdata)
		assert.Error(t, err)
	}

	{
		const testingdata string = `
run 3 {
	@action1
	run @action2

	@action3
}
	`
		_, err := loadTestingdata(testingdata)
		assert.Error(t, err)
	}

	{
		const testingdata string = `
run {
	@action1
	input k v

	@action3
}
	`
		_, err := loadTestingdata(testingdata)
		assert.Error(t, err)
	}

	{
		const testingdata string = `
run xyz {
	@action1

	@action3
}
	`
		_, err := loadTestingdata(testingdata)
		assert.Error(t, err)
	}

	{
		const testingdata string = `
run 3 {
	@action1
	@action3
}
	`
		_, err := loadTestingdata(testingdata)
		assert.Error(t, err)
	}

}
