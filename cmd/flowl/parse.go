package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"

	co "github.com/cofunclabs/cofunc"
)

func parseFlowl(name string, all bool) error {
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
	rq, ast, err := co.ParseFlowl(f)
	if err != nil {
		return err
	}
	if all {
		printAST(ast, name)
	}
	printRunQ(rq, name)
	return nil
}

func printAST(ast *co.AST, name string) {
	fmt.Printf("blocks in %s:\n", name)
	ast.Foreach(func(b *co.Block) error {
		fmt.Printf("  %s\n", b.String())
		return nil
	})
}

func printRunQ(rq *co.RunQ, name string) {
	fmt.Printf("run queue in %s:\n", name)
	i := 0
	rq.Forstage(func(stage int, nodes []*co.FuncNode) error {
		var buf bytes.Buffer
		i += 1
		buf.WriteString("Stage ")
		buf.WriteString(strconv.Itoa(i))
		buf.WriteString(": ")
		for _, node := range nodes {
			buf.WriteString(node.String())
			buf.WriteString(" ")
		}
		fmt.Printf("  %s\n", buf.String())
		return nil
	})
}
