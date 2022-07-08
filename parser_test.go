package cofunc

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func loadTestingdata(data string) ([]*Block, error) {
	rd := strings.NewReader(data)
	ast, err := ParseAST(rd)
	if err != nil {
		return nil, err
	}
	var blocks []*Block
	ast.Foreach(func(b *Block) error {
		blocks = append(blocks, b)
		return nil
	})
	return blocks, nil
}

func TestParseBlocksFull(t *testing.T) {
	const testingdata string = `
	// Here is a comment
	load "cmd:root/function1"
	load "cmd:url/function2"
	load "cmd:path/function3"
	load "go:function4"
	 
	// 这里是一个注释
	var a = "1"
	var b = "$(a)00"
	var c
	var d="hello word"

	run f1
	run	f2
	run	function3
	run function4

	fn f1 = function1 {
		args = {
			"k1": "v1"
		}
	}
	
	fn f2=function2 {
	}
	`
	blocks, err := loadTestingdata(testingdata)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	_ = blocks
	for _, b := range blocks {
		{
			val, cached := b.CalcVar("a")
			assert.True(t, cached)
			assert.Equal(t, "1", val)
		}
		{
			val, cached := b.CalcVar("b")
			assert.True(t, cached)
			assert.Equal(t, "100", val)
		}
		{
			val, cached := b.CalcVar("c")
			assert.False(t, cached)
			assert.Equal(t, "", val)
		}
		{
			val, cached := b.CalcVar("d")
			assert.True(t, cached)
			assert.Equal(t, "hello word", val)
		}
	}
}

// Only load part
func TestParseBlocksOnlyLoad(t *testing.T) {
	const testingdata string = `
load "cmd:function1"
  load 			 "go:function2"

load "cmd:function3"

	load 	"go:function4"
	`
	blocks, err := loadTestingdata(testingdata)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	check := func(b *Block, path string) {
		assert.Equal(t, "load", b.kind.String())
		assert.Equal(t, path, b.target.String())
	}
	// 0 is global block
	check(blocks[1], "cmd:function1")
	check(blocks[2], "go:function2")
	check(blocks[3], "cmd:function3")
	check(blocks[4], "go:function4")
}

func TestParseBlocksOnlyfn(t *testing.T) {
	const testingdata string = `
	fn f1 = function1 {
		args = {
			"k1":"v1"
			"k3":"v3"
		}
	}

fn f2=function2{ 
}

fn f3 = function3 {
	args = {


	}
}
	`
	blocks, err := loadTestingdata(testingdata)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	_ = blocks
}

func TestParseBlocksFnWithError(t *testing.T) {
	// testingdata is an error data
	{
		const testingdata1 string = `
fn f1= function1 {
	args = {
		"k": "v"
	}


fn f2= function2 {
}
	`
		_, err := loadTestingdata(testingdata1)
		assert.Error(t, err)
	}

	{
		const testingdata2 string = `
	fn f1 = function1 {
		args = {
			"k1":"v1"
			"k2": "v2"
			"k3":"v3"
		}
	}
	}
	`
		_, err := loadTestingdata(testingdata2)
		assert.Error(t, err)
	}
}

func TestParseBlocksOnlyRun(t *testing.T) {
	const testingdata string = `
	run function1
	run 	function2{
		"k1":"v1"
		"k2":"v2"
	}

run function3 {
	"k" : "{(1+2+3)}"

	"multi1": "hello1
hello2
"

	"multi2": "
hello1
hello2
"

	"multi3":"
hello1
hello2"
}

	`
	blocks, err := loadTestingdata(testingdata)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	check := func(b *Block, obj string) {
		assert.Len(t, b.child, 0)
		assert.NotNil(t, b.parent)
		assert.Equal(t, "run", b.kind.String())
		assert.Equal(t, obj, b.target.String())

		if obj == "function2" {
			kvs := b.bbody.(*FMap).ToMap()
			assert.Len(t, kvs, 2)
		}
		if obj == "function3" {
			kvs := b.bbody.(*FMap).ToMap()
			assert.Len(t, kvs, 4)
			assert.Equal(t, "{(1+2+3)}", kvs["k"])
			assert.Equal(t, "hello1\nhello2\n", kvs["multi1"])
			assert.Equal(t, "\nhello1\nhello2\n", kvs["multi2"])
			assert.Equal(t, "\nhello1\nhello2", kvs["multi3"])
		}
	}
	// 0 is global block
	check(blocks[1], "function1")
	check(blocks[2], "function2")
	check(blocks[3], "function3")
}

// Parallel run testing
func TestParseBlocksOnlyRun2(t *testing.T) {
	{
		const testingdata string = `
run {

	function1
	function2

	function3

}
	`
		blocks, err := loadTestingdata(testingdata)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Len(t, blocks, 2)
		// 0 is global block
		b := blocks[1]
		assert.Equal(t, "run", b.kind.String())
		assert.True(t, b.target.IsEmpty())
		assert.True(t, b.operator.IsEmpty())
		assert.True(t, b.typevalue.IsEmpty())

		slice := b.bbody.(*FList).ToSlice()
		assert.Len(t, slice, 3)
		e1, e2, e3 := slice[0], slice[1], slice[2]
		assert.Equal(t, "function1", e1)
		assert.Equal(t, "function2", e2)
		assert.Equal(t, "function3", e3)
	}

	{
		const testingdata string = `
		run{
	function1
	function2

	function3

}
	`
		blocks, err := loadTestingdata(testingdata)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Len(t, blocks, 2)
		// 0 is global block
		b := blocks[1]
		assert.Equal(t, "run", b.kind.String())
		assert.True(t, b.target.IsEmpty())
		assert.True(t, b.operator.IsEmpty())
		assert.True(t, b.typevalue.IsEmpty())

		slice := b.bbody.(*FList).ToSlice()
		assert.Len(t, slice, 3)
		e1, e2, e3 := slice[0], slice[1], slice[2]
		assert.Equal(t, "function1", e1)
		assert.Equal(t, "function2", e2)
		assert.Equal(t, "function3", e3)
	}
}

func TestParseBlocksOnlyRun2WithError(t *testing.T) {
	{
		const testingdata string = `
run {
	function1

	load "xxxx"
	function2

	function3
}
	`
		_, err := loadTestingdata(testingdata)
		assert.Error(t, err)
	}

	{
		const testingdata string = `
run {
	function1
	run function2

	function3
}
	`
		_, err := loadTestingdata(testingdata)
		assert.Error(t, err)
	}

	{
		const testingdata string = `
run {
	function1
	input k v

	function3
}
	`
		_, err := loadTestingdata(testingdata)
		assert.Error(t, err)
	}

	{
		const testingdata string = `
run xyz {
	function1

	function3
}
	`
		_, err := loadTestingdata(testingdata)
		assert.Error(t, err)
	}

	{
		const testingdata string = `
run 3 {
	function1
	function3
}
	`
		_, err := loadTestingdata(testingdata)
		assert.Error(t, err)
	}

}

func TestParseBlocksOnlyRun3(t *testing.T) {
	{
		const testingdata string = `
		var out
run function1 -> out
run function2
	`
		blocks, err := loadTestingdata(testingdata)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Len(t, blocks, 3)
		// 0 is global block
		b := blocks[1]
		assert.Equal(t, "run", b.kind.String())
		assert.Equal(t, "function1", b.target.String())
		assert.Equal(t, "->", b.operator.String())
		assert.Equal(t, "out", b.typevalue.String())
	}
}
