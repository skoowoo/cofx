package cofunc

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func loadTestingdata2(data string) ([]*Block, *AST, *RunQ, error) {
	rd := strings.NewReader(data)
	rq, bl, err := ParseFlowl(rd)
	if err != nil {
		return nil, nil, nil, err
	}
	var blocks []*Block
	bl.Foreach(func(b *Block) error {
		blocks = append(blocks, b)
		return nil
	})
	return blocks, bl, rq, nil
}

func TestForLoopWithRunq(t *testing.T) {
	{
		const testingdata string = `
load "go:print"
load "go:sleep"
load "go:time"

var t

for {
    co time -> t
    co print {
        "Time": "$(t.Now)"
    }
    co sleep
}
		`
		blocks, bl, rq, err := loadTestingdata2(testingdata)
		assert.NoError(t, err)
		assert.NotNil(t, blocks)
		assert.NotNil(t, bl)
		assert.NotNil(t, rq)

		assert.Len(t, rq.stages, 5)
		for_node := rq.stages[0].(*ForNode)
		assert.Equal(t, "FOR", for_node.Name())
		assert.Equal(t, "time", rq.stages[1].Name())
		assert.Equal(t, "print", rq.stages[2].Name())
		assert.Equal(t, "sleep", rq.stages[3].Name())
		btf_node := rq.stages[4].(*BtfNode)
		assert.Equal(t, "BTF", btf_node.Name())

		assert.Equal(t, 0, for_node.idx)
		assert.Equal(t, 4, for_node.btfIdx)
		assert.Equal(t, 4, btf_node.idx)
		assert.Equal(t, 0, btf_node.forIdx)
	}
}

func TestParseFullWithRunq(t *testing.T) {
	{
		const testingdata string = `
	load "go:function1"
	load "go:function2"
	load "cmd:/tmp/function3"
	load "cmd:/tmp/function4"
	load "cmd:/tmp/function5"

	fn f1 = function1 {
		args = {
			"k": "v1"
			"hello": "world"
		}
	}

	co f1
	co	function2 {
		"k" : "v2"
	}
	co	function3
	co {
		function4
		function5
	}
	co	function3 {
		"k": "v3"
	}
	`

		blocks, bl, rq, err := loadTestingdata2(testingdata)
		assert.NoError(t, err)
		assert.NotNil(t, blocks)
		assert.NotNil(t, bl)
		assert.NotNil(t, rq)

		assert.Len(t, rq.configuredNodes, 1)
		assert.Equal(t, "function1", rq.configuredNodes["f1"].driver.FunctionName())
		assert.Len(t, rq.stages, 5)

		rq.ForstageAndExec(func(stage int, nodes []*FuncNode) error {
			if stage == 1 {
				node := nodes[0]
				assert.Equal(t, "f1", node.name)
				assert.Len(t, node.Args(), 2)
				assert.Equal(t, "v1", node.Args()["k"])
			}
			if stage == 2 {
				node := nodes[0]
				assert.Equal(t, "function2", node.name)
				assert.Len(t, node.Args(), 1)
				assert.Equal(t, "v2", node.Args()["k"])
			}
			if stage == 3 {
				node := nodes[0]
				assert.Equal(t, "function3", node.name)
				assert.Len(t, node.Args(), 0)
			}
			if stage == 4 {
				node := nodes[0]
				assert.Equal(t, "function4", node.name)
				assert.NotNil(t, node.parallel)
				assert.Equal(t, "function5", node.parallel.name)
			}
			if stage == 5 {
				node := nodes[0]
				assert.Equal(t, "function3", node.name)
				assert.Len(t, node.Args(), 1)
				assert.Equal(t, "v3", node.Args()["k"])
			}
			return nil
		})
	}
}

func TestParseFullWithRunqWithErr(t *testing.T) {
	{
		const testingdata string = `
	load "go:function1"
	load "go:function2"

	fn function1 = function1 {
		args = {

		}
	}

	co function1
	`

		blocks, bl, rq, err := loadTestingdata2(testingdata)
		assert.Error(t, err)
		_ = blocks
		_ = bl
		_ = rq
	}
}
