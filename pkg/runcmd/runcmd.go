package runcmd

import (
	"bytes"
	"context"
	"io"
	"os/exec"
	"strconv"
	"sync"
	"time"

	"github.com/cofxlabs/cofx/pkg/output"
	"github.com/cofxlabs/cofx/pkg/textparse"
)

type Wrap struct {
	// Name is the name of the command.
	Name string
	// Args is the arguments for the command.
	Args []string
	// Env is the environment variables for the command.
	Env []string
	// Dir is the working directory of the command.
	Dir string
	// SplitSep is the separator to split the output line.
	Split string
	// Extract is the fields to be extracted from output line.
	Extract []int
	// QueryColumns is the columns to be queried.
	QueryColumns []string
	// QueryWhere is the where condition for query.
	QueryWhere string
}

// Run runs the command and parses the output content.
func (w *Wrap) Run(ctx context.Context) (textparse.Rows, error) {
	nst, err := textparse.New(w.Name+strconv.Itoa(time.Now().Nanosecond()), w.Split, w.Extract)
	if err != nil {
		return nil, err
	}

	var buff bytes.Buffer
	out := &output.Output{
		W: &buff,
		HandleFunc: func(line []byte) {
			_ = nst.ParseLine(ctx, string(line))
		},
	}

	// start the command
	cmd := exec.CommandContext(ctx, w.Name, w.Args...)
	cmd.Env = append(cmd.Env, w.Env...)
	if w.Dir != "" {
		cmd.Dir = w.Dir
	}
	opipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	epipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	// copy stdout and stderr to out object
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer out.Close()
		io.Copy(out, opipe)
		io.Copy(out, epipe)
	}()
	wg.Wait()
	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	defer nst.Clear(ctx)
	// query output in sqlite db
	return nst.Query(ctx, w.QueryColumns, w.QueryWhere)
}
