package cofunc

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func loadTestingdata2(data string) ([]*Block, *AST, *RunQueue, error) {
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

func TestParseFullWithRunq(t *testing.T) {
	{
		const testingdata string = `
	load go:function1
	load go:function2
	load cmd:/tmp/function3
	load cmd:/tmp/function4
	load cmd:/tmp/function5

	fn f1 = function1 {
		args = {
			k: v1
			"hello": "world"
		}
	}

	run f1
	run	function2 {
		k : v2
	}
	run	function3
	run {
		function4
		function5
	}
	run	function3 {
		k: v3
	}
	`

		blocks, bl, rq, err := loadTestingdata2(testingdata)
		assert.NoError(t, err)
		assert.NotNil(t, blocks)
		assert.NotNil(t, bl)
		assert.NotNil(t, rq)

		assert.Len(t, rq.configuredNodes, 1)
		assert.Equal(t, "function1", rq.configuredNodes["f1"].Driver.FunctionName())
		assert.Len(t, rq.queue, 5)

		rq.Forstage(func(stage int, node *Node) error {
			if stage == 1 {
				assert.Equal(t, "f1", node.Name)
				assert.Len(t, node.args, 2)
				assert.Equal(t, "v1", node.args["k"])
			}
			if stage == 2 {
				assert.Equal(t, "function2", node.Name)
				assert.Len(t, node.args, 1)
				assert.Equal(t, "v2", node.args["k"])
			}
			if stage == 3 {
				assert.Equal(t, "function3", node.Name)
				assert.Len(t, node.args, 0)
			}
			if stage == 4 {
				assert.Equal(t, "function4", node.Name)
				assert.NotNil(t, node.Parallel)
				assert.Equal(t, "function5", node.Parallel.Name)
			}
			if stage == 5 {
				assert.Equal(t, "function3", node.Name)
				assert.Len(t, node.args, 1)
				assert.Equal(t, "v3", node.args["k"])
			}
			return nil
		})
	}
}

func TestParseFullWithRunqWithErr(t *testing.T) {
	{
		const testingdata string = `
	load go:function1
	load go:function2

	fn function1 = function1 {
		args = {

		}
	}

	run function1
	`

		blocks, bl, rq, err := loadTestingdata2(testingdata)
		assert.Error(t, err)
		_ = blocks
		_ = bl
		_ = rq
	}
}
