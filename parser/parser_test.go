package parser

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func loadTestingdata(data string) ([]*Block, error) {
	rd := strings.NewReader(data)
	ast, err := New(rd)
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
	load "shell:root/function1"
	load "shell:url/function2"
	load "shell:path/function3"
	load "go:function4"
	 
	// 这里是一个注释
	var a = "1"
	var b = "$(a)00"
	var c
	var d="hello word"

	c <- "test"

	co f1
	co f2
	co function3

	switch {
		case "a" > "b" {
			co function4
		}
	}

	for {
		switch {
			case 2 > 1 {
				co f1
			}
		}
		co f2
	}

	//---
	fn f1 = function1 {
		args = {
			"k1": "v1"
		}
		var fa = "f1"
		var fb = $(fa)
	}
	
	fn f2=function2 {
	}
	`
	blocks, err := loadTestingdata(testingdata)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	for _, b := range blocks {
		{
			val, cached := b.calcVar("a")
			assert.True(t, cached)
			assert.Equal(t, "1", val)
		}
		{
			val, cached := b.calcVar("b")
			assert.True(t, cached)
			assert.Equal(t, "100", val)
		}
		{
			val, cached := b.calcVar("c")
			assert.True(t, cached)
			assert.Equal(t, "", val)
		}
		{
			val, cached := b.calcVar("d")
			assert.True(t, cached)
			assert.Equal(t, "hello word", val)
		}

		if b.IsFn() && b.target1.String() == "f1" {
			{
				val, cached := b.calcVar("fa")
				assert.True(t, cached)
				assert.Equal(t, "f1", val)
			}
			{
				val, cached := b.calcVar("fb")
				assert.True(t, cached)
				assert.Equal(t, "f1", val)
			}
		}
	}
}

// Only load part
func TestParseBlocksOnlyLoad(t *testing.T) {
	const testingdata string = `
load "shell:function1"
  load 			 "go:function2"

load "shell:function3"

	load 	"go:function4"
	`
	blocks, err := loadTestingdata(testingdata)
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	check := func(b *Block, path string) {
		assert.True(t, b.IsLoad())
		assert.Equal(t, path, b.target1.String())
	}
	// 0 is global block
	check(blocks[1], "shell:function1")
	check(blocks[2], "go:function2")
	check(blocks[3], "shell:function3")
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

func TestParseBlocksOnlyco(t *testing.T) {
	const testingdata string = `
	co function1
	co 	function2{
		"k1":"v1"
		"k2":"v2"
	}

co function3 {
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
		assert.True(t, b.IsCo())
		assert.Equal(t, obj, b.target1.String())

		if obj == "function2" {
			kvs := b.body.(*MapBody).ToMap()
			assert.Len(t, kvs, 2)
		}
		if obj == "function3" {
			kvs := b.body.(*MapBody).ToMap()
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

// Parallel co testing
func TestParseBlocksOnlyco2(t *testing.T) {
	{
		const testingdata string = `
co {

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
		assert.True(t, b.IsCo())
		assert.True(t, b.target1.IsEmpty())
		assert.True(t, b.operator.IsEmpty())
		assert.True(t, b.target2.IsEmpty())

		slice := b.body.(*ListBody).ToSlice()
		assert.Len(t, slice, 3)
		e1, e2, e3 := slice[0], slice[1], slice[2]
		assert.Equal(t, "function1", e1)
		assert.Equal(t, "function2", e2)
		assert.Equal(t, "function3", e3)
	}

	{
		const testingdata string = `
		co{
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
		assert.True(t, b.IsCo())
		assert.True(t, b.target1.IsEmpty())
		assert.True(t, b.operator.IsEmpty())
		assert.True(t, b.target2.IsEmpty())

		slice := b.body.(*ListBody).ToSlice()
		assert.Len(t, slice, 3)
		e1, e2, e3 := slice[0], slice[1], slice[2]
		assert.Equal(t, "function1", e1)
		assert.Equal(t, "function2", e2)
		assert.Equal(t, "function3", e3)
	}
}

func TestParseBlocksOnlyco2WithError(t *testing.T) {
	{
		const testingdata string = `
co {
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
co {
	function1
	co function2

	function3
}
	`
		_, err := loadTestingdata(testingdata)
		assert.Error(t, err)
	}

	{
		const testingdata string = `
co {
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
co xyz {
	function1

	function3
}
	`
		_, err := loadTestingdata(testingdata)
		assert.Error(t, err)
	}

	{
		const testingdata string = `
co 3 {
	function1
	function3
}
	`
		_, err := loadTestingdata(testingdata)
		assert.Error(t, err)
	}

}

func TestParseBlocksOnlyco3(t *testing.T) {
	{
		const testingdata string = `
		var out
co function1 -> out
co function2
	`
		blocks, err := loadTestingdata(testingdata)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Len(t, blocks, 3)
		// 0 is global block
		b := blocks[1]
		assert.True(t, b.IsCo())
		assert.Equal(t, "function1", b.target1.String())
		assert.Equal(t, "->", b.operator.String())
		assert.Equal(t, "out", b.target2.String())
	}
}

func TestParseBlocksOnlyFor(t *testing.T) {
	{
		const testingdata string = `
var out

for {
	co function1 -> out
	co function2
}
	`
		blocks, err := loadTestingdata(testingdata)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.Len(t, blocks, 5)
		// 0 is global block
		b := blocks[1]
		assert.True(t, b.IsFor())
	}
}

func TestParseBlocksOnlyForErr(t *testing.T) {
	{
		const testingdata string = `
for {
	co function1 -> out
	co function2
}
	`
		_, err := loadTestingdata(testingdata)
		assert.Error(t, err)
	}
}

func TestInferTree(t *testing.T) {
	{
		var tokens []*Token = []*Token{
			{
				str: "a",
				typ: _ident_t,
			},
			{
				str: "<-",
				typ: _symbol_t,
			},
			{
				str: "hello",
				typ: _string_t,
			},
		}
		infer := buildInferTree()
		_, err := infer.lookup(tokens)
		assert.NoError(t, err)
	}
}

func TestVarCycleCheck(t *testing.T) {
	// {
	// 	const testingdata string = `
	// 	var a = "1"
	// 	var b = $(a)
	// 	var c = $(b)

	// 	a <- $(c)
	// `
	// 	_, err := loadTestingdata(testingdata)
	// 	assert.Error(t, err)
	// }
	// {
	// 	const testingdata string = `
	// 	var a = "1"
	// 	var b = "$(a)"
	// 	a <- $(b)
	// `
	// 	_, err := loadTestingdata(testingdata)
	// 	assert.Error(t, err)
	// }
	{
		const testingdata string = `
		var a = "1"
		var b = "$(a)"
		a <- "$(a) 2"
	`
		_, err := loadTestingdata(testingdata)
		assert.NoError(t, err)
	}
}

func TestVarDefine(t *testing.T) {
	{
		const testingdata string = `
		var a = 100
		var b = $(a)
		var c = "$(a)"
		var d = "foo"
		var e = 0.1
		var f = 1 + 1
		var g = $(a) + 1
	`
		_, err := loadTestingdata(testingdata)
		assert.NoError(t, err)
	}
	{
		const testingdata string = `
		var a = a100
		var d = foo
	`
		_, err := loadTestingdata(testingdata)
		assert.Error(t, err)
	}
}

func TestRewriteVar(t *testing.T) {
	{
		const testingdata string = `
		var a = 100
		var b = $(a)
		var c = "$(a)"
		var d = "foo"
		var e = 0.1
		var f = 1 + 1
		var g = $(a) + 1

		a <- 100
		b <- "bar"
		c <- -1
		c <- 1 + 1
		c <- (1 + 1)
	`
		_, err := loadTestingdata(testingdata)
		assert.NoError(t, err)
	}
}

func TestVarExpression(t *testing.T) {
	{
		const testingdata string = `
		var a = 100
		var b = (($(a) + 1) * 1024) / (512 * 2)

		var c = (2+1)
		var d = 2 > 1
		var e = "a" > "b"
	`
		blocks, err := loadTestingdata(testingdata)
		if err != nil {
			t.FailNow()
		}
		for _, b := range blocks {
			v, _ := b.calcVar("a")
			assert.Equal(t, "100", v)

			v, _ = b.calcVar("b")
			assert.Equal(t, "101", v)

			v, _ = b.calcVar("c")
			assert.Equal(t, "3", v)

			v, _ = b.calcVar("d")
			assert.Equal(t, "true", v)

			v, _ = b.calcVar("e")
			assert.Equal(t, "false", v)
		}
	}
}

func TestEvent(t *testing.T) {
	{
		const testingdata string = `
		var out
		event {
			co function1 -> out
		}
	`
		blocks, err := loadTestingdata(testingdata)
		assert.NoError(t, err)

		assert.Len(t, blocks, 3)
		assert.Equal(t, "global", blocks[0].kind.String())
		assert.Equal(t, _kw_event, blocks[1].kind.String())
		assert.Equal(t, _kw_co, blocks[2].kind.String())
	}
	{
		const testingdata string = `
		var out
		event {
			co function1 -> out
			co function2 -> out
			co function3 -> out
		}
	`
		blocks, err := loadTestingdata(testingdata)
		assert.NoError(t, err)

		assert.Len(t, blocks, 5)
		assert.Equal(t, "global", blocks[0].kind.String())
		assert.Equal(t, _kw_event, blocks[1].kind.String())
		assert.Equal(t, _kw_co, blocks[2].kind.String())
		assert.Equal(t, _kw_co, blocks[3].kind.String())
		assert.Equal(t, _kw_co, blocks[4].kind.String())
	}
	{
		const testingdata string = `
		var out
		event {
			co function1 -> out
			co function2 -> out {
				"k": "v"
			}
		}
	`
		blocks, err := loadTestingdata(testingdata)
		assert.NoError(t, err)

		assert.Len(t, blocks, 4)
		assert.Equal(t, "global", blocks[0].kind.String())
		assert.Equal(t, _kw_event, blocks[1].kind.String())
		assert.Equal(t, _kw_co, blocks[2].kind.String())
		assert.Equal(t, _kw_co, blocks[3].kind.String())
	}
}

func TestIf(t *testing.T) {
	{
		const testingdata string = `
		var v = 1
		if $(v) == 1 {
			co function
		}
	`
		blocks, err := loadTestingdata(testingdata)
		assert.NoError(t, err)

		assert.Len(t, blocks, 3)
		assert.Equal(t, "global", blocks[0].kind.String())
		assert.Equal(t, _kw_if, blocks[1].kind.String())
		assert.Equal(t, _kw_co, blocks[2].kind.String())

		cob := blocks[2]
		assert.True(t, cob.ExecCondition())
	}
	{
		const testingdata string = `
		var v = "1"
		if $(v) == 1 {
			co function
		}
	`
		blocks, err := loadTestingdata(testingdata)
		assert.NoError(t, err)

		assert.Len(t, blocks, 3)
		assert.Equal(t, "global", blocks[0].kind.String())
		assert.Equal(t, _kw_if, blocks[1].kind.String())
		assert.Equal(t, _kw_co, blocks[2].kind.String())

		cob := blocks[2]
		assert.True(t, cob.ExecCondition())
	}
	{
		const testingdata string = `
		var v = "1"
		if "$(v)" == "1" {
			co function
		}
	`
		blocks, err := loadTestingdata(testingdata)
		assert.NoError(t, err)

		assert.Len(t, blocks, 3)
		assert.Equal(t, "global", blocks[0].kind.String())
		assert.Equal(t, _kw_if, blocks[1].kind.String())
		assert.Equal(t, _kw_co, blocks[2].kind.String())

		cob := blocks[2]
		assert.True(t, cob.ExecCondition())
	}
	{
		const testingdata string = `
		var a = 1
		var b = 2
		if $(b) == $(a) + 1 {
			co function
		}
	`
		blocks, err := loadTestingdata(testingdata)
		assert.NoError(t, err)

		assert.Len(t, blocks, 3)
		assert.Equal(t, "global", blocks[0].kind.String())
		assert.Equal(t, _kw_if, blocks[1].kind.String())
		assert.Equal(t, _kw_co, blocks[2].kind.String())

		cob := blocks[2]
		assert.True(t, cob.ExecCondition())
	}
	{
		const testingdata string = `
		var a = "hello"
		if $(a) == "" {
			co function
		}
	`
		blocks, err := loadTestingdata(testingdata)
		assert.NoError(t, err)

		assert.Len(t, blocks, 3)
		assert.Equal(t, "global", blocks[0].kind.String())
		assert.Equal(t, _kw_if, blocks[1].kind.String())
		assert.Equal(t, _kw_co, blocks[2].kind.String())

		cob := blocks[2]
		assert.False(t, cob.ExecCondition())
	}
}

func TestSWitchCase(t *testing.T) {
	{
		const testingdata string = `
		var v = 1
		switch {
			case $(v) == 1 {
				co function
			}
		}
	`
		blocks, err := loadTestingdata(testingdata)
		assert.NoError(t, err)

		assert.Len(t, blocks, 4)
		assert.Equal(t, "global", blocks[0].kind.String())
		assert.Equal(t, _kw_switch, blocks[1].kind.String())
		assert.Equal(t, _kw_case, blocks[2].kind.String())
		assert.Equal(t, _kw_co, blocks[3].kind.String())

		cob := blocks[3]
		assert.True(t, cob.ExecCondition())
	}
	{
		const testingdata string = `
		var a = "hello"
		switch {
			case $(a) == "" {
				co function
			}
			default{
				co function2
			}
		}
	`
		blocks, err := loadTestingdata(testingdata)
		assert.NoError(t, err)

		assert.Len(t, blocks, 6)
		assert.Equal(t, "global", blocks[0].kind.String())
		assert.Equal(t, _kw_switch, blocks[1].kind.String())
		assert.Equal(t, _kw_case, blocks[2].kind.String())
		assert.Equal(t, _kw_co, blocks[3].kind.String())
		assert.Equal(t, _kw_default, blocks[4].kind.String())
		assert.Equal(t, _kw_co, blocks[5].kind.String())

		cob1 := blocks[3]
		assert.False(t, cob1.ExecCondition())
		cob2 := blocks[5]
		assert.True(t, cob2.ExecCondition())
	}
}
