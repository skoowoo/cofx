package actuator

import (
	"context"
	"strings"
	"testing"

	"github.com/cofxlabs/cofx/parser"
	"github.com/stretchr/testify/assert"
)

func loadTestingdata2(data string) ([]*parser.Block, *parser.AST, *RunQueue, error) {
	rd := strings.NewReader(data)
	rq, ast, err := New(rd)
	if err != nil {
		return nil, nil, nil, err
	}
	err = rq.WalkNode(func(n Node) error {
		return n.Init(context.TODO())
	})
	if err != nil {
		return nil, nil, nil, err
	}
	var blocks []*parser.Block
	ast.Foreach(func(b *parser.Block) error {
		blocks = append(blocks, b)
		return nil
	})
	return blocks, ast, rq, nil
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
		if err != nil {
			t.FailNow()
		}
		assert.NoError(t, err)
		assert.NotNil(t, blocks)
		assert.NotNil(t, bl)
		assert.NotNil(t, rq)

		assert.Len(t, rq.steps, 5)
		for_node := rq.steps[0].(*ForNode)
		assert.Equal(t, "FOR", for_node.Name())
		assert.Equal(t, "time", rq.steps[1].Name())
		assert.Equal(t, "print", rq.steps[2].Name())
		assert.Equal(t, "sleep", rq.steps[3].Name())
		btf_node := rq.steps[4].(*BtfNode)
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
	load "shell:/tmp/function3"
	load "shell:/tmp/function4"
	load "shell:/tmp/function5"

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
		if err != nil {
			t.Logf("%v\n", err)
			t.FailNow()
		}
		assert.NoError(t, err)
		assert.NotNil(t, blocks)
		assert.NotNil(t, bl)
		assert.NotNil(t, rq)

		assert.Len(t, rq.configured, 1)
		assert.Len(t, rq.steps, 5)

		rq.WalkAndExec(context.Background(), func(nodes []Node) error {
			node := nodes[0].(*TaskNode)
			if node.step == 1 {
				assert.Equal(t, "f1", node.name)
				assert.Len(t, node.args(), 2)
				assert.Equal(t, "v1", node.args()["k"])
			}
			if node.step == 2 {
				assert.Equal(t, "function2", node.name)
				assert.Len(t, node.args(), 1)
				assert.Equal(t, "v2", node.args()["k"])
			}
			if node.step == 3 {
				assert.Equal(t, "function3", node.name)
				assert.Len(t, node.args(), 0)
			}
			if node.step == 4 {
				assert.Equal(t, "function4", node.name)
				assert.NotNil(t, node.parallel)
				assert.Equal(t, "function5", node.parallel.name)
			}
			if node.step == 5 {
				assert.Equal(t, "function3", node.name)
				assert.Len(t, node.args(), 1)
				assert.Equal(t, "v3", node.args()["k"])
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

func TestBuiltiDirective(t *testing.T) {
	{
		const testingdata string = `
		var v = 1
		if $(v) == 1 {
			sleep "1s"
		}
	`

		_, _, rq, err := loadTestingdata2(testingdata)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.NoError(t, err)
		err = rq.WalkAndExec(context.Background(), nil)
		assert.NoError(t, err)
	}
	{
		const testingdata string = `
		var v = 1
		switch {
			case $(v) == 1 {
				println "error"
			}
			default {
				println "default"
			}
		}

		if $(v) == 1 {
			sleep "1s"
		}
	`

		_, _, rq, err := loadTestingdata2(testingdata)
		assert.NoError(t, err)
		err = rq.WalkAndExec(context.Background(), nil)
		assert.NoError(t, err)
	}
	{
		const testingdata string = `
		exit "hello"
	`

		_, _, rq, err := loadTestingdata2(testingdata)
		if err != nil {
			assert.FailNow(t, err.Error())
		}
		assert.NoError(t, err)
		err = rq.WalkAndExec(context.Background(), nil)
		assert.Error(t, err)
	}
}
