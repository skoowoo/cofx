package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	co "github.com/cofunclabs/cofunc"
	"github.com/cofunclabs/cofunc/generator"
	"github.com/cofunclabs/cofunc/parser"
)

func parseflowl(name string, all bool) error {
	if !co.IsFlowl(name) {
		return errors.New("file is not a flowl: " + name)
	}

	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
	}()
	rq, ast, err := generator.New(f)
	if err != nil {
		return err
	}
	if all {
		printAST(ast, name)
	}
	printRunQ(rq, name)
	return nil
}

func printAST(ast *parser.AST, name string) {
	fmt.Printf("blocks in %s:\n", name)
	ast.Foreach(func(b *parser.Block) error {
		fmt.Printf("  %s\n", b.String())
		return nil
	})
}

func printRunQ(rq *generator.RunQueue, name string) {
	fmt.Printf("run queue in %s:\n", name)
	i := 0
	rq.ForstageAndExec(context.Background(), func(stage int, nodes []generator.Node) error {
		var buf bytes.Buffer
		i += 1
		buf.WriteString("Stage ")
		buf.WriteString(strconv.Itoa(i))
		buf.WriteString(": ")
		for _, node := range nodes {
			buf.WriteString(node.FormatString())
			buf.WriteString(" ")
		}
		fmt.Printf("  %s\n", buf.String())
		return nil
	})
}
