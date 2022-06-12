package main

import (
	"bytes"
	"fmt"
	"os"
	"strconv"

	"github.com/cofunclabs/cofunc/internal/flowl"
)

func parseFlowLAndPrint(name string, all bool) error {
	if err := flowl.ValidateFileName(name); err != nil {
		return err
	}
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer func() {
		f.Close()
	}()
	rq, bs, err := flowl.Parse(f)
	if err != nil {
		return err
	}
	if all {
		printBlocks(bs, name)
	}
	printRunQueue(rq, name)
	return nil
}

func printBlocks(bs *flowl.BlockStore, name string) {
	fmt.Printf("blocks in %s:\n", name)
	bs.Foreach(func(b *flowl.Block) error {
		fmt.Printf("  %s\n", b.String())
		return nil
	})
}

func printRunQueue(rq *flowl.RunQueue, name string) {
	fmt.Printf("run queue in %s:\n", name)
	i := 0
	rq.Step(func(batch *flowl.FunctionNode) {
		var buf bytes.Buffer
		i += 1
		buf.WriteString("Step ")
		buf.WriteString(strconv.Itoa(i))
		buf.WriteString(": ")
		for p := batch; p != nil; p = p.Parallel {
			buf.WriteString(p.Driver.Name())
			buf.WriteString(":")
			buf.WriteString(p.Name)
			buf.WriteString(" ")
		}

		fmt.Printf("  %s\n", buf.String())
	})
}
